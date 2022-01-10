package redis

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-faster/errors"
	"github.com/go-redis/redis/v8"
	"go.uber.org/multierr"

	"github.com/gotd/contrib/storage"
)

var _ storage.PeerStorage = PeerStorage{}

// PeerStorage is a peer storage based on redis.
type PeerStorage struct {
	redis *redis.Client
}

// NewPeerStorage creates new peer storage using redis.
func NewPeerStorage(client *redis.Client) *PeerStorage {
	return &PeerStorage{redis: client}
}

type redisIterator struct {
	client  *redis.Client
	iter    *redis.ScanIterator
	lastErr error
	value   storage.Peer
}

func (p *redisIterator) Close() error {
	return nil
}

func (p *redisIterator) Next(ctx context.Context) bool {
	if !p.iter.Next(ctx) {
		return false
	}

	key := p.iter.Val()
	value, err := p.client.Get(ctx, key).Result()
	if err != nil {
		p.lastErr = errors.Errorf("get %q: %w", key, err)
		return false
	}

	r := strings.NewReader(value)
	if err := json.NewDecoder(r).Decode(&p.value); err != nil {
		p.lastErr = errors.Errorf("unmarshal: %w", err)
		return false
	}

	return true
}

func (p *redisIterator) Err() error {
	return multierr.Append(p.lastErr, p.iter.Err())
}

func (p *redisIterator) Value() storage.Peer {
	return p.value
}

// Iterate creates and returns new PeerIterator.
func (s PeerStorage) Iterate(ctx context.Context) (storage.PeerIterator, error) {
	var b strings.Builder
	b.Grow(len(storage.PeerKeyPrefix) + 1)
	b.Write(storage.PeerKeyPrefix)
	b.WriteByte('*')

	result := s.redis.Scan(ctx, 0, b.String(), 0)
	return &redisIterator{
		client: s.redis,
		iter:   result.Iterator(),
	}, result.Err()
}

func (s PeerStorage) add(ctx context.Context, associated []string, value storage.Peer) (rerr error) {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.Errorf("marshal: %w", err)
	}
	id := storage.KeyFromPeer(value).String()

	if len(associated) == 0 {
		if err := s.redis.Set(ctx, id, data, 0).Err(); err != nil {
			return errors.Errorf("set id <-> data: %w", err)
		}

		return nil
	}

	tx := s.redis.TxPipeline()
	defer func() {
		multierr.AppendInto(&rerr, tx.Close())
	}()

	if err := tx.Set(ctx, id, data, 0).Err(); err != nil {
		return errors.Errorf("set id <-> data: %w", err)
	}

	for _, key := range associated {
		if err := tx.Set(ctx, key, id, 0).Err(); err != nil {
			return errors.Errorf("set key <-> id: %w", err)
		}
	}

	if _, err := tx.Exec(ctx); err != nil {
		return errors.Errorf("exec: %w", err)
	}

	return nil
}

// Add adds given peer to the storage.
func (s PeerStorage) Add(ctx context.Context, value storage.Peer) error {
	return s.add(ctx, value.Keys(), value)
}

// Find finds peer using given key.
func (s PeerStorage) Find(ctx context.Context, key storage.PeerKey) (storage.Peer, error) {
	id := key.String()

	data, err := s.redis.Get(ctx, id).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, errors.Errorf("get %q: %w", key, err)
	}

	var b storage.Peer
	if err := json.Unmarshal(data, &b); err != nil {
		return storage.Peer{}, errors.Errorf("unmarshal: %w", err)
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
		if errors.Is(err, redis.Nil) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, errors.Errorf("get %q: %w", key, err)
	}

	// Find object by id.
	data, err := s.redis.Get(ctx, id).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return storage.Peer{}, storage.ErrPeerNotFound
		}
		return storage.Peer{}, errors.Errorf("get %q: %w", id, err)
	}

	var b storage.Peer
	if err := json.Unmarshal(data, &b); err != nil {
		return storage.Peer{}, errors.Errorf("unmarshal: %w", err)
	}

	return b, nil
}
