package bg

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testKey string

type testClient struct {
	tt *testing.T
}

func (t testClient) Run(ctx context.Context, f func(ctx context.Context) error) error {
	assert.Equal(t.tt, "bar", ctx.Value(testKey("foo")))
	return f(ctx)
}

func TestConnect(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, testKey("foo"), "bar")
	stop, err := Connect(testClient{tt: t}, WithContext(ctx))
	require.NoError(t, err)
	require.NoError(t, stop())
}

// readyClient becomes ready immediately by invoking the callback, like a real
// client whose primary connection is established.
type readyClient struct{}

func (readyClient) Run(ctx context.Context, f func(ctx context.Context) error) error {
	return f(ctx)
}

func TestConnectStop(t *testing.T) {
	stop, err := Connect(readyClient{})
	require.NoError(t, err)
	// Stop must cancel the client and wait for Run to return.
	require.NoError(t, stop())
}

// errClient returns the given error from Run without ever calling the callback,
// emulating a startup failure.
type errClient struct {
	err error
}

func (c errClient) Run(context.Context, func(ctx context.Context) error) error {
	return c.err
}

func TestConnectRunError(t *testing.T) {
	want := errors.New("boom")
	stop, err := Connect(errClient{err: want})
	require.ErrorIs(t, err, want)
	// StopFunc must be safe to call even on error.
	require.NoError(t, stop())
}

// blockingClient never becomes ready: it blocks until the context is canceled
// without ever calling the callback, emulating an endless reconnect loop.
type blockingClient struct{}

func (blockingClient) Run(ctx context.Context, _ func(ctx context.Context) error) error {
	<-ctx.Done()
	return ctx.Err()
}

func TestConnectStartupTimeout(t *testing.T) {
	stop, err := Connect(blockingClient{}, WithStartupTimeout(10*time.Millisecond))
	require.ErrorIs(t, err, ErrStartupTimeout)
	require.NoError(t, stop())
}

func TestConnectContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	stop, err := Connect(blockingClient{}, WithContext(ctx))
	require.ErrorIs(t, err, context.Canceled)
	require.NoError(t, stop())
}
