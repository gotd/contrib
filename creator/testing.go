package creator

import (
	"context"
	"crypto/rand"

	"github.com/gotd/td/telegram"
	"golang.org/x/xerrors"
)

// TestClient creates and authenticates user telegram.Client
// using Telegram staging server.
func TestClient(ctx context.Context, opts telegram.Options, cb ClientCallback) error {
	if opts.Addr == "" {
		opts.Addr = telegram.AddrTest
	}

	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, opts)
	return client.Run(ctx, func(ctx context.Context) error {
		if err := telegram.NewAuth(
			telegram.TestAuth(rand.Reader, 2),
			telegram.SendCodeOptions{},
		).Run(ctx, client); err != nil {
			return xerrors.Errorf("auth flow: %w", err)
		}

		return cb(ctx, client)
	})
}
