package tests

import (
	"context"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// RetryUntilAvailable calls callback repeatedly until callback return nil error
// or until test timeout will be reached.
func RetryUntilAvailable(t *testing.T, serviceName, addr string, f func(ctx context.Context) error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = time.Second * 5

	if err := backoff.Retry(func() error {
		t.Logf("Trying to connect to %s %s", serviceName, addr)
		return f(ctx)
	}, backoff.WithContext(bo, ctx)); err != nil {
		t.Fatalf("Could not connect to %s: %s", serviceName, err)
		return
	}
}
