package bg

import (
	"context"
	"testing"

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
