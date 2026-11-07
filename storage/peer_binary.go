package storage

import (
	"encoding/json"
	"time"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
)

// peerFormatBinary prefixes the compact binary encoding of Peer.
//
// The JSON encoding always starts with '{', so both can be told apart.
const peerFormatBinary byte = 0x01

// Entity flags of the binary encoding.
const (
	peerHasUser = 1 << iota
	peerHasChat
	peerHasChannel
)

// EncodePeer encodes peer in the compact binary format.
//
// Use DecodePeer to decode it: it also accepts the JSON format of
// Peer.MarshalJSON, so values stored before stay readable.
func EncodePeer(p Peer) ([]byte, error) {
	var metadata []byte
	if p.Metadata != nil {
		raw, err := json.Marshal(p.Metadata)
		if err != nil {
			return nil, errors.Wrap(err, "marshal metadata")
		}
		metadata = raw
	}

	b := &bin.Buffer{Buf: make([]byte, 0, 256)}
	b.Buf = append(b.Buf, peerFormatBinary)
	b.PutInt(p.Version)
	b.PutInt(int(p.Key.Kind))
	b.PutLong(p.Key.ID)
	b.PutLong(p.Key.AccessHash)
	b.PutLong(p.CreatedAt.Unix())
	b.PutBytes(metadata)

	var flags int
	if p.User != nil {
		flags |= peerHasUser
	}
	if p.Chat != nil {
		flags |= peerHasChat
	}
	if p.Channel != nil {
		flags |= peerHasChannel
	}
	b.PutInt(flags)
	for _, obj := range []bin.Encoder{p.User, p.Chat, p.Channel} {
		if obj == nil || isNilEncoder(obj) {
			continue
		}
		if err := obj.Encode(b); err != nil {
			return nil, errors.Wrap(err, "encode entity")
		}
	}
	return b.Buf, nil
}

func isNilEncoder(e bin.Encoder) bool {
	switch v := e.(type) {
	case *tg.User:
		return v == nil
	case *tg.Chat:
		return v == nil
	case *tg.Channel:
		return v == nil
	default:
		return false
	}
}

// DecodePeer decodes peer encoded by EncodePeer or Peer.MarshalJSON.
//
// Like Peer.UnmarshalJSON, it returns ErrPeerUnmarshalMustInvalidate for
// outdated data, which must be treated as a cache miss.
func DecodePeer(data []byte, p *Peer) error {
	if len(data) == 0 {
		return errors.New("empty data")
	}
	if data[0] != peerFormatBinary {
		return json.Unmarshal(data, p)
	}

	b := &bin.Buffer{Buf: data[1:]}
	version, err := b.Int()
	if err != nil {
		return errors.Wrap(err, "version")
	}
	if version != LatestVersion {
		return ErrPeerUnmarshalMustInvalidate
	}

	var r Peer
	r.Version = version
	kind, err := b.Int()
	if err != nil {
		return errors.Wrap(err, "kind")
	}
	r.Key.Kind = dialogs.PeerKind(kind)
	if r.Key.ID, err = b.Long(); err != nil {
		return errors.Wrap(err, "id")
	}
	if r.Key.AccessHash, err = b.Long(); err != nil {
		return errors.Wrap(err, "access_hash")
	}
	createdAt, err := b.Long()
	if err != nil {
		return errors.Wrap(err, "created_at")
	}
	r.CreatedAt = time.Unix(createdAt, 0)
	metadata, err := b.Bytes()
	if err != nil {
		return errors.Wrap(err, "metadata")
	}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &r.Metadata); err != nil {
			return errors.Wrap(err, "unmarshal metadata")
		}
	}
	flags, err := b.Int()
	if err != nil {
		return errors.Wrap(err, "flags")
	}
	if flags&peerHasUser != 0 {
		r.User = new(tg.User)
		if err := r.User.Decode(b); err != nil {
			return decodeEntityErr(err, "user")
		}
	}
	if flags&peerHasChat != 0 {
		r.Chat = new(tg.Chat)
		if err := r.Chat.Decode(b); err != nil {
			return decodeEntityErr(err, "chat")
		}
	}
	if flags&peerHasChannel != 0 {
		r.Channel = new(tg.Channel)
		if err := r.Channel.Decode(b); err != nil {
			return decodeEntityErr(err, "channel")
		}
	}

	*p = r
	return nil
}

func decodeEntityErr(err error, name string) error {
	// Entity encoded with another TL layer.
	if _, ok := errors.Into[*bin.UnexpectedIDErr](err); ok {
		return ErrPeerUnmarshalMustInvalidate
	}
	return errors.Wrap(err, name)
}
