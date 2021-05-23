package main

import (
	"context"
	"flag"
	"fmt"
	"go/build"
	"os"
	"os/signal"
	"reflect"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"

	"github.com/gotd/contrib/auth/terminal"
	"github.com/gotd/contrib/binding/yaegi"
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

func parseOptions() (telegram.Options, *flag.FlagSet) {
	set := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	options := telegram.Options{}

	set.IntVar(&options.DC, "dc", 2, "Telegram DC ID")
	set.BoolVar(&options.NoUpdates, "noupdates", false, "Disable updates")

	return options, set
}

func run(ctx context.Context) error {
	options, set := parseOptions()
	if err := set.Parse(os.Args[1:]); err != nil {
		return err
	}

	sessionFile := flag.String("session", "", "Path to session file")
	testDC := flag.Bool("test", false, "Use test DC list")
	if sessionFile != nil && *sessionFile != "" {
		options.SessionStorage = &session.FileStorage{Path: *sessionFile}
	}
	if *testDC {
		options.DCList = dcs.Staging()
	}

	client, err := telegram.ClientFromEnvironment(options)
	if err != nil {
		return fmt.Errorf("setup client: %w", err)
	}

	return client.Run(ctx, func(ctx context.Context) error {
		if err := client.AuthIfNecessary(ctx, telegram.NewAuth(
			telegram.EnvAuth("", telegram.CodeAuthenticatorFunc(terminal.OS().Code)),
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
