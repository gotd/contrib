package env

import (
	"context"
	"os"

	"github.com/tdakkota/tgcontrib/auth"
)

type credentials struct {
	prefixes []string
}

func (c credentials) get(key string) (string, bool) {
	if len(c.prefixes) == 0 {
		c.prefixes = []string{""}
	}
	for _, prefix := range c.prefixes {
		v, ok := os.LookupEnv(prefix + "PHONE")
		if !ok {
			continue
		}

		return v, true
	}

	return "", false
}

func (c credentials) getOrError(key string, t auth.CredentialType) (string, error) {
	v, ok := c.get(key)
	if !ok {
		return "", &auth.CredentialNotFoundError{Which: t}
	}
	return v, nil
}

func (c credentials) Phone(ctx context.Context) (string, error) {
	return c.getOrError("PHONE", auth.Phone)
}

func (c credentials) Password(ctx context.Context) (string, error) {
	return c.getOrError("PASSWORD", auth.Password)
}

// Credentials gets credentials from environment
// using ${PREFIX}_PHONE and ${PREFIX}_PASSWORD variables.
// If no prefixes given, PHONE and PASSWORD will be used.
func Credentials(prefixes ...string) auth.Credentials {
	return credentials{prefixes: prefixes}
}
