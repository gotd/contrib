package etcd_test

import (
	"context"
	"os"
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/gotd/contrib/etcd"
	"github.com/gotd/contrib/internal/tests"
)

func TestE2E(t *testing.T) {
	addr := os.Getenv("ETCD_ADDR")
	if addr == "" {
		t.Skip("Set ETCD_ADDR to run E2E test")
	}

	var client *clientv3.Client
	tests.RetryUntilAvailable(t, "etcd", addr, func(ctx context.Context) error {
		c, err := clientv3.New(clientv3.Config{
			Endpoints: []string{addr},
		})
		if err != nil {
			return err
		}

		client = c
		return nil
	})

	tests.TestSessionStorage(t, etcd.NewSessionStorage(client, "session"))
	tests.TestCredentials(t, etcd.NewCredentials(client))
	tests.TestPeerStorage(t, etcd.NewPeerStorage(client).WithIterLimit(2))
}
