package bbolt_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/go-faster/errors"
	bboltdb "go.etcd.io/bbolt"

	"github.com/gotd/td/telegram"

	"github.com/gotd/contrib/bbolt"
)

func bboltStorage(ctx context.Context) error {
	db, err := bboltdb.Open("bbolt.db", 0666, &bboltdb.Options{}) // nolint:gocritic
	if err != nil {
		return errors.Errorf("create bbolt storage: %w", err)
	}
	storage := bbolt.NewSessionStorage(db, "session", []byte("bucket"))

	client, err := telegram.ClientFromEnvironment(telegram.Options{
		SessionStorage: storage,
	})
	if err != nil {
		return errors.Errorf("create client: %w", err)
	}

	return client.Run(ctx, func(ctx context.Context) error {
		_, err := client.Auth().Bot(ctx, os.Getenv("BOT_TOKEN"))
		return err
	})
}

func ExampleSessionStorage() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := bboltStorage(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
