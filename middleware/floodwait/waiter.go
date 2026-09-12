package floodwait

import (
	"context"
	"time"

	"github.com/go-faster/errors"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

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
}

// FloodWait event.
type FloodWait struct {
	Duration time.Duration
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
		clock:      w.clock,
		sch:        w.sch,
		tick:       w.tick,
		maxWait:    w.maxWait,
		maxRetries: w.maxRetries,
		onWait:     w.onWait,
	}
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
	defer w.running.Store(false)

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

				for _, s := range requests {
					ret, err := w.send(s)
					if ret {
						select {
						case s.request.result <- err:
						default:
						}
					}
				}
			case <-ctx.Done():
				return nil
			}
		}
	})

	return wg.Wait()
}

func (w *Waiter) send(s scheduled) (bool, error) {
	// Mark the context so that any invoke this call makes while the send loop
	// is blocked on it is recognised as re-entrant and handled inline.
	ctx := context.WithValue(s.request.ctx, reentrantKey{}, struct{}{})
	err := s.request.next.Invoke(ctx, s.request.input, s.request.output)

	// Detect flood wait and similar retriable wait errors, mirroring TDLib.
	d, perType, ok := asWait(err)
	if !ok {
		w.sch.nice(s.request.key)
		return true, err
	}

	// Notify about flood wait.
	w.onWait(s.request.ctx, FloodWait{
		Duration: d,
	})

	s.request.retry++

	if v := w.maxRetries; v != 0 && s.request.retry > v {
		return true, errors.Wrapf(err, "flood wait retry limit exceeded (%d > %d)", s.request.retry, v)
	}

	if v := w.maxWait; v != 0 && d > v {
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

// reentrantKey marks a context as already executing inside the scheduler's send
// loop, so a nested invoke can be recognised and handled inline.
type reentrantKey struct{}

// Handle implements telegram.Middleware.
func (w *Waiter) Handle(next tg.Invoker) telegram.InvokeFunc {
	return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		if !w.running.Load() {
			// Return explicit error if waiter is not running.
			return errors.New("the Waiter middleware is not running: Run(ctx) method is not called or exited")
		}
		if ctx.Value(reentrantKey{}) != nil {
			// This invoke was started by another invoke that the send loop is
			// still blocked on. Queueing it would deadlock: the send loop
			// cannot reach it until the outer call returns, and the outer call
			// cannot return until this one does. Wait inline instead.
			return w.invokeInline(ctx, input, output, next)
		}
		select {
		case err := <-w.sch.new(ctx, input, output, next):
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// invokeInline runs a nested invoke without the scheduler, retrying flood waits
// on the calling goroutine. It honours the same maxWait and maxRetries bounds as
// the scheduled path, but does not update per-type scheduler state: a nested
// call is part of an operation already in flight, not new traffic to pace.
func (w *Waiter) invokeInline(
	ctx context.Context, input bin.Encoder, output bin.Decoder, next tg.Invoker,
) error {
	var (
		timer   clock.Timer
		retries int
	)
	for {
		err := next.Invoke(ctx, input, output)

		d, _, ok := asWait(err)
		if !ok {
			return err
		}

		w.onWait(ctx, FloodWait{Duration: d})

		retries++
		if v := w.maxRetries; v != 0 && retries > v {
			return errors.Wrapf(err, "flood wait retry limit exceeded (%d > %d)", retries, v)
		}
		if v := w.maxWait; v != 0 && d > v {
			return errors.Wrapf(err, "flood wait argument is too big (%v > %v)", d, v)
		}

		if timer == nil {
			timer = w.clock.Timer(d)
			defer timer.Stop()
		} else {
			timer.Reset(d)
		}

		select {
		case <-timer.C():
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
