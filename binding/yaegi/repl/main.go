package main

import (
	"context"
	"flag"
	"fmt"
	"go/build"
	"os"
	"os/signal"
	"reflect"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"

	"github.com/tdakkota/tgcontrib/auth/terminal"
	"github.com/tdakkota/tgcontrib/binding/yaegi"
)

func setupInterp(ctx context.Context, client *telegram.Client) (*interp.Interpreter, error) {
	i := interp.New(interp.Options{
		GoPath: build.Default.GOPATH,
	})

	i.Use(stdlib.Symbols)
	i.Use(interp.Symbols)
	i.Use(yaegi.Symbols)

	i.Use(map[string]map[string]reflect.Value{
		"repl": helpers(ctx, client),
	})
	_, err := i.EvalWithContext(ctx, "import . \"repl\"")
	if err != nil {
		return nil, err
	}

	return i, nil
}

func run(ctx context.Context) error {
	options := telegram.Options{}

	sessionFile := flag.String("session", "", "path to session file")
	flag.IntVar(&options.DC, "dc", 2, "Telegram DC ID")
	flag.StringVar(&options.Addr, "addr", telegram.AddrProduction, "Telegram DC address")
	if sessionFile != nil && *sessionFile != "" {
		options.SessionStorage = &session.FileStorage{Path: *sessionFile}
	}

	client, err := telegram.ClientFromEnvironment(options)
	if err != nil {
		return fmt.Errorf("setup client: %w", err)
	}

	return client.Run(ctx, func(ctx context.Context) error {
		if err := client.AuthIfNecessary(ctx, telegram.NewAuth(
			telegram.EnvAuth("", telegram.CodeAuthenticatorFunc(terminal.NewTerminal().Code)),
			telegram.SendCodeOptions{},
		)); err != nil {
			return fmt.Errorf("auth: %w", err)
		}

		i, err := setupInterp(ctx, client)
		if err != nil {
			return fmt.Errorf("setup interpreter: %w", err)
		}

		v, err := i.REPL()
		if err != nil {
			return fmt.Errorf("repl: %w", err)
		}
		fmt.Println(v)
		return nil
	})
}

func main() {
	ctx := context.Background()
	signal.NotifyContext(ctx, os.Interrupt)

	if err := run(ctx); err != nil {
		fmt.Println(err)
		return
	}
}
