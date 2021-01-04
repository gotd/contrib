package vault

import (
	"context"
	"errors"

	"github.com/gotd/td/telegram"
	"github.com/hashicorp/vault/api"

	"github.com/tdakkota/tgcontrib/auth"
)

const (
	phoneKey    = "phone"
	passwordKey = "password"
)

// Auth is telegram.UserAuthenticator implementation
type Auth struct {
	auth.Ask
	Credentials
}

var _ telegram.UserAuthenticator = Auth{}

// NewAuth creates new Auth.
func NewAuth(code auth.Ask, client *api.Client, path string) Auth {
	return Auth{
		Ask:         code,
		Credentials: NewCredentials(client, path),
	}
}

// Credentials stores user credentials to Vault.
type Credentials struct {
	vault vaultClient
	path  string
}

var _ auth.Credentials = Credentials{}

// NewCredentials creates new Credentials.
func NewCredentials(client *api.Client, path string) Credentials {
	return Credentials{
		vault: vaultClient{Client: client},
		path:  path,
	}
}

// SavePhone stores given phone to the Vault.
func (a Credentials) SavePhone(ctx context.Context, phone string) error {
	return a.vault.add(ctx, a.path, phoneKey, phone)
}

// SavePassword stores given password to the Vault.
func (a Credentials) SavePassword(ctx context.Context, password string) error {
	return a.vault.add(ctx, a.path, passwordKey, password)
}

// Phone loads phone from the Vault.
func (a Credentials) Phone(ctx context.Context) (p string, err error) {
	p, err = a.vault.get(ctx, a.path, phoneKey)
	if errors.Is(err, errSecretNotFound) {
		return "", &auth.CredentialNotFoundError{Which: auth.Phone}
	}
	return
}

// Password loads password from the Vault.
func (a Credentials) Password(ctx context.Context) (p string, err error) {
	p, err = a.vault.get(ctx, a.path, passwordKey)
	if errors.Is(err, errSecretNotFound) {
		return "", &auth.CredentialNotFoundError{Which: auth.Password}
	}
	return
}
