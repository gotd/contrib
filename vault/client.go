package vault

import (
	"context"
	"path"

	"github.com/go-faster/errors"
	"github.com/hashicorp/vault/api"
	"go.uber.org/multierr"

	"github.com/gotd/contrib/auth/kv"
)

type vaultClient struct {
	client *api.Client
	path   string
}

func (c vaultClient) Set(ctx context.Context, k, v string) error {
	return c.add(ctx, k, v)
}

func (c vaultClient) Get(ctx context.Context, k string) (string, error) {
	return c.get(ctx, k)
}

func (c vaultClient) getPath() string {
	return path.Join("/v1/", c.path)
}

func (c vaultClient) putAll(ctx context.Context, data map[string]interface{}) error {
	req := c.client.NewRequest("PUT", c.getPath())

	err := req.SetJSONBody(data)
	if err != nil {
		return errors.Errorf("request encode: %w", err)
	}

	resp, err := c.client.RawRequestWithContext(ctx, req)
	if resp != nil {
		defer func() {
			multierr.AppendInto(&err, resp.Body.Close())
		}()
	}
	if err != nil {
		return errors.Errorf("secret send: %w", err)
	}

	return nil
}

func (c vaultClient) add(ctx context.Context, key, value string) error {
	s, err := c.getAll(ctx)
	data := map[string]interface{}{}
	if err != nil && !errors.Is(err, kv.ErrKeyNotFound) {
		return err
	}
	if err == nil {
		data = s.Data
	}

	data[key] = value
	return c.putAll(ctx, data)
}

func (c vaultClient) getAll(ctx context.Context) (*api.Secret, error) {
	req := c.client.NewRequest("GET", c.getPath())

	resp, err := c.client.RawRequestWithContext(ctx, req)
	if resp != nil {
		defer func() {
			multierr.AppendInto(&err, resp.Body.Close())
		}()

		if resp.StatusCode == 404 {
			return nil, kv.ErrKeyNotFound
		}
	}
	if err != nil {
		return nil, errors.Errorf("secret fetch: %w", err)
	}

	secret, err := api.ParseSecret(resp.Body)
	if err != nil {
		return nil, errors.Errorf("secret parsing: %w", err)
	}

	return secret, nil
}

func (c vaultClient) get(ctx context.Context, key string) (string, error) {
	secret, err := c.getAll(ctx)
	if err != nil {
		return "", err
	}

	data, ok := secret.Data[key]
	if !ok {
		return "", kv.ErrKeyNotFound
	}

	session, ok := data.(string)
	if !ok {
		return "", errors.Errorf("expected %q have string type, got %T", key, data)
	}

	return session, nil
}
