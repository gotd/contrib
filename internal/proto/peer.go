package proto

import (
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

// LatestVersion is a latest supported version of data.
const LatestVersion = 1

// Peer is abstraction for peer object.
type Peer struct {
	Version int
	Key     peer.DialogKey
	User    *tg.User    `json:",omitempty"`
	Chat    *tg.Chat    `json:",omitempty"`
	Channel *tg.Channel `json:",omitempty"`
}

// FromInputPeer creates Peer object from given tg.InputPeerClass.
func FromInputPeer(p tg.InputPeerClass) (Peer, error) {
	k := peer.DialogKey{}
	if err := k.FromInputPeer(p); err != nil {
		return Peer{}, xerrors.Errorf("unpack input peer: %w", err)
	}

	return Peer{
		Version: LatestVersion,
		Key:     k,
	}, nil
}
