package vault_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"testing"

	"github.com/hashicorp/vault/api"

	"github.com/gotd/contrib/internal/tests"
	"github.com/gotd/contrib/vault"
)

func TestE2E(t *testing.T) {
	addr := os.Getenv("VAULT_ADDR")
	if addr == "" {
		t.Skip("Set VAULT_ADDR to run E2E test")
	}

	token, ok := os.LookupEnv("VAULT_TOKEN")
	if !ok {
		var data [16]byte
		if _, err := io.ReadFull(rand.Reader, data[:]); err != nil {
			t.Fatalf("Failed to generate token: %s", err)
		}
		token = hex.EncodeToString(data[:])
	}

	cfg := api.DefaultConfig()
	cfg.Address = addr
	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("Can't create client: %s", err)
		return
	}
	client.SetToken(token)

	tests.RetryUntilAvailable(t, "Vault", addr, func(ctx context.Context) error {
		_, err := client.Sys().Health()
		return err
	})

	tests.TestSessionStorage(t, vault.NewSessionStorage(client, "cubbyhole/testsession", "session"))
	tests.TestCredentials(t, vault.NewCredentials(client, "cubbyhole/testauth"))
}
