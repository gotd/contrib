package vault

import (
	"github.com/hashicorp/vault/api"

	"github.com/gotd/contrib/auth/kv"
)

// Credentials stores user credentials to Vault.
type Credentials struct {
	kv.Credentials
}

// NewCredentials creates new Credentials.
func NewCredentials(client *api.Client, path string) Credentials {
	s := vaultClient{client: client, path: path}
	return Credentials{
		Credentials: kv.NewCredentials(s),
	}
}
