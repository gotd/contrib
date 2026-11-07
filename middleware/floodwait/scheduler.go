package floodwait

import (
	"context"
	"sync"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/tg"
)

type scheduler struct {
	state map[key]time.Duration
	mux   sync.Mutex
	queue *queue

	clock clock.Clock
	dec   time.Duration

	// Adaptive pacing, enabled if maxInterval > 0.
	minInterval time.Duration
	maxInterval time.Duration
	// interval is the minimum spacing of requests per key.
	interval map[key]time.Duration
	// last is the planned send time of the last request per key.
	last map[key]time.Time
	// flooded is the time of the last per-type flood wait per key.
	flooded map[key]time.Time
	// names are TL names of keys, for snapshots.
	names map[key]string
}

func newScheduler(c clock.Clock, dec time.Duration) *scheduler {
	const initialCapacity = 16

	return &scheduler{
		state:    make(map[key]time.Duration, initialCapacity),
		queue:    newQueue(initialCapacity),
		clock:    c,
		dec:      dec,
		interval: map[key]time.Duration{},
		last:     map[key]time.Time{},
		flooded:  map[key]time.Time{},
		names:    map[key]string{},
	}
}

// withClock returns a new scheduler with the same settings and given clock.
//
// Must be called before the scheduler is used: the state is not copied.
func (s *scheduler) withClock(c clock.Clock) *scheduler {
	n := newScheduler(c, s.dec)
	n.minInterval = s.minInterval
	n.maxInterval = s.maxInterval
	return n
}

// KeyState is a pacing state of a request type, see Waiter.Snapshot.
type KeyState struct {
	// TypeID is the TL type ID of the request (unwrapped from invokeWith*).
	TypeID uint32
	// Name is the TL name of the request, if known.
	Name string
	// Wait is the delay added to new requests of the type after flood wait.
	Wait time.Duration
	// Interval is the minimum spacing of requests of the type (adaptive
	// pacing).
	Interval time.Duration
}

func (s *scheduler) snapshot() []KeyState {
	s.mux.Lock()
	defer s.mux.Unlock()
	keys := map[key]struct{}{}
	for k := range s.state {
		keys[k] = struct{}{}
	}
	for k := range s.interval {
		keys[k] = struct{}{}
	}
	r := make([]KeyState, 0, len(keys))
	for k := range keys {
		r = append(r, KeyState{
			TypeID:   uint32(k),
			Name:     s.names[k],
			Wait:     s.state[k],
			Interval: s.interval[k],
		})
	}
	return r
}

func (s *scheduler) new(ctx context.Context, input bin.Encoder, output bin.Decoder, next tg.Invoker) <-chan error {
	var k key
	k.fromEncoder(input)
	r := request{
		ctx:    ctx,
		input:  input,
		output: output,
		next:   next,
		key:    k,
		result: make(chan error, 1),
	}

	r.policy = policyFrom(ctx)

	s.mux.Lock()
	defer s.mux.Unlock()
	if _, ok := s.names[k]; !ok {
		s.names[k] = typeName(input)
	}
	s.schedule(r)
	return r.result
}

// schedule adds request to the queue.
// Assumes the mutex is locked.
func (s *scheduler) schedule(r request) {
	k := r.key

	var t time.Time
	if state, ok := s.state[k]; ok {
		t = s.clock.Now().Add(state)
	} else {
		t = s.clock.Now()
	}
	if iv, ok := s.interval[k]; ok && iv > 0 {
		if next := s.last[k].Add(iv); next.After(t) {
			t = next
		}
		s.last[k] = t
	}
	s.queue.add(r, t)
}

// drain removes all queued requests, failing them with err.
func (s *scheduler) drain(err error) {
	// Far enough in the future to include any wait.
	for _, r := range s.queue.gather(s.clock.Now().Add(100*365*24*time.Hour), nil) {
		select {
		case r.request.result <- err:
		default:
		}
	}
}

func (s *scheduler) gather(r []scheduled) []scheduled {
	return s.queue.gather(s.clock.Now(), r)
}

func (s *scheduler) nice(k key) {
	s.mux.Lock()
	if state, ok := s.state[k]; ok && state-s.dec > 0 {
		s.state[k] = state - s.dec
	} else {
		delete(s.state, k)
	}
	// Decay interval on success down to minInterval, and drop it only after
	// paceQuiet without flood waits: a single success (usually the retried
	// request itself) must not remove the spacing that avoids the next wait.
	if iv, ok := s.interval[k]; ok {
		iv -= iv / 8
		if iv < s.minInterval {
			iv = s.minInterval
		}
		if s.clock.Now().Sub(s.flooded[k]) >= paceQuiet {
			delete(s.interval, k)
			delete(s.last, k)
			delete(s.flooded, k)
		} else {
			s.interval[k] = iv
		}
	}
	s.mux.Unlock()
}

// paceQuiet is how long a request type must go without flood waits before
// its adaptive pacing interval is dropped.
const paceQuiet = 10 * time.Minute

// penalize records per-type flood wait d of key k.
// Assumes the mutex is locked.
func (s *scheduler) penalize(k key, now time.Time, d time.Duration) {
	if state, ok := s.state[k]; !ok || state < d {
		s.state[k] = d
	}
	if s.maxInterval > 0 {
		// Grow interval on each flood wait of the type.
		iv := s.interval[k] * 2
		if iv < s.minInterval {
			iv = s.minInterval
		}
		if iv > s.maxInterval {
			iv = s.maxInterval
		}
		s.interval[k] = iv
		s.flooded[k] = now
	}
}

// record records per-type flood wait without re-scheduling the request, for
// requests that are not retried.
func (s *scheduler) record(k key, d time.Duration) {
	s.mux.Lock()
	s.penalize(k, s.clock.Now(), d)
	s.mux.Unlock()
}

func (s *scheduler) flood(req request, d time.Duration) {
	k := req.key

	s.mux.Lock()
	defer s.mux.Unlock()
	now := s.clock.Now()
	s.penalize(k, now, d)
	// Move queued requests of the type before adding this one, so it is not
	// delayed twice.
	s.queue.move(k, now, d)
	s.queue.add(req, now.Add(d))
}

// retry re-schedules a single request after d without updating the per-type
// state. It is used for chat- or operation-specific waits (e.g. SLOWMODE_WAIT)
// that must not throttle unrelated requests of the same type.
func (s *scheduler) retry(req request, d time.Duration) {
	s.queue.add(req, s.clock.Now().Add(d))
}
