package creator

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gotd/td/session"

	"github.com/tdakkota/tgcontrib/auth/terminal"

	"github.com/tdakkota/tgcontrib/auth"
	"github.com/tdakkota/tgcontrib/auth/env"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/transport"
	"golang.org/x/net/proxy"
	"golang.org/x/xerrors"
)

func sessionDir() (string, error) {
	dir, ok := os.LookupEnv("SESSION_DIR")
	if ok {
		return filepath.Abs(dir)
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		dir = "./"
	}

	return filepath.Abs(filepath.Join(dir, ".td"))
}

// OptionsFromEnvironment fills unfilled field in opts parameter
// using environment variables.
func OptionsFromEnvironment(opts telegram.Options) (telegram.Options, error) {
	// Setting up session storage if not provided.
	if opts.SessionStorage == nil {
		sessionFile, ok := os.LookupEnv("SESSION_FILE")
		if !ok {
			dir, err := sessionDir()
			if err != nil {
				return telegram.Options{}, xerrors.Errorf("SESSION_DIR not set or invalid: %w", err)
			}
			sessionFile = filepath.Join(dir, "session.join")
		}

		dir, _ := filepath.Split(sessionFile)
		if err := os.MkdirAll(dir, 0600); err != nil {
			return telegram.Options{}, xerrors.Errorf("session dir creation: %w", err)
		}

		opts.SessionStorage = &session.FileStorage{
			Path: sessionFile,
		}
	}

	if opts.Transport == nil {
		opts.Transport = transport.Intermediate(transport.DialFunc(proxy.Dial))
	}

	return opts, nil
}

// ClientFromEnvironment creates client using environment variables
// but not connects to server.
func ClientFromEnvironment(opts telegram.Options) (*telegram.Client, error) {
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		return nil, xerrors.Errorf("APP_ID not set or invalid: %w", err)
	}

	appHash := os.Getenv("APP_HASH")
	if appHash == "" {
		return nil, xerrors.New("no APP_HASH provided")
	}

	opts, err = OptionsFromEnvironment(opts)
	if err != nil {
		return nil, err
	}

	return telegram.NewClient(appID, appHash, opts), nil
}

// BotFromEnvironment creates bot client using environment variables
// connects to server and authenticates it.
func BotFromEnvironment(ctx context.Context, opts telegram.Options, cb ClientCallback) error {
	client, err := ClientFromEnvironment(opts)
	if err != nil {
		return err
	}

	return client.Run(ctx, func(ctx context.Context) error {
		status, err := client.AuthStatus(ctx)
		if err != nil {
			return xerrors.Errorf("auth status: %w", err)
		}

		if !status.Authorized {
			if err := client.AuthBot(ctx, os.Getenv("BOT_TOKEN")); err != nil {
				return xerrors.Errorf("login: %w", err)
			}
		}

		return cb(ctx, client)
	})
}

// UserFromEnvironment creates user client using environment variables
// connects to server and authenticates it.
func UserFromEnvironment(ctx context.Context, opts telegram.Options, ask auth.Ask, cb ClientCallback) error {
	client, err := ClientFromEnvironment(opts)
	if err != nil {
		return err
	}

	if ask == nil {
		ask = terminal.NewTerminal()
	}

	return client.Run(ctx, func(ctx context.Context) error {
		status, err := client.AuthStatus(ctx)
		if err != nil {
			return xerrors.Errorf("auth status: %w", err)
		}

		if !status.Authorized {
			flow := telegram.NewAuth(auth.Build(env.Credentials(), ask), telegram.SendCodeOptions{})
			if err := flow.Run(ctx, client); err != nil {
				return xerrors.Errorf("login: %w", err)
			}
		}

		return cb(ctx, client)
	})
}
