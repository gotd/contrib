package floodwait

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/neo"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// runWaiter runs waiter and returns invoker wrapping next.
func runWaiter(t *testing.T, w *Waiter, next tg.Invoker) *tg.Client {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	started := make(chan struct{})
	go func() {
		done <- w.Run(ctx, func(ctx context.Context) error {
			close(started)
			<-ctx.Done()
			return nil
		})
	}()
	<-started
	t.Cleanup(func() {
		cancel()
		require.NoError(t, <-done)
	})
	return tg.NewClient(w.Handle(next))
}

type invokerFunc func(ctx context.Context, input bin.Encoder, output bin.Decoder) error

func (f invokerFunc) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return f(ctx, input, output)
}

func TestWaiterReentrant(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	w := NewWaiter()
	var api *tg.Client
	next := invokerFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		switch input.(type) {
		case *tg.UploadGetFileRequest:
			// Nested request through the same middleware chain, like auth
			// export on FILE_MIGRATE in gotd/td.
			if _, err := api.AuthExportAuthorization(ctx, 2); err != nil {
				return err
			}
			*output.(*tg.UploadFileBox) = tg.UploadFileBox{File: &tg.UploadFile{Type: &tg.StorageFileUnknown{}}}
			return nil
		case *tg.AuthExportAuthorizationRequest:
			*output.(*tg.AuthExportedAuthorization) = tg.AuthExportedAuthorization{ID: 1}
			return nil
		}
		return errors.Errorf("unexpected %T", input)
	})
	api = runWaiter(t, w, next)

	_, err := api.UploadGetFile(ctx, &tg.UploadGetFileRequest{Location: &tg.InputDocumentFileLocation{}})
	require.NoError(t, err, "nested invoke must not deadlock")
}

func TestWaiterKeyUnwrap(t *testing.T) {
	ctx := context.Background()
	history := &tg.MessagesGetHistoryRequest{Peer: &tg.InputPeerSelf{}}
	takeout := &tg.InvokeWithTakeoutRequest{TakeoutID: 1, Query: history}

	var k key
	k.fromEncoder(takeout)
	require.Equal(t, key(tg.MessagesGetHistoryRequestTypeID), k)
	require.Equal(t, "messages.getHistory", typeName(takeout))

	nested := &tg.InvokeWithoutUpdatesRequest{Query: takeout}
	k = 0
	k.fromEncoder(nested)
	require.Equal(t, key(tg.MessagesGetHistoryRequestTypeID), k)

	// Flood of wrapped request is accounted to the inner method.
	var events []FloodWait
	var calls atomic.Int64
	w := NewWaiter().WithCallback(func(ctx context.Context, wait FloodWait) {
		events = append(events, wait)
	})
	next := invokerFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		if calls.Add(1) == 1 {
			return tgerr.New(420, "FLOOD_WAIT_1")
		}
		return nil
	})
	inv := w.Handle(next)
	_ = runWaiter(t, w, next)
	done := make(chan error, 1)
	go func() { done <- inv.Invoke(ctx, takeout, &tg.MessagesMessagesBox{}) }()

	// Wait is active while the request is delayed.
	require.Eventually(t, func() bool {
		snap := w.Snapshot()
		return len(snap) == 1 &&
			snap[0].TypeID == uint32(tg.MessagesGetHistoryRequestTypeID) &&
			snap[0].Name == "messages.getHistory"
	}, 5*time.Second, time.Millisecond)
	require.NoError(t, <-done)
	require.Len(t, events, 1)
	require.Equal(t, uint32(tg.MessagesGetHistoryRequestTypeID), events[0].TypeID)
	require.Equal(t, "messages.getHistory", events[0].Name)
	require.True(t, events[0].PerType)
	require.True(t, tgerr.Is(events[0].Err, "FLOOD_WAIT"))
	require.Empty(t, w.Snapshot(), "wait decays on success")
}

func TestWaiterExempt(t *testing.T) {
	var calls int
	w := NewWaiter().WithExempt(tg.UploadGetFileRequestTypeID)
	next := invokerFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		calls++
		return tgerr.New(420, "FLOOD_WAIT_100")
	})
	// Not running: exempt requests bypass the Waiter.
	err := w.Handle(next).Invoke(context.Background(), &tg.UploadGetFileRequest{}, &tg.UploadFileBox{})
	require.True(t, tgerr.Is(err, "FLOOD_WAIT"))
	require.Equal(t, 1, calls)
}

