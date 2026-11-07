package storage

import (
	"context"
	"io"

	"github.com/go-faster/errors"
	"go.uber.org/multierr"
)

// ErrPeerNotFound is a special error to return when peer not found.
var ErrPeerNotFound = errors.New("peer not found")

// PeerStorage is abstraction for peer storage.
type PeerStorage interface {
	// Add adds given peer to the storage.
	Add(ctx context.Context, value Peer) error
	// Find finds peer using given key.
	// If peer not found, it returns ErrPeerNotFound error.
	Find(ctx context.Context, key PeerKey) (Peer, error)

	// Assign adds given peer to the storage and associates it to the given key.
	Assign(ctx context.Context, key string, value Peer) error
	// Resolve finds peer using associated key.
	// If peer not found, it returns ErrPeerNotFound error.
	Resolve(ctx context.Context, key string) (Peer, error)

	// Iterate creates and returns new PeerIterator.
	Iterate(ctx context.Context) (PeerIterator, error)
}

// BatchPeerStorage is a PeerStorage that can add many peers at once,
// e.g. in a single transaction.
//
// UpdateHook and PeerCollector use it, if implemented, to write all peers
// of an update or of a batch with one write.
type BatchPeerStorage interface {
	PeerStorage
	// AddMany adds given peers to the storage.
	AddMany(ctx context.Context, values []Peer) error
}

// addMany adds peers using AddMany, if supported, or Add.
func addMany(ctx context.Context, s PeerStorage, values []Peer) error {
	if len(values) == 0 {
		return nil
	}
	if b, ok := s.(BatchPeerStorage); ok {
		return b.AddMany(ctx, values)
	}
	var rerr error
	for _, value := range values {
		multierr.AppendInto(&rerr, s.Add(ctx, value))
	}
	return rerr
}

// PeerIterator is a peer iterator.
type PeerIterator interface {
	Next(ctx context.Context) bool
	Err() error
	Value() Peer
	io.Closer
}
