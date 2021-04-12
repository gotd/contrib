package vault_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/hashicorp/vault/api"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"

	"github.com/gotd/contrib/auth/vault"
)

func vaultStorage(ctx context.Context) error {
	vaultClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return xerrors.Errorf("create Vault client: %w", err)
	}
	storage := vault.NewSessionStorage(vaultClient, "cubbyhole/telegram/user", "session")

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

	if err := vaultStorage(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