func TestWaiterPolicy(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	newNext := func(fails int, err *tgerr.Error) (tg.Invoker, *atomic.Int64) {
		var calls atomic.Int64
		return invokerFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
			if int(calls.Add(1)) <= fails {
				return err
			}
			return nil
		}), &calls
	}

	t.Run("NoRetry", func(t *testing.T) {
		w := NewWaiter()
		next, calls := newNext(1, tgerr.New(420, "FLOOD_WAIT_1"))
		inv := runWaiter(t, w, next).Invoker()
		start := time.Now()
		err := inv.Invoke(WithPolicy(ctx, Policy{NoRetry: true}), &tg.HelpGetConfigRequest{}, &tg.Config{})
		require.True(t, tgerr.Is(err, "FLOOD_WAIT"))
		require.Less(t, time.Since(start), time.Second)
		require.EqualValues(t, 1, calls.Load())
	})
	t.Run("MaxWait", func(t *testing.T) {
		w := NewWaiter().WithMaxWait(500 * time.Millisecond)
		next, _ := newNext(1, tgerr.New(420, "FLOOD_WAIT_1"))
		inv := runWaiter(t, w, next).Invoker()
		err := inv.Invoke(ctx, &tg.HelpGetConfigRequest{}, &tg.Config{})
		require.ErrorContains(t, err, "too big")

		next, calls := newNext(1, tgerr.New(420, "FLOOD_WAIT_1"))
		w = NewWaiter().WithMaxWait(500 * time.Millisecond)
		inv = runWaiter(t, w, next).Invoker()
		// Background: wait as long as it takes.
		require.NoError(t, inv.Invoke(WithPolicy(ctx, Policy{MaxWait: -1}), &tg.HelpGetNearestDCRequest{}, &tg.NearestDC{}))
		require.EqualValues(t, 2, calls.Load())
	})
	t.Run("ServerError", func(t *testing.T) {
		next, calls := newNext(1, tgerr.New(500, "INTERNAL"))
		inv := runWaiter(t, NewWaiter(), next).Invoker()
		err := inv.Invoke(ctx, &tg.HelpGetConfigRequest{}, &tg.Config{})
		require.True(t, tgerr.Is(err, "INTERNAL"), "disabled by default")

		next, calls = newNext(1, tgerr.New(500, "INTERNAL"))
		inv = runWaiter(t, NewWaiter().WithServerErrorRetries(1), next).Invoker()
		require.NoError(t, inv.Invoke(ctx, &tg.HelpGetConfigRequest{}, &tg.Config{}))
		require.EqualValues(t, 2, calls.Load())

		next, _ = newNext(10, tgerr.New(500, "INTERNAL"))
		inv = runWaiter(t, NewWaiter().WithServerErrorRetries(1), next).Invoker()
		require.True(t, tgerr.Is(inv.Invoke(ctx, &tg.HelpGetConfigRequest{}, &tg.Config{}), "INTERNAL"))
	})
}

func TestSchedulerAdaptivePacing(t *testing.T) {
	a := require.New(t)
	n := neo.NewTime(time.Unix(0, 0))
	sch := newScheduler(n, time.Second)
	sch.minInterval = 100 * time.Millisecond
	sch.maxInterval = 400 * time.Millisecond

	r := request{key: 1}
	sch.flood(r, time.Second)
	a.Equal(100*time.Millisecond, sch.interval[1])
	a.Len(sch.gather(nil), 0)
	// Flood also moves the request itself by the wait.
	n.Travel(2*time.Second + time.Millisecond)
	a.Len(sch.gather(nil), 1)

	for range 3 {
		sch.flood(r, time.Second)
		n.Travel(2*time.Second + time.Millisecond)
		a.Len(sch.gather(nil), 1)
	}
	a.Equal(400*time.Millisecond, sch.interval[1], "doubles up to max")

	// Requests of the type are spaced by interval.
	sch.nice(1)
	iv := sch.interval[1]
	a.Less(iv, 400*time.Millisecond, "decays on success")
	sch.mux.Lock()
	sch.state = map[key]time.Duration{}
	sch.schedule(request{key: 1})
	sch.schedule(request{key: 1})
	sch.mux.Unlock()
	a.Len(sch.gather(nil), 1)
	n.Travel(iv)
	a.Len(sch.gather(nil), 1)

	for range 100 {
		sch.nice(1)
	}
	_, ok := sch.interval[1]
	a.False(ok, "removed below min")
}

var _ telegram.Middleware = (*Waiter)(nil)

func TestWaiterConcurrentDispatch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	release := make(chan struct{})
	var once sync.Once
	next := invokerFunc(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		if _, ok := input.(*tg.HelpGetConfigRequest); ok {
			<-release
		}
		return nil
	})
	inv := runWaiter(t, NewWaiter(), next).Invoker()
	slow := make(chan error, 1)
	go func() { slow <- inv.Invoke(ctx, &tg.HelpGetConfigRequest{}, &tg.Config{}) }()

	// Slow request does not block others.
	require.NoError(t, inv.Invoke(ctx, &tg.HelpGetNearestDCRequest{}, &tg.NearestDC{}))
	once.Do(func() { close(release) })
	require.NoError(t, <-slow)
}
