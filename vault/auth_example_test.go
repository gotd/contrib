package vault_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/go-faster/errors"
	"github.com/hashicorp/vault/api"

	"github.com/gotd/td/telegram"
	tgauth "github.com/gotd/td/telegram/auth"

	"github.com/gotd/contrib/auth"
	"github.com/gotd/contrib/auth/terminal"
	"github.com/gotd/contrib/vault"
)

func vaultAuth(ctx context.Context) error {
	vaultClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return errors.Errorf("create Vault client: %w", err)
	}
	cred := vault.NewCredentials(vaultClient, "cubbyhole/telegram/user").
		WithPhoneKey("phone").
		WithPasswordKey("password")

	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return errors.Errorf("create client: %w", err)
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

	if err := vaultAuth(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
