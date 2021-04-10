package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/session"

	"github.com/gotd/contrib/auth"
)

// Credentials is a KV credential storage abstraction.
type Credentials interface {
	auth.Credentials
	SavePhone(ctx context.Context, phone string) error
	SavePassword(ctx context.Context, password string) error
}

// TestStorage runs different tests for given implementations.
func TestStorage(t *testing.T, storage session.Storage, cred Credentials) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("Session", func(t *testing.T) {
		a := require.New(t)

		data := []byte("mytoken")
		_, err := storage.LoadSession(ctx)
		a.Error(err, "no session expected")
		a.NoError(storage.StoreSession(ctx, data))

		vaultData, err := storage.LoadSession(ctx)
		a.NoError(err)
		a.Equal(data, vaultData)
	})

	t.Run("Credentials", func(t *testing.T) {
		a := require.New(t)

		phone, password := "phone", "password"
		a.NoError(cred.SavePhone(ctx, phone))
		a.NoError(cred.SavePassword(ctx, password))

		gotPhone, err := cred.Phone(ctx)
		a.NoError(err)
		a.Equal(phone, gotPhone)

		gotPassword, err := cred.Password(ctx)
		a.NoError(err)
		a.Equal(password, gotPassword)
	})
}
