package redis_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	redisclient "github.com/go-redis/redis/v8"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/redis"
)

func redisCache(ctx context.Context) error {
	redisClient := redisclient.NewClient(&redisclient.Options{})

	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return xerrors.Errorf("create client: %w", err)
	}

	return client.Run(ctx, func(ctx context.Context) error {
		raw := tg.NewClient(client)
		resolver := redis.NewResolverCache(peer.DefaultResolver(raw), redisClient)
		s := message.NewSender(raw).WithResolver(resolver)

		_, err := s.Resolve("durov").Text(ctx, "Hi!")
		return err
	})
}

func ExampleResolverCache() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := redisCache(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
