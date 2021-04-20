package etcd_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"go.etcd.io/etcd/client/v3"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"

	"github.com/gotd/contrib/etcd"
)

func etcdStorage(ctx context.Context) error {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return xerrors.Errorf("create etcd storage: %w", err)
	}
	storage := etcd.NewSessionStorage(c, "session")

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

	if err := etcdStorage(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
