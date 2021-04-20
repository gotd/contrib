package storage

import (
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

// LatestVersion is a latest supported version of data.
const LatestVersion = 1

// Peer is abstraction for peer object.
type Peer struct {
	Version   int
	Key       peer.DialogKey
	CreatedAt int64
	User      *tg.User    `json:",omitempty"`
	Chat      *tg.Chat    `json:",omitempty"`
	Channel   *tg.Channel `json:",omitempty"`
}

// FromInputPeer fills Peer object using given tg.InputPeerClass.
func (p *Peer) FromInputPeer(input tg.InputPeerClass) error {
	k := peer.DialogKey{}
	if err := k.FromInputPeer(input); err != nil {
		return xerrors.Errorf("unpack input peer: %w", err)
	}

	*p = Peer{
		Version:   LatestVersion,
		CreatedAt: time.Now().Unix(),
		Key:       k,
	}

	return nil
}
