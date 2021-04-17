package pebble

import (
	"context"

	"github.com/cockroachdb/pebble"
	"golang.org/x/xerrors"

	"github.com/gotd/contrib/auth/kv"
)

type pebbleStorage struct {
	db   *pebble.DB
	opts *pebble.WriteOptions
}

func (p pebbleStorage) Set(ctx context.Context, k, v string) error {
	return p.db.Set(s2b(k), s2b(v), p.opts)
}

func (p pebbleStorage) Get(ctx context.Context, k string) (string, error) {
	r, closer, err := p.db.Get(s2b(k))
	if err != nil {
		if xerrors.Is(err, pebble.ErrNotFound) {
			return "", kv.ErrKeyNotFound
		}
		return "", err
	}
	v := string(r)

	return v, closer.Close()
}
