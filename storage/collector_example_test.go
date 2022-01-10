package storage_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/pebble"
	"github.com/gotd/contrib/storage"
)

func peerCollector(ctx context.Context) error {
	db, err := pebbledb.Open("pebble.db", &pebbledb.Options{})
	if err != nil {
		return errors.Errorf("create pebble storage: %w", err)
	}
	s := pebble.NewPeerStorage(db)
	collector := storage.CollectPeers(s)

	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return errors.Errorf("create client: %w", err)
	}
	raw := tg.NewClient(client)

	return client.Run(ctx, func(ctx context.Context) error {
		// Fills storage with user dialogs peers metadata.
		return collector.Dialogs(ctx, query.GetDialogs(raw).Iter())
	})
}

func ExampleCollectPeers() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := peerCollector(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
