package pebble

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/cockroachdb/pebble"
	"github.com/go-faster/errors"
	"go.uber.org/multierr"

	"github.com/gotd/contrib/storage"
)

var (
	_ storage.PeerStorage      = PeerStorage{}
	_ storage.BatchPeerStorage = PeerStorage{}
)

// PeerStorage is a peer storage based on pebble.
type PeerStorage struct {
	pebble    *pebble.DB
	writeOpts *pebble.WriteOptions
}

// NewPeerStorage creates new peer storage using pebble.
func NewPeerStorage(db *pebble.DB) *PeerStorage {
	return &PeerStorage{pebble: db}
}

// WithWriteOptions sets pebble's write options for write operations.
//
// Default is nil, which pebble treats as pebble.Sync: every write is synced
// to disk. Use pebble.NoSync to trade durability of the last writes for
// speed.
func (s *PeerStorage) WithWriteOptions(writeOpts *pebble.WriteOptions) *PeerStorage {
	s.writeOpts = writeOpts
	return s
}

type pebbleIterator struct {
	snap    *pebble.Snapshot
	iter    *pebble.Iterator
	lastErr error
	value   storage.Peer
}

func (p *pebbleIterator) Close() error {
	return multierr.Append(p.iter.Close(), p.snap.Close())
}

func (p *pebbleIterator) Next(ctx context.Context) bool {
	if !p.iter.Valid() {
		return false
	}

	for !bytes.HasPrefix(p.iter.Key(), storage.PeerKeyPrefix) {
		if !p.iter.Next() {
			return false
		}
	}

	if err := json.Unmarshal(p.iter.Value(), &p.value); err != nil {
		p.lastErr = errors.Wrap(err, "unmarshal")
		return false
	}

	p.iter.Next()
	return true
}

func (p *pebbleIterator) Err() error {
	return p.lastErr
}

func (p *pebbleIterator) Value() storage.Peer {
	return p.value
}

func keyUpperBound(b []byte) []byte {
	end := make([]byte, len(b))
	copy(end, b)
	for i := len(end) - 1; i >= 0; i-- {
		end[i]++
		if end[i] != 0 {
			return end[:i+1]
		}
	}
	return nil // no upper-bound
}

func prefixIterOptions(prefix []byte) *pebble.IterOptions {
	return &pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: keyUpperBound(prefix),
	}
}

// Iterate creates and returns new PeerIterator.
func (s PeerStorage) Iterate(ctx context.Context) (storage.PeerIterator, error) {
	snap := s.pebble.NewSnapshot()
	iter, err := snap.NewIter(prefixIterOptions(storage.PeerKeyPrefix))
	if err != nil {
		_ = snap.Close()
		return nil, errors.Wrap(err, "new iter")
	}
	iter.First()

	return &pebbleIterator{
		snap: snap,
		iter: iter,
	}, nil
}

func addToBatch(b *pebble.Batch, associated []string, value storage.Peer) error {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "marshal")
	}
	id := storage.KeyFromPeer(value).Bytes(nil)

	set := b.SetDeferred(len(id), len(data))
	copy(set.Key, id)
	copy(set.Value, data)
	_ = set.Finish()

	for _, key := range associated {
		deferred := b.SetDeferred(len(key), len(id))
		copy(deferred.Key, key)
		copy(deferred.Value, id)
		_ = deferred.Finish()
	}

	return nil
}

func (s PeerStorage) commit(fill func(b *pebble.Batch) error) (rerr error) {
	b := s.pebble.NewBatch()
	defer func() {
		multierr.AppendInto(&rerr, b.Close())
	}()

	if err := fill(b); err != nil {
		return err
	}
	if b.Empty() {
		return nil
	}
	if err := b.Commit(s.writeOpts); err != nil {
		return errors.Wrap(err, "commit")
	}

	return nil
}

func (s PeerStorage) add(associated []string, value storage.Peer) error {
	return s.commit(func(b *pebble.Batch) error {
		return addToBatch(b, associated, value)
	})
}

// Add adds given peer to the storage.
func (s PeerStorage) Add(ctx context.Context, value storage.Peer) (rerr error) {
	return s.add(value.Keys(), value)
}

// AddMany adds given peers to the storage in one batch.
//
// A peer that cannot be encoded is skipped: the others are still written,
// and the encoding errors are returned.
func (s PeerStorage) AddMany(ctx context.Context, values []storage.Peer) (rerr error) {
	if err := s.commit(func(b *pebble.Batch) error {
		for _, value := range values {
			if err := addToBatch(b, value.Keys(), value); err != nil {
				multierr.AppendInto(&rerr, errors.Wrapf(err, "add %s", value))
			}
		}
		return nil
	}); err != nil {
		multierr.AppendInto(&rerr, err)
	}
	return rerr
}

// Find finds peer using given key.
func (s PeerStorage) Find(ctx context.Context, key storage.PeerKey) (_ storage.Peer, rerr error) {
	id := key.Bytes(nil)

	data, closer, err := s.pebble.Get(id)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, errors.Wrapf(err, "get %q", id)
	}
	defer func() {
		multierr.AppendInto(&rerr, closer.Close())
	}()

	var b storage.Peer
	if err := json.Unmarshal(data, &b); err != nil {
		if errors.Is(err, storage.ErrPeerUnmarshalMustInvalidate) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, errors.Wrap(err, "unmarshal")
	}

	return b, nil
}

// Assign adds given peer to the storage and associate it to the given key.
func (s PeerStorage) Assign(ctx context.Context, key string, value storage.Peer) (rerr error) {
	return s.add(append(value.Keys(), key), value)
}

// Resolve finds peer using associated key.
func (s PeerStorage) Resolve(ctx context.Context, key string) (_ storage.Peer, rerr error) {
	// Create database snapshot.
	snap := s.pebble.NewSnapshot()
	defer func() {
		multierr.AppendInto(&rerr, snap.Close())
	}()

	// Find id by key.
	id, idCloser, err := snap.Get([]byte(key))
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, errors.Wrapf(err, "get %q", key)
	}
	defer func() {
		multierr.AppendInto(&rerr, idCloser.Close())
	}()

	// Find object by id.
	data, dataCloser, err := snap.Get(id)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, errors.Wrapf(err, "get %q", id)
	}
	defer func() {
		multierr.AppendInto(&rerr, dataCloser.Close())
	}()

	var b storage.Peer
	if err := json.Unmarshal(data, &b); err != nil {
		if errors.Is(err, storage.ErrPeerUnmarshalMustInvalidate) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, errors.Wrap(err, "unmarshal")
	}

	return b, nil
}
