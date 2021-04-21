package storage

import (
	"context"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

// FindPeer finds peer using given storage.
func FindPeer(ctx context.Context, s PeerStorage, p tg.PeerClass) (Peer, error) {
	var key peer.DialogKey

	if err := key.FromPeer(p); err != nil {
		return Peer{}, err
	}

	return s.Find(ctx, Key{
		Kind: key.Kind,
		ID:   key.ID,
	})
}
