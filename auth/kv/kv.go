package kv

import (
	"context"

	"golang.org/x/xerrors"
)

// Storage represents generic KV storage.
type Storage interface {
	Set(ctx context.Context, k, v string) error
	Get(ctx context.Context, k string) (string, error)
}

// ErrKeyNotFound is a special error to return then given key not found.
var ErrKeyNotFound = xerrors.New("key not found")
