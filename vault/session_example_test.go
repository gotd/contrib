package vault_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/go-faster/errors"
	"github.com/hashicorp/vault/api"

	"github.com/gotd/td/telegram"

	"github.com/gotd/contrib/vault"
)

func vaultStorage(ctx context.Context) error {
	vaultClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return errors.Errorf("create Vault client: %w", err)
	}
	storage := vault.NewSessionStorage(vaultClient, "cubbyhole/telegram/user", "session")

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

	if err := vaultStorage(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
