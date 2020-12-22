package vault

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/api"
)

var errSecretNotFound = errors.New("secret not found")

type vaultClient struct {
	*api.Client
}

func (c vaultClient) put(ctx context.Context, path, key, value string) error {
	req := c.NewRequest("PUT", "/v1/"+path)

	err := req.SetJSONBody(map[string]interface{}{
		key: value,
	})
	if err != nil {
		return fmt.Errorf("request encode: %w", err)
	}

	resp, err := c.RawRequestWithContext(ctx, req)
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if err != nil {
		return fmt.Errorf("secret send: %w", err)
	}

	return nil
}

func (c vaultClient) get(ctx context.Context, path, key string) (string, error) {
	req := c.NewRequest("GET", "/v1/"+path)

	resp, err := c.RawRequestWithContext(ctx, req)
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if resp != nil && resp.StatusCode == 404 {
		return "", errSecretNotFound
	}
	if err != nil {
		return "", fmt.Errorf("secret fetch: %w", err)
	}

	secret, err := api.ParseSecret(resp.Body)
	if err != nil {
		return "", fmt.Errorf("secret parsing: %w", err)
	}

	data, ok := secret.Data[key]
	if !ok {
		return "", errSecretNotFound
	}

	session, ok := data.(string)
	if !ok {
		return "", fmt.Errorf("expected %q have string type, got %T", key, data)
	}

	return session, nil
}
