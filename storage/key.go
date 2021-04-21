package storage

import (
	"bytes"
	"strconv"
	"strings"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message/peer"
)

// keyPrefix is a key prefix of peer key.
var keyPrefix = []byte("peer") // nolint:gochecknoglobals

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
	r = append(r, keyPrefix...)
	r = strconv.AppendInt(r, int64(k.Kind), 10)
	r = append(r, keySeparator)
	r = strconv.AppendInt(r, int64(k.ID), 10)
	return r
}

// String returns string representation of key.
func (k Key) String() string {
	var (
		b   strings.Builder
		buf [64]byte
	)
	b.Write(keyPrefix)
	b.Write(strconv.AppendInt(buf[:0], int64(k.Kind), 10))
	b.WriteRune(keySeparator)
	b.Write(strconv.AppendInt(buf[:0], int64(k.ID), 10))
	return b.String()
}

var invalidKey = xerrors.New("invalid key") // nolint:gochecknoglobals

// Parse parses bytes representation from given slice.
func (k *Key) Parse(r []byte) error {
	if !bytes.HasPrefix(r, keyPrefix) {
		return invalidKey
	}
	r = r[len(keyPrefix):]

	idx := bytes.IndexByte(r, keySeparator)
	// Check that slice contains _ and it's not a first or last character.
	if idx <= 0 || idx == len(r)-1 {
		return invalidKey
	}

	{
		v, err := strconv.Atoi(string(r[:idx]))
		if err != nil {
			return xerrors.Errorf("parse kind: %w", err)
		}
		if v > int(peer.Channel) {
			return xerrors.Errorf("invalid kind %d", v)
		}
		k.Kind = peer.Kind(v)
	}

	{
		v, err := strconv.Atoi(string(r[idx+1:]))
		if err != nil {
			return xerrors.Errorf("parse id: %w", err)
		}
		k.ID = v
	}

	return nil
}
