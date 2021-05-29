package pebble_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	pebbledb "github.com/cockroachdb/pebble"
	"golang.org/x/xerrors"

	tgauth "github.com/gotd/td/telegram/auth"

	"github.com/gotd/td/telegram"

	"github.com/gotd/contrib/auth"
	"github.com/gotd/contrib/auth/terminal"
	"github.com/gotd/contrib/pebble"
)

func pebbleAuth(ctx context.Context) error {
	db, err := pebbledb.Open("pebble.db", &pebbledb.Options{})
	if err != nil {
		return xerrors.Errorf("create pebble storage: %w", err)
	}
	cred := pebble.NewCredentials(db).
		WithPhoneKey("phone").
		WithPasswordKey("password")

	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return xerrors.Errorf("create client: %w", err)
	}

	return client.Run(ctx, func(ctx context.Context) error {
		return client.Auth().IfNecessary(
			ctx,
			tgauth.NewFlow(auth.Build(cred, terminal.OS()), tgauth.SendCodeOptions{}),
		)
	})
}

func ExampleCredentials() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := pebbleAuth(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
