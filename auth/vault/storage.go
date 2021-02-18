package vault

import (
	"context"

	"github.com/gotd/td/session"
	"github.com/hashicorp/vault/api"
	"golang.org/x/xerrors"
)

var _ session.Storage = SessionStorage{}

// SessionStorage is a MTProto session Vault storage.
type SessionStorage struct {
	vault vaultClient
	path  string
	key   string
}

// NewSessionStorage creates new SessionStorage.
func NewSessionStorage(client *api.Client, path, key string) SessionStorage {
	return SessionStorage{vault: vaultClient{client}, path: path, key: key}
}

// LoadSession loads session from Vault.
func (s SessionStorage) LoadSession(ctx context.Context) ([]byte, error) {
	data, err := s.vault.get(ctx, s.path, s.key)
	if err != nil {
		if xerrors.Is(err, errSecretNotFound) {
			return nil, session.ErrNotFound
		}
		return nil, xerrors.Errorf("load session: %w", err)
	}

	return []byte(data), nil
}

// StoreSession stores session to Vault.
func (s SessionStorage) StoreSession(ctx context.Context, data []byte) error {
	if err := s.vault.put(ctx, s.path, s.key, string(data)); err != nil {
		return xerrors.Errorf("store session: %w", err)
	}

	return nil
}
