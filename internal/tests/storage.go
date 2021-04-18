package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/auth"
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

type mockResolver struct {
	returnErr     bool
	domain, phone string
	peer          tg.InputPeerClass
	t             testing.TB
}

func (m *mockResolver) ResolveDomain(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	if m.returnErr {
		return nil, fmt.Errorf("test error: %q", m.domain)
	}

	if domain != m.domain {
		err := fmt.Errorf("expected domain %q, got %q", m.domain, domain)
		m.t.Error(err)
		return nil, err
	}
	return m.peer, nil
}

func (m *mockResolver) ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error) {
	if m.returnErr {
		return nil, fmt.Errorf("test error: %q", m.phone)
	}

	if phone != m.phone {
		err := fmt.Errorf("expected phone %q, got %q", m.phone, phone)
		m.t.Error(err)
		return nil, err
	}
	return m.peer, nil
}

// TestResolverCache runs different tests for given resolver cache storage implementation.
func TestResolverCache(t *testing.T, c func(next peer.Resolver) peer.Resolver) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	expectedDomain := "telegram"
	expectedPhone := "1223"
	expected := &tg.InputPeerUser{
		UserID: 10,
	}

	t.Run("Resolver", func(t *testing.T) {
		t.Run("Domain", func(t *testing.T) {
			a := require.New(t)
			resolver := &mockResolver{
				domain: expectedDomain,
				peer:   expected,
				t:      t,
			}
			cache := c(resolver)

			r, err := cache.ResolveDomain(ctx, expectedDomain)
			a.NoError(err)
			a.Equal(expected, r)

			r, err = cache.ResolveDomain(ctx, expectedDomain)
			a.NoError(err)
			a.Equal(expected, r)
		})

		t.Run("Phone", func(t *testing.T) {
			a := require.New(t)
			resolver := &mockResolver{
				phone: expectedPhone,
				peer:  expected,
				t:     t,
			}
			cache := c(resolver)

			r, err := cache.ResolvePhone(ctx, expectedPhone)
			a.NoError(err)
			a.Equal(expected, r)

			r, err = cache.ResolvePhone(ctx, expectedPhone)
			a.NoError(err)
			a.Equal(expected, r)
		})
	})
}
