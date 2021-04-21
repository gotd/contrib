package redis

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/gotd/contrib/storage"
)

// PeerStorage is a peer storage based on redis.
type PeerStorage struct {
	redis *redis.Client
}

// NewPeerStorage creates new peer storage using redis.
func NewPeerStorage(client *redis.Client) *PeerStorage {
	return &PeerStorage{redis: client}
}

func (s PeerStorage) add(ctx context.Context, associated []string, value storage.Peer) (rerr error) {
	data, err := json.Marshal(value)
	if err != nil {
		return xerrors.Errorf("marshal: %w", err)
	}
	id := storage.KeyFromPeer(value).String()

	if len(associated) == 0 {
		if err := s.redis.Set(ctx, id, data, 0).Err(); err != nil {
			return xerrors.Errorf("set id <-> data: %w", err)
		}

		return nil
	}

	tx := s.redis.TxPipeline()
	defer func() {
		multierr.AppendInto(&rerr, tx.Close())
	}()

	if err := tx.Set(ctx, id, data, 0).Err(); err != nil {
		return xerrors.Errorf("set id <-> data: %w", err)
	}

	for _, key := range associated {
		if err := tx.Set(ctx, key, id, 0).Err(); err != nil {
			return xerrors.Errorf("set key <-> id: %w", err)
		}
	}

	if _, err := tx.Exec(ctx); err != nil {
		return xerrors.Errorf("exec: %w", err)
	}

	return nil
}

// Add adds given peer to the storage.
func (s PeerStorage) Add(ctx context.Context, value storage.Peer) error {
	return s.add(ctx, value.Keys(), value)
}

// Find finds peer using given key.
func (s PeerStorage) Find(ctx context.Context, key storage.Key) (storage.Peer, error) {
	id := key.String()

	data, err := s.redis.Get(ctx, id).Bytes()
	if err != nil {
		if xerrors.Is(err, redis.Nil) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, xerrors.Errorf("get %q: %w", key, err)
	}

	var b storage.Peer
	if err := json.Unmarshal(data, &b); err != nil {
		return storage.Peer{}, xerrors.Errorf("unmarshal: %w", err)
	}

	return b, nil
}

// Assign adds given peer to the storage and associate it to the given key.
func (s PeerStorage) Assign(ctx context.Context, key string, value storage.Peer) (rerr error) {
	return s.add(ctx, append(value.Keys(), key), value)
}

// Resolve finds peer using associated key.
func (s PeerStorage) Resolve(ctx context.Context, key string) (storage.Peer, error) {
	// Find id by domain.
	id, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if xerrors.Is(err, redis.Nil) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, xerrors.Errorf("get %q: %w", key, err)
	}

	// Find object by id.
	data, err := s.redis.Get(ctx, id).Bytes()
	if err != nil {
		if xerrors.Is(err, redis.Nil) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, xerrors.Errorf("get %q: %w", id, err)
	}

	var b storage.Peer
	if err := json.Unmarshal(data, &b); err != nil {
		return storage.Peer{}, xerrors.Errorf("unmarshal: %w", err)
	}

	return b, nil
}
