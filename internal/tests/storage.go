package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"

	"github.com/gotd/td/session"

	"github.com/gotd/contrib/auth"
	"github.com/gotd/contrib/storage"
)

// Credentials is a KV credential storage abstraction.
type Credentials interface {
	auth.Credentials
	SavePhone(ctx context.Context, phone string) error
	SavePassword(ctx context.Context, password string) error
}

// TestSessionStorage runs different tests for given session storage implementation.
func TestSessionStorage(t *testing.T, storage session.Storage) {
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
}

// TestCredentials runs different tests for given credentials storage implementation.
func TestCredentials(t *testing.T, cred Credentials) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

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

// TestPeerStorage runs different tests for given peer storage implementation.
func TestPeerStorage(t *testing.T, st storage.PeerStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	t.Run("PeerStorage", func(t *testing.T) {
		a := require.New(t)

		_, err := st.Resolve(ctx, "abc")
		a.ErrorIs(err, storage.ErrPeerNotFound)

		var p storage.Peer
		a.NoError(p.FromInputPeer(&tg.InputPeerUser{
			UserID:     10,
			AccessHash: 10,
		}))
		key := storage.KeyFromPeer(p)

		_, err = st.Find(ctx, key)
		a.ErrorIs(err, storage.ErrPeerNotFound)

		a.NoError(st.Add(ctx, p))
		_, err = st.Find(ctx, key)
		a.NoError(err)

		a.NoError(st.Assign(ctx, "abc", p))
		_, err = st.Resolve(ctx, "abc")
		a.NoError(err)
	})
}
