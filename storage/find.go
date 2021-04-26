package storage

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

// FindPeer finds peer using given storage.
func FindPeer(ctx context.Context, s PeerStorage, p tg.PeerClass) (Peer, error) {
	var key peer.DialogKey

	if err := key.FromPeer(p); err != nil {
		return Peer{}, err
	}

	return s.Find(ctx, PeerKey{
		Kind: key.Kind,
		ID:   key.ID,
	})
}

// ForEach calls callback on every iterator element.
func ForEach(ctx context.Context, iterator PeerIterator, cb func(Peer) error) error {
	for iterator.Next(ctx) {
		if err := cb(iterator.Value()); err != nil {
			return xerrors.Errorf("callback: %w", err)
		}
	}
	return iterator.Err()
}
