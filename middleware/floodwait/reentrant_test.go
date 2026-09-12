package floodwait

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// nestedInvoker performs one nested invoke from inside its own Invoke,
// reproducing the shape of a datacenter migration in gotd/td: upload.getFile
// fails with FILE_MIGRATE_X, and the client handles it by calling
// auth.exportAuthorization through the same middleware chain while the original
// call is still in flight (telegram/invoke.go invokeDirect -> invokeSub ->
// telegram/pool.go dc -> telegram/transfer.go transfer).
type nestedInvoker struct {
	nested tg.Invoker
	calls  int
}

func (n *nestedInvoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	n.calls++
	if n.calls > 1 {
		return nil
	}
	return n.nested.Invoke(ctx, input, output)
}

// TestWaiterNestedInvoke covers an invoke started by another invoke that the
// send loop is still blocked on.
//
// Run's send loop calls Invoke synchronously, so it cannot dispatch anything
// else until that call returns. A nested invoke queued through the scheduler
// would therefore never be sent, while the outer call waits for its result:
// both block forever. Before this was handled, every cross-datacenter download
// in gotd/cli hung indefinitely, produced a zero-byte file and reported no
// error.
func TestWaiterNestedInvoke(t *testing.T) {
	t.Parallel()

	w := NewWaiter().WithTick(time.Millisecond)

	inner := w.Handle(invokerFunc(func(context.Context, bin.Encoder, bin.Decoder) error {
		return nil
	}))
	outer := w.Handle(&nestedInvoker{nested: inner})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, w.Run(ctx, func(ctx context.Context) error {
		return outer.Invoke(ctx, nil, nil)
	}))
}

// TestWaiterNestedInvokeRetriesFloodWait checks that a nested invoke still
// honours flood waits rather than simply bypassing them.
func TestWaiterNestedInvokeRetriesFloodWait(t *testing.T) {
	t.Parallel()

	var attempts int
	w := NewWaiter().WithTick(time.Millisecond).WithMaxWait(time.Minute)

	inner := w.Handle(invokerFunc(func(context.Context, bin.Encoder, bin.Decoder) error {
		attempts++
		if attempts == 1 {
			return tgerr.New(420, "FLOOD_WAIT_1")
		}
		return nil
	}))
	outer := w.Handle(&nestedInvoker{nested: inner})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	require.NoError(t, w.Run(ctx, func(ctx context.Context) error {
		return outer.Invoke(ctx, nil, nil)
	}))
	require.Equal(t, 2, attempts, "nested invoke should retry after the flood wait")
}

// TestWaiterNestedInvokeRespectsMaxWait checks that the inline path applies the
// same bounds as the scheduled one.
func TestWaiterNestedInvokeRespectsMaxWait(t *testing.T) {
	t.Parallel()

	w := NewWaiter().WithTick(time.Millisecond).WithMaxWait(time.Second)

	inner := w.Handle(invokerFunc(func(context.Context, bin.Encoder, bin.Decoder) error {
		return tgerr.New(420, "FLOOD_WAIT_3600")
	}))
	outer := w.Handle(&nestedInvoker{nested: inner})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := w.Run(ctx, func(ctx context.Context) error {
		return outer.Invoke(ctx, nil, nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "too big")
}

type invokerFunc func(ctx context.Context, input bin.Encoder, output bin.Decoder) error

func (f invokerFunc) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return f(ctx, input, output)
}

var _ telegram.Middleware = (*Waiter)(nil)
