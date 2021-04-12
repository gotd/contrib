package redis_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	redisclient "github.com/go-redis/redis/v8"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"

	"github.com/gotd/contrib/auth/redis"
)

func redisStorage(ctx context.Context) error {
	redisClient := redisclient.NewClient(&redisclient.Options{})
	storage := redis.NewSessionStorage(redisClient, "session")

	client, err := telegram.ClientFromEnvironment(telegram.Options{
		SessionStorage: storage,
	})
	if err != nil {
		return xerrors.Errorf("create client: %w", err)
	}

	return client.Run(ctx, func(ctx context.Context) error {
		_, err := client.AuthBot(ctx, os.Getenv("BOT_TOKEN"))
		return err
	})
}

func ExampleSessionStorage() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := redisStorage(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
