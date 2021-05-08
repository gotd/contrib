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
	User      *tg.User               `json:",omitempty"`
	Chat      *tg.Chat               `json:",omitempty"`
	Channel   *tg.Channel            `json:",omitempty"`
	Metadata  map[string]interface{} `json:",omitempty"`
}

func addIfNotEmpty(r []string, k string) []string {
	if k == "" {
		return r
	}
	return append(r, k)
}

// Keys returns list of all associated keys (phones, usernames, etc.) stored in the peer.
func (p *Peer) Keys() []string {
	// Chat does not contain usernames or phones.
	if p.Chat != nil {
		return nil
	}

	r := make([]string, 0, 4)
	switch {
	case p.User != nil:
		r = addIfNotEmpty(r, p.User.Username)
		r = addIfNotEmpty(r, p.User.Phone)
	case p.Channel != nil:
		r = addIfNotEmpty(r, p.Channel.Username)
	}

	return r
}

// FromInputPeer fills Peer object using given tg.InputPeerClass.
func (p *Peer) FromInputPeer(input tg.InputPeerClass) error {
	k := peer.DialogKey{}
	if err := k.FromInputPeer(input); err != nil {
		return xerrors.Errorf("unpack input peer: %w", err)
	}

	*p = Peer{
		Version:   LatestVersion,
		Key:       k,
		CreatedAt: time.Now().Unix(),
	}

	return nil
}

// FromChat fills Peer object using given tg.ChatClass.
func (p *Peer) FromChat(chat tg.ChatClass) bool {
	r := Peer{
		Version:   LatestVersion,
		CreatedAt: time.Now().Unix(),
	}

	switch c := chat.(type) {
	case *tg.Chat:
		r.Key.ID = c.ID
		r.Key.Kind = peer.Chat
		r.Chat = c
	case *tg.ChatForbidden:
		r.Key.ID = c.ID
		r.Key.Kind = peer.Chat
	case *tg.Channel:
		if c.Min {
			return false
		}
		r.Key.ID = c.ID
		r.Key.AccessHash = c.AccessHash
		r.Key.Kind = peer.Channel
		r.Channel = c
	case *tg.ChannelForbidden:
		r.Key.ID = c.ID
		r.Key.AccessHash = c.AccessHash
		r.Key.Kind = peer.Channel
	default:
		return false
	}

	*p = r
	return true
}

// FromUser fills Peer object using given tg.UserClass.
func (p *Peer) FromUser(user tg.UserClass) bool {
	u, ok := user.AsNotEmpty()
	if !ok {
		return false
	}

	*p = Peer{
		Version: LatestVersion,
		Key: peer.DialogKey{
			Kind:       peer.User,
			ID:         u.ID,
			AccessHash: u.AccessHash,
		},
		CreatedAt: time.Now().Unix(),
		User:      u,
	}

	return true
}
