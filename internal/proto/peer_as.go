package proto

import (
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

// AsInputUser tries to convert peer to *tg.InputUser.
func (p Peer) AsInputUser() (*tg.InputUser, bool) {
	if p.Key.Kind != peer.User {
		return nil, false
	}

	return &tg.InputUser{
		UserID:     p.Key.ID,
		AccessHash: p.Key.AccessHash,
	}, true
}

// AsInputChannel tries to convert peer to *tg.InputChannel.
func (p Peer) AsInputChannel() (*tg.InputChannel, bool) {
	if p.Key.Kind != peer.Channel {
		return nil, false
	}

	return &tg.InputChannel{
		ChannelID:  p.Key.ID,
		AccessHash: p.Key.AccessHash,
	}, true
}

// AsInputPeer tries to convert peer to tg.InputPeerClass.
func (p Peer) AsInputPeer() tg.InputPeerClass {
	switch p.Key.Kind {
	case peer.User:
		return &tg.InputPeerUser{
			UserID:     p.Key.ID,
			AccessHash: p.Key.AccessHash,
		}
	case peer.Chat:
		return &tg.InputPeerChat{
			ChatID: p.Key.ID,
		}
	case peer.Channel:
		return &tg.InputPeerChannel{
			ChannelID:  p.Key.ID,
			AccessHash: p.Key.AccessHash,
		}
	default:
		panic("unreachable")
	}
}
