package bbolt_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	bboltdb "go.etcd.io/bbolt"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	tgauth "github.com/gotd/td/telegram/auth"

	"github.com/gotd/contrib/auth"
	"github.com/gotd/contrib/auth/terminal"
	"github.com/gotd/contrib/bbolt"
)

func bboltAuth(ctx context.Context) error {
	db, err := bboltdb.Open("bbolt.db", 0666, &bboltdb.Options{}) // nolint:gocritic
	if err != nil {
		return xerrors.Errorf("create bbolt storage: %w", err)
	}
	cred := bbolt.NewCredentials(db, []byte("bucket")).
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

	if err := bboltAuth(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
