package bbolt

import (
	"context"
	"encoding/binary"

	"github.com/go-faster/errors"
	bolt "go.etcd.io/bbolt"

	"github.com/gotd/td/telegram/peers"
)

var _ peers.Storage = (*PeersStorage)(nil)

// Bucket names of peers.Storage.
var (
	bucketPeersKeys   = []byte("tdpeers_keys")   // nolint:gochecknoglobals
	bucketPeersPhones = []byte("tdpeers_phones") // nolint:gochecknoglobals
	bucketPeersMeta   = []byte("tdpeers_meta")   // nolint:gochecknoglobals
	keyContactsHash   = []byte("contacts_hash")  // nolint:gochecknoglobals
)

// PeersStorage is peers.Storage implementation using bbolt.
//
// It stores access hashes for peers.Manager of
// github.com/gotd/td/telegram/peers, unlike PeerStorage, which stores
// storage.Peer of github.com/gotd/contrib/storage. It uses its own buckets,
// so it can share one db with other storages of this package.
type PeersStorage struct {
	db *bolt.DB
}

// NewPeersStorage creates new peers.Storage using bbolt.
//
// Caller is responsible for db.Close() invocation.
func NewPeersStorage(db *bolt.DB) *PeersStorage {
	return &PeersStorage{db: db}
}

// encodePeersKey encodes key as prefix, 0 separator and big-endian ID, so
// prefixes can't collide.
func encodePeersKey(key peers.Key) []byte {
	b := make([]byte, 0, len(key.Prefix)+9)
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

func encodeHash(v int64) []byte {
	return binary.BigEndian.AppendUint64(nil, uint64(v))
}

func decodeHash(b []byte) (int64, error) {
	if len(b) != 8 {
		return 0, errors.Errorf("invalid value length %d", len(b))
	}
	return int64(binary.BigEndian.Uint64(b)), nil
}

func (s *PeersStorage) put(bucket, key, value []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
}

// findHash returns access hash of key in tx.
func findHash(tx *bolt.Tx, key peers.Key) (peers.Value, bool, error) {
	b := tx.Bucket(bucketPeersKeys)
	if b == nil {
		return peers.Value{}, false, nil
	}
	v := b.Get(encodePeersKey(key))
	if v == nil {
		return peers.Value{}, false, nil
	}
	hash, err := decodeHash(v)
	if err != nil {
		return peers.Value{}, false, errors.Wrapf(err, "access hash of %s%d", key.Prefix, key.ID)
	}
	return peers.Value{AccessHash: hash}, true, nil
}

// Save implements peers.Storage.
func (s *PeersStorage) Save(_ context.Context, key peers.Key, value peers.Value) error {
	return s.put(bucketPeersKeys, encodePeersKey(key), encodeHash(value.AccessHash))
}

// Find implements peers.Storage.
func (s *PeersStorage) Find(_ context.Context, key peers.Key) (value peers.Value, found bool, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		value, found, err = findHash(tx, key)
		return err
	})
	return value, found, err
}

// SavePhone implements peers.Storage.
func (s *PeersStorage) SavePhone(_ context.Context, phone string, key peers.Key) error {
	return s.put(bucketPeersPhones, []byte(phone), encodePeersKey(key))
}

// FindPhone implements peers.Storage.
func (s *PeersStorage) FindPhone(_ context.Context, phone string) (key peers.Key, value peers.Value, found bool, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPeersPhones)
		if b == nil {
			return nil
		}
		v := b.Get([]byte(phone))
		if v == nil {
			return nil
		}
		k, err := decodePeersKey(v)
		if err != nil {
			return errors.Wrap(err, "phone key")
		}
		value, found, err = findHash(tx, k)
		if err != nil {
			return err
		}
		key = k
		return nil
	})
	if err != nil || !found {
		return peers.Key{}, peers.Value{}, false, err
	}
	return key, value, true, nil
}

// GetContactsHash implements peers.Storage.
func (s *PeersStorage) GetContactsHash(context.Context) (hash int64, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketPeersMeta)
		if b == nil {
			return nil
		}
		v := b.Get(keyContactsHash)
		if v == nil {
			return nil
		}
		h, err := decodeHash(v)
		if err != nil {
			return errors.Wrap(err, "contacts hash")
		}
		hash = h
		return nil
	})
	return hash, err
}

// SaveContactsHash implements peers.Storage.
func (s *PeersStorage) SaveContactsHash(_ context.Context, hash int64) error {
	return s.put(bucketPeersMeta, keyContactsHash, encodeHash(hash))
}
