package redis_test

import (
	"context"
	"os"
	"testing"

	redisclient "github.com/go-redis/redis/v8"

	"github.com/gotd/td/telegram/message/peer"

	"github.com/gotd/contrib/internal/tests"
	"github.com/gotd/contrib/redis"
)

func TestE2E(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("Set REDIS_ADDR to run E2E test")
	}

	client := redisclient.NewClient(&redisclient.Options{
		Addr: addr,
	})
	tests.RetryUntilAvailable(t, "Redis", addr, func(ctx context.Context) error {
		return client.Ping(ctx).Err()
	})

	tests.TestSessionStorage(t, redis.NewSessionStorage(client, "session"))
	tests.TestCredentials(t, redis.NewCredentials(client))
	tests.TestResolverCache(t, func(next peer.Resolver) peer.Resolver {
		return redis.NewResolverCache(next, client)
	})
}
