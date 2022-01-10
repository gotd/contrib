package storage_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"

	"github.com/gotd/td/telegram"

	"github.com/gotd/contrib/pebble"
	"github.com/gotd/contrib/storage"
)

func resolverCache(ctx context.Context) error {
	db, err := pebbledb.Open("pebble.db", &pebbledb.Options{})
	if err != nil {
		return errors.Errorf("create pebble storage: %w", err)
	}

	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return errors.Errorf("create client: %w", err)
	}

	return client.Run(ctx, func(ctx context.Context) error {
		raw := tg.NewClient(client)
		resolver := storage.NewResolverCache(peer.Plain(raw), pebble.NewPeerStorage(db))
		s := message.NewSender(raw).WithResolver(resolver)

		_, err := s.Resolve("durov").Text(ctx, "Hi!")
		return err
	})
}

func ExampleResolverCache() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := resolverCache(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
