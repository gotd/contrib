package pebble

import (
	"context"
	"encoding/json"

	"github.com/cockroachdb/pebble"
	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/gotd/contrib/storage"
)

// PeerStorage is a peer storage based on pebble.
type PeerStorage struct {
	pebble *pebble.DB
}

// NewPeerStorage creates new peer storage using pebble.
func NewPeerStorage(db *pebble.DB) *PeerStorage {
	return &PeerStorage{pebble: db}
}

func (s PeerStorage) add(associated []string, value storage.Peer) (rerr error) {
	data, err := json.Marshal(value)
	if err != nil {
		return xerrors.Errorf("marshal: %w", err)
	}
	id := storage.KeyFromPeer(value).Bytes(nil)

	b := s.pebble.NewBatch()
	defer func() {
		multierr.AppendInto(&rerr, b.Close())
	}()

	set := b.SetDeferred(len(id), len(data))
	copy(set.Key, id)
	copy(set.Value, data)
	set.Finish()

	for _, key := range associated {
		deferred := b.SetDeferred(len(key), len(id))
		copy(deferred.Key, key)
		copy(deferred.Value, id)
		deferred.Finish()
	}

	if err := b.Commit(nil); err != nil {
		return xerrors.Errorf("commit: %w", err)
	}

	return nil
}

// Add adds given peer to the storage.
func (s PeerStorage) Add(ctx context.Context, value storage.Peer) (rerr error) {
	return s.add(value.Keys(), value)
}

// Find finds peer using given key.
func (s PeerStorage) Find(ctx context.Context, key storage.Key) (_ storage.Peer, rerr error) {
	id := key.Bytes(nil)

	data, closer, err := s.pebble.Get(id)
	if err != nil {
		if xerrors.Is(err, pebble.ErrNotFound) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, xerrors.Errorf("get %q: %w", id, err)
	}
	defer func() {
		multierr.AppendInto(&rerr, closer.Close())
	}()

	var b storage.Peer
	if err := json.Unmarshal(data, &b); err != nil {
		return storage.Peer{}, xerrors.Errorf("unmarshal: %w", err)
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
		if xerrors.Is(err, pebble.ErrNotFound) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, xerrors.Errorf("get %q: %w", key, err)
	}
	defer func() {
		multierr.AppendInto(&rerr, idCloser.Close())
	}()

	// Find object by id.
	data, dataCloser, err := snap.Get(id)
	if err != nil {
		if xerrors.Is(err, pebble.ErrNotFound) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, xerrors.Errorf("get %q: %w", id, err)
	}
	defer func() {
		multierr.AppendInto(&rerr, dataCloser.Close())
	}()

	var b storage.Peer
	if err := json.Unmarshal(data, &b); err != nil {
		return storage.Peer{}, xerrors.Errorf("unmarshal: %w", err)
	}

	return b, nil
}
