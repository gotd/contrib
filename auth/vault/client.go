package vault

import (
	"context"

	"github.com/hashicorp/vault/api"
	"golang.org/x/xerrors"
)

var errSecretNotFound = xerrors.New("secret not found")

type vaultClient struct {
	*api.Client
}

func (c vaultClient) putAll(ctx context.Context, path string, data map[string]interface{}) error {
	req := c.NewRequest("PUT", "/v1/"+path)

	err := req.SetJSONBody(data)
	if err != nil {
		return xerrors.Errorf("request encode: %w", err)
	}

	resp, err := c.RawRequestWithContext(ctx, req)
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if err != nil {
		return xerrors.Errorf("secret send: %w", err)
	}

	return nil
}

func (c vaultClient) put(ctx context.Context, path, key, value string) error {
	return c.putAll(ctx, path, map[string]interface{}{
		key: value,
	})
}

func (c vaultClient) add(ctx context.Context, path, key, value string) error {
	s, err := c.getAll(ctx, path, key)
	data := map[string]interface{}{}
	if err != nil && !xerrors.Is(err, errSecretNotFound) {
		return err
	}
	if err == nil {
		data = s.Data
	}

	data[key] = value
	return c.putAll(ctx, path, data)
}

func (c vaultClient) getAll(ctx context.Context, path, key string) (*api.Secret, error) {
	req := c.NewRequest("GET", "/v1/"+path)

	resp, err := c.RawRequestWithContext(ctx, req)
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if resp != nil && resp.StatusCode == 404 {
		return nil, errSecretNotFound
	}
	if err != nil {
		return nil, xerrors.Errorf("secret fetch: %w", err)
	}

	secret, err := api.ParseSecret(resp.Body)
	if err != nil {
		return nil, xerrors.Errorf("secret parsing: %w", err)
	}

	return secret, nil
}

func (c vaultClient) get(ctx context.Context, path, key string) (string, error) {
	secret, err := c.getAll(ctx, path, key)
	if err != nil {
		return "", err
	}

	data, ok := secret.Data[key]
	if !ok {
		return "", errSecretNotFound
	}

	session, ok := data.(string)
	if !ok {
		return "", xerrors.Errorf("expected %q have string type, got %T", key, data)
	}

	return session, nil
}
