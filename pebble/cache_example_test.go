package pebble_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	pebbledb "github.com/cockroachdb/pebble"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/pebble"
)

func pebbleCache(ctx context.Context) error {
	db, err := pebbledb.Open("pebble.db", &pebbledb.Options{})
	if err != nil {
		return xerrors.Errorf("create pebble storage: %w", err)
	}

	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return xerrors.Errorf("create client: %w", err)
	}

	return client.Run(ctx, func(ctx context.Context) error {
		raw := tg.NewClient(client)
		resolver := pebble.NewResolverCache(peer.DefaultResolver(raw), db)
		s := message.NewSender(raw).WithResolver(resolver)

		_, err := s.Resolve("durov").Text(ctx, "please migrate to the gRPC")
		return err
	})
}

func ExampleResolverCache() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := pebbleCache(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
