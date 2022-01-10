package storage_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/pebble"
	"github.com/gotd/contrib/storage"
)

func updatesHook(ctx context.Context) error {
	db, err := pebbledb.Open("pebble.db", &pebbledb.Options{})
	if err != nil {
		return errors.Errorf("create pebble storage: %w", err)
	}
	s := pebble.NewPeerStorage(db)

	dispatcher := tg.NewUpdateDispatcher()
	handler := storage.UpdateHook(dispatcher, s)
	client, err := telegram.ClientFromEnvironment(telegram.Options{
		UpdateHandler: handler,
	})
	if err != nil {
		return errors.Errorf("create client: %w", err)
	}
	raw := tg.NewClient(client)
	sender := message.NewSender(raw)

	dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok {
			return nil
		}

		// Use PeerID to find peer because *Short updates does not contain any entities, so it necessary to
		// store some entities.
		// Storage can be filled using PeerCollector.
		p, err := storage.FindPeer(ctx, s, msg.GetPeerID())
		if err != nil {
			return err
		}

		_, err = sender.To(p.AsInputPeer()).Text(ctx, msg.GetMessage())
		return err
	})

	return client.Run(ctx, func(ctx context.Context) error {
		return telegram.RunUntilCanceled(ctx, client)
	})
}

func ExampleUpdateHook() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := updatesHook(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
