package floodwait

import (
	"context"
	"sync"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
)

const (
	defaultTick       = time.Millisecond
	defaultMaxWait    = time.Minute
	defaultMaxRetries = 5
)

// Waiter is a tg.Invoker that handles flood wait errors on underlying invoker.
//
// This implementation uses a request scheduler and is more suitable for long-running
// programs with high level of concurrency and parallelism.
//
// You should use Waiter if unsure which waiter implementation to use.
//
// See SimpleWaiter for a simple timer-based implementation.
type Waiter struct {
	clock clock.Clock
	sch   *scheduler

	running    atomic.Bool
	tick       time.Duration
	maxWait    time.Duration
	maxRetries int
	onWait     func(ctx context.Context, wait FloodWait)
	// exempt are type IDs passed to next invoker directly.
	exempt map[uint32]struct{}
	// serverRetries is the limit of retries of 5xx errors, 0 disables.
	serverRetries int
}

// FloodWait event.
type FloodWait struct {
	Duration time.Duration
	// TypeID is the TL type ID of the request, unwrapped from invokeWith*
	// (e.g. the method called in a takeout session).
	TypeID uint32
	// Name is the TL name of the request, if known.
	Name string
	// PerType is true if the wait applies to all requests of the type
	// (FLOOD_WAIT), false if only to this request (e.g. SLOWMODE_WAIT,
	// server errors).
	PerType bool
	// Err is the error that caused the wait.
	Err error
}

// NewWaiter returns a new invoker that waits on the flood wait errors.
//
// NB: You MUST use Run method. Example:
//
//	if err := waiter.Run(ctx, func(ctx context.Context) error {
//		// Client should be started after waiter.
//		return client.Run(ctx, handler)
//	}); err != nil {
//		return errors.Wrap(err, "run client")
//	}
func NewWaiter() *Waiter {
	return &Waiter{
		clock:      clock.System,
		sch:        newScheduler(clock.System, time.Second),
		tick:       defaultTick,
		maxWait:    defaultMaxWait,
		maxRetries: defaultMaxRetries,
		onWait:     func(ctx context.Context, wait FloodWait) {},
	}
}

// WithCallback sets callback for flood wait event.
func (w *Waiter) WithCallback(f func(ctx context.Context, wait FloodWait)) *Waiter {
	w = w.clone()
	w.onWait = f
	return w
}

// clone returns a copy of the Waiter.
func (w *Waiter) clone() *Waiter {
	return &Waiter{
		clock:         w.clock,
		sch:           w.sch,
		tick:          w.tick,
		maxWait:       w.maxWait,
		maxRetries:    w.maxRetries,
		onWait:        w.onWait,
		exempt:        w.exempt,
		serverRetries: w.serverRetries,
	}
}

// WithExempt makes requests of given types bypass the Waiter, e.g. requests
// with their own pacing like upload.getFile on file pools
// (tg.UploadGetFileRequestTypeID). Requests are matched after unwrapping
// invokeWith* wrappers.
func (w *Waiter) WithExempt(typeIDs ...uint32) *Waiter {
	w = w.clone()
	exempt := make(map[uint32]struct{}, len(w.exempt)+len(typeIDs))
	for id := range w.exempt {
		exempt[id] = struct{}{}
	}
	for _, id := range typeIDs {
		exempt[id] = struct{}{}
	}
	w.exempt = exempt
	return w
}

// WithAdaptivePacing enables adaptive pacing: each flood wait of a request
// type doubles the minimum spacing of requests of that type (starting from
// minInterval, up to maxInterval), and each success decays it down to
// minInterval, so long-running programs do not hit the same limit again right
// after the wait. The spacing is removed after 10 minutes without flood waits
// of the type.
//
// Must be called before Run. Disabled by default. The returned Waiter has its
// own scheduler: the original Waiter is not changed.
func (w *Waiter) WithAdaptivePacing(minInterval, maxInterval time.Duration) *Waiter {
	w = w.clone()
	if minInterval <= 0 {
		minInterval = 50 * time.Millisecond
	}
	if maxInterval < minInterval {
		maxInterval = minInterval
	}
	sch := newScheduler(w.sch.clock, w.sch.dec)
	sch.minInterval = minInterval
	sch.maxInterval = maxInterval
	w.sch = sch
	return w
}

// WithServerErrorRetries enables retries of transient server errors (code
// 500 and above, e.g. INTERNAL or Timeout) with exponential backoff from 1s
// up to 32s, at most n times. Disabled by default; WORKER_BUSY_TOO_LONG_RETRY
// is always retried as a wait error.
func (w *Waiter) WithServerErrorRetries(n int) *Waiter {
	w = w.clone()
	w.serverRetries = n
	return w
}

// Snapshot returns pacing state of request types with active waits or
// intervals, e.g. for metrics.
func (w *Waiter) Snapshot() []KeyState {
	return w.sch.snapshot()
}

// WithClock sets clock to use. Default is to use system clock.
func (w *Waiter) WithClock(c clock.Clock) *Waiter {
	w = w.clone()
	w.clock = c
	return w
}

// WithMaxWait limits wait time per attempt. Waiter will return an error if flood wait
// time exceeds that limit. Default is to wait at most a minute.
//
// To limit total wait time use a context.Context with timeout or deadline set.
func (w *Waiter) WithMaxWait(m time.Duration) *Waiter {
	w = w.clone()
	w.maxWait = m
	return w
}

