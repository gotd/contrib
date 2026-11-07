package pebble

import (
	"context"
	"encoding/binary"

	"github.com/cockroachdb/pebble"
	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/peers"
)

var _ peers.Storage = (*PeersStorage)(nil)

// Key prefixes of peers.Storage.
//
// They must not start with storage.PeerKeyPrefix ("peer"), which PeerStorage
// iterates over.
const (
	prefixPeersKey      = "tdpeers_key_"
	prefixPeersPhone    = "tdpeers_phone_"
	keyPeersContactHash = "tdpeers_contacts_hash"
)

// PeersStorage is peers.Storage implementation using pebble.
//
// It stores access hashes for peers.Manager of
// github.com/gotd/td/telegram/peers, unlike PeerStorage, which stores
// storage.Peer of github.com/gotd/contrib/storage.
type PeersStorage struct {
	db        *pebble.DB
	writeOpts *pebble.WriteOptions
}

// NewPeersStorage creates new peers.Storage using pebble.
//
// Caller is responsible for db.Close() invocation.
func NewPeersStorage(db *pebble.DB) *PeersStorage {
	return &PeersStorage{db: db}
}

// WithWriteOptions sets pebble's write options for write operations.
//
// Default is nil, which pebble treats as pebble.Sync.
func (s *PeersStorage) WithWriteOptions(writeOpts *pebble.WriteOptions) *PeersStorage {
	s.writeOpts = writeOpts
	return s
}

// encodePeersKey encodes key as prefix, 0 separator and big-endian ID, so
// prefixes can't collide.
func encodePeersKey(b []byte, key peers.Key) []byte {
	b = append(b, key.Prefix...)
	b = append(b, 0)
	return binary.BigEndian.AppendUint64(b, uint64(key.ID))
}

func decodePeersKey(b []byte) (peers.Key, error) {
	if len(b) < 9 || b[len(b)-9] != 0 {
		return peers.Key{}, errors.Errorf("invalid encoded key %x", b)
	}
	return peers.Key{
		Prefix: string(b[:len(b)-9]),
		ID:     int64(binary.BigEndian.Uint64(b[len(b)-8:])),
	}, nil
}

// Save implements peers.Storage.
func (s *PeersStorage) Save(_ context.Context, key peers.Key, value peers.Value) error {
	k := encodePeersKey([]byte(prefixPeersKey), key)
	return s.db.Set(k, binary.BigEndian.AppendUint64(nil, uint64(value.AccessHash)), s.writeOpts)
}

// Find implements peers.Storage.
func (s *PeersStorage) Find(_ context.Context, key peers.Key) (value peers.Value, found bool, err error) {
	v, ok, err := get(s.db, encodePeersKey([]byte(prefixPeersKey), key))
	if err != nil || !ok {
		return peers.Value{}, false, err
	}
	hash, err := decodeInt64(v)
	if err != nil {
		return peers.Value{}, false, errors.Wrapf(err, "access hash of %s%d", key.Prefix, key.ID)
	}
	return peers.Value{AccessHash: hash}, true, nil
}

// SavePhone implements peers.Storage.
func (s *PeersStorage) SavePhone(_ context.Context, phone string, key peers.Key) error {
	return s.db.Set([]byte(prefixPeersPhone+phone), encodePeersKey(nil, key), s.writeOpts)
}

// FindPhone implements peers.Storage.
func (s *PeersStorage) FindPhone(ctx context.Context, phone string) (key peers.Key, value peers.Value, found bool, err error) {
	v, ok, err := get(s.db, []byte(prefixPeersPhone+phone))
	if err != nil || !ok {
		return peers.Key{}, peers.Value{}, false, err
	}
	key, err = decodePeersKey(v)
	if err != nil {
		return peers.Key{}, peers.Value{}, false, errors.Wrap(err, "phone key")
	}
	value, found, err = s.Find(ctx, key)
	if err != nil || !found {
		return peers.Key{}, peers.Value{}, false, err
	}
	return key, value, true, nil
}

// GetContactsHash implements peers.Storage.
func (s *PeersStorage) GetContactsHash(context.Context) (int64, error) {
	v, ok, err := get(s.db, []byte(keyPeersContactHash))
	if err != nil || !ok {
		return 0, err
	}
	hash, err := decodeInt64(v)
	if err != nil {
		return 0, errors.Wrap(err, "contacts hash")
	}
	return hash, nil
}

// SaveContactsHash implements peers.Storage.
func (s *PeersStorage) SaveContactsHash(_ context.Context, hash int64) error {
	return s.db.Set([]byte(keyPeersContactHash), binary.BigEndian.AppendUint64(nil, uint64(hash)), s.writeOpts)
}
