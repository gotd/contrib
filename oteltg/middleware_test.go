package oteltg

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

type invoker func(ctx context.Context, input bin.Encoder, output bin.Decoder) error

func (i invoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return i(ctx, input, output)
}

func TestMiddleware_Handle(t *testing.T) {
	m, err := New(noop.NewMeterProvider(), trace.NewNoopTracerProvider())
	require.NoError(t, err)

	okInvoker := invoker(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		return nil
	})
	errInvoker := invoker(func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		return tgerr.New(0, tgerr.ErrFloodWait)
	})

	ctx := context.Background()
	input := &tg.UsersGetUsersRequest{}
	require.NoError(t, m.Handle(okInvoker).Invoke(ctx, input, nil))
	require.NoError(t, m.Handle(okInvoker).Invoke(ctx, nil, nil))
	require.True(t, tgerr.Is(m.Handle(errInvoker).Invoke(ctx, input, nil), tgerr.ErrFloodWait))
}
