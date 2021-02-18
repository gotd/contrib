package vault

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/hashicorp/vault/api"
	"golang.org/x/xerrors"
)

const (
	phoneKey    = "phone"
	passwordKey = "password"
)

// Credentials stores user credentials to Vault.
type Credentials struct {
	vault vaultClient
	path  string
}

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
	if xerrors.Is(err, errSecretNotFound) {
		return "", err
	}
	return
}

// Password loads password from the Vault.
func (a Credentials) Password(ctx context.Context) (p string, err error) {
	p, err = a.vault.get(ctx, a.path, passwordKey)
	if xerrors.Is(err, errSecretNotFound) {
		return "", telegram.ErrPasswordNotProvided
	}
	return
}
