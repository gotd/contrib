package proto

import (
	"bytes"
	"strconv"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message/peer"
)

// Key is unique key of peer object.
type Key struct {
	Kind peer.Kind
	ID   int
}

// KeyFromPeer creates key from peer.
func KeyFromPeer(p Peer) Key {
	return Key{
		Kind: p.Key.Kind,
		ID:   p.Key.ID,
	}
}

const keySeparator = '_'

// Bytes returns bytes representation of key.
func (k Key) Bytes(r []byte) []byte {
	r = strconv.AppendInt(r, int64(k.Kind), 10)
	r = append(r, keySeparator)
	r = strconv.AppendInt(r, int64(k.ID), 10)
	return r
}

var invalidKey = xerrors.New("invalid key") // nolint:gochecknoglobals

// Parse parses bytes representation from given slice.
func (k *Key) Parse(r []byte) error {
	idx := bytes.IndexByte(r, keySeparator)
	// Check that slice contains _ and it's not a first or last character.
	if idx <= 0 || idx == len(r)-1 {
		return invalidKey
	}

	{
		v, err := strconv.ParseInt(string(r[:idx]), 10, 64)
		if err != nil {
			return xerrors.Errorf("parse kind: %w", err)
		}
		if v > int64(peer.Channel) {
			return xerrors.Errorf("invalid kind %d", v)
		}
		k.Kind = peer.Kind(v)
	}

	{
		v, err := strconv.ParseInt(string(r[idx:]), 10, 64)
		if err != nil {
			return xerrors.Errorf("parse id: %w", err)
		}
		k.ID = int(v)
	}

	return nil
}
