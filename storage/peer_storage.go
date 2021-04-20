package storage

import (
	"context"

	"golang.org/x/xerrors"
)

// ErrPeerNotFound is a special error to return when peer not found.
var ErrPeerNotFound = xerrors.New("peer not found")

// PeerStorage is abstraction for peer storage.
type PeerStorage interface {
	// Add adds given peer to the storage.
	Add(ctx context.Context, p Peer) error
	// Find finds peer using given key.
	// If peer not found, it returns ErrPeerNotFound error.
	Find(ctx context.Context, key Key) (Peer, error)

	// Assign adds given peer to the storage and associates it to the given key.
	Assign(ctx context.Context, key string, p Peer) error
	// Resolve finds peer using associated key.
	// If peer not found, it returns ErrPeerNotFound error.
	Resolve(ctx context.Context, key string) (Peer, error)
}
