package vault

import (
	"github.com/hashicorp/vault/api"

	"github.com/gotd/td/session"

	"github.com/gotd/contrib/auth/kv"
)

var _ session.Storage = SessionStorage{}

// SessionStorage is a MTProto session Vault storage.
type SessionStorage struct {
	kv.Session
}

// NewSessionStorage creates new SessionStorage.
func NewSessionStorage(client *api.Client, path, key string) SessionStorage {
	s := vaultClient{client: client, path: path}
	return SessionStorage{
		Session: kv.NewSession(s, key),
	}
}