// WithMaxRetries sets max number of retries before giving up. Default is to retry at most 5 times.
func (w *Waiter) WithMaxRetries(m int) *Waiter {
	w = w.clone()
	w.maxRetries = m
	return w
}

// WithTick sets gather tick interval for Waiter. Default is 1ms.
func (w *Waiter) WithTick(t time.Duration) *Waiter {
	w = w.clone()
	if t <= 0 {
		t = time.Nanosecond
	}
	w.tick = t
	return w
}

// Run runs send loop.
//
// Example:
//
//	if err := waiter.Run(ctx, func(ctx context.Context) error {
//		// Client should be started after waiter.
//		return client.Run(ctx, handler)
//	}); err != nil {
//		return errors.Wrap(err, "run client")
//	}
func (w *Waiter) Run(ctx context.Context, f func(ctx context.Context) error) (err error) {
	w.running.Store(true)
	// Wait for in-flight sends, then fail requests left in the queue, so
	// that callers do not wait for a send loop that is gone.
	var dispatch sync.WaitGroup
	defer func() {
		dispatch.Wait()
		w.running.Store(false)
		w.sch.drain(errors.New("the Waiter middleware exited"))
	}()

	ctx, cancel := context.WithCancel(ctx)
	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		defer cancel()
		return f(ctx)
	})
	wg.Go(func() error {
		ticker := w.clock.Ticker(w.tick)
		defer ticker.Stop()

		var requests []scheduled
		for {
			select {
			case <-ticker.C():
				requests = w.sch.gather(requests[:0])
				if len(requests) < 1 {
					continue
				}

				// Dispatch concurrently: a request may invoke nested requests
				// through the same middleware chain (e.g. auth export on
				// FILE_MIGRATE in gotd/td), which must not wait for the outer
				// one, and one slow request must not delay others.
				for _, s := range requests {
					dispatch.Add(1)
					go func(s scheduled) {
						defer dispatch.Done()
						ret, err := w.send(s)
						if ret {
							select {
							case s.request.result <- err:
							default:
							}
						}
					}(s)
				}
			case <-ctx.Done():
				return nil
			}
		}
	})

	return wg.Wait()
}

func (w *Waiter) send(s scheduled) (bool, error) {
	err := s.request.next.Invoke(s.request.ctx, s.request.input, s.request.output)

	// Detect flood wait and similar retriable wait errors, mirroring TDLib.
	d, perType, ok := asWait(err)
	if !ok {
		if d, ok := w.serverBackoff(s.request, err); ok {
			s.request.serverRetry++
			w.notify(s.request, d, false, err)
			w.sch.retry(s.request, d)
			return false, nil
		}
		if err == nil {
			w.sch.nice(s.request.key)
		}
		return true, err
	}

	// Notify about flood wait.
	w.notify(s.request, d, perType, err)

	if s.request.policy.NoRetry {
		if perType {
			// Not retried, but other requests of the type are still delayed.
			w.sch.record(s.request.key, d)
		}
		return true, err
	}

	s.request.retry++

	if v := limit(s.request.policy.MaxRetries, w.maxRetries); v != 0 && s.request.retry > v {
		return true, errors.Wrapf(err, "flood wait retry limit exceeded (%d > %d)", s.request.retry, v)
	}

	if v := limit(s.request.policy.MaxWait, w.maxWait); v != 0 && d > v {
		return true, errors.Wrapf(err, "flood wait argument is too big (%v > %v)", d, v)
	}

	if perType {
		// Per-method rate limit: proactively delay future requests of this type.
		w.sch.flood(s.request, d)
	} else {
		// Chat- or operation-specific wait: retry only this request.
		w.sch.retry(s.request, d)
	}
	return false, nil
}

func (w *Waiter) notify(r request, d time.Duration, perType bool, err error) {
	w.onWait(r.ctx, FloodWait{
		Duration: d,
		TypeID:   uint32(r.key),
		Name:     typeName(r.input),
		PerType:  perType,
		Err:      err,
	})
}

// maxServerBackoff limits server error backoff.
const maxServerBackoff = 32 * time.Second

// serverBackoff returns backoff for retrying transient server error.
func (w *Waiter) serverBackoff(r request, err error) (time.Duration, bool) {
	if w.serverRetries <= 0 || r.policy.NoRetry || err == nil {
		return 0, false
	}
	rpcErr, ok := tgerr.As(err)
	if !ok || rpcErr.Code < 500 {
		return 0, false
	}
	if r.serverRetry >= w.serverRetries {
		return 0, false
	}
	d := time.Second << r.serverRetry
	if d > maxServerBackoff || d <= 0 {
		d = maxServerBackoff
	}
	return d, true
}

func (w *Waiter) isExempt(input bin.Encoder) bool {
	if len(w.exempt) == 0 {
		return false
	}
	obj, ok := unwrap(input).(object)
	if !ok {
		return false
	}
	_, ok = w.exempt[obj.TypeID()]
	return ok
}

// Handle implements telegram.Middleware.
func (w *Waiter) Handle(next tg.Invoker) telegram.InvokeFunc {
	return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		if w.isExempt(input) {
			return next.Invoke(ctx, input, output)
		}
		if !w.running.Load() {
			// Return explicit error if waiter is not running.
			return errors.New("the Waiter middleware is not running: Run(ctx) method is not called or exited")
		}
		select {
		case err := <-w.sch.new(ctx, input, output, next):
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
