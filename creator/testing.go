package creator

import (
	"context"
	"crypto/rand"

	"github.com/gotd/td/telegram"
	"golang.org/x/xerrors"
)

// TestClient creates and authenticates user telegram.Client
// using Telegram staging server.
func TestClient(ctx context.Context, opts telegram.Options) (_ *telegram.Client, err error) {
	if opts.Addr == "" {
		opts.Addr = telegram.AddrTest
	}

	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, opts)
	if err := client.Connect(ctx); err != nil {
		return nil, xerrors.Errorf("connect: %w", err)
	}
	defer func() {
		if err != nil {
			_ = client.Close()
		}
	}()

	if err := telegram.NewAuth(
		telegram.TestAuth(rand.Reader, 2),
		telegram.SendCodeOptions{},
	).Run(ctx, client); err != nil {
		return nil, xerrors.Errorf("auth flow: %w", err)
	}

	return client, nil
}
