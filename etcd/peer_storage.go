package etcd

import (
	"context"
	"encoding/json"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/xerrors"

	"github.com/gotd/contrib/storage"
)

// PeerStorage is a peer storage based on etcd.
type PeerStorage struct {
	etcd *clientv3.Client
}

// NewPeerStorage creates new peer storage using etcd.
func NewPeerStorage(etcd *clientv3.Client) *PeerStorage {
	return &PeerStorage{etcd: etcd}
}

func (s PeerStorage) add(ctx context.Context, associated []string, value storage.Peer) (rerr error) {
	var buf strings.Builder
	if err := json.NewEncoder(&buf).Encode(value); err != nil {
		return xerrors.Errorf("marshal: %w", err)
	}
	id := storage.KeyFromPeer(value).String()

	if len(associated) == 0 {
		if _, err := s.etcd.Put(ctx, id, buf.String()); err != nil {
			return xerrors.Errorf("set id <-> data: %w", err)
		}

		return nil
	}

	tx := s.etcd.Txn(ctx).Then(
		clientv3.OpPut(id, buf.String()),
	)

	for _, key := range associated {
		tx.Then(clientv3.OpPut(key, id))
	}

	if _, err := tx.Commit(); err != nil {
		return xerrors.Errorf("commit: %w", err)
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

	resp, err := s.etcd.Get(ctx, id)
	if err != nil {
		return storage.Peer{}, xerrors.Errorf("get %q: %w", id, err)
	}
	if resp.Count < 1 || len(resp.Kvs) < 1 {
		return storage.Peer{}, storage.ErrPeerNotFound
	}

	var b storage.Peer
	if err := json.Unmarshal(resp.Kvs[0].Value, &b); err != nil {
		return storage.Peer{}, xerrors.Errorf("unmarshal: %w", err)
	}

	return b, nil
}

// Assign adds given peer to the storage and associate it to the given key.
func (s PeerStorage) Assign(ctx context.Context, key string, value storage.Peer) error {
	return s.add(ctx, append(value.Keys(), key), value)
}

// Resolve finds peer using associated key.
func (s PeerStorage) Resolve(ctx context.Context, key string) (storage.Peer, error) {
	resp, err := s.etcd.Get(ctx, key)
	if err != nil {
		return storage.Peer{}, xerrors.Errorf("get %q: %w", key, err)
	}
	if resp.Count < 1 || len(resp.Kvs) < 1 {
		return storage.Peer{}, storage.ErrPeerNotFound
	}

	id := string(resp.Kvs[0].Value)
	resp, err = s.etcd.Get(ctx, id)
	if err != nil {
		return storage.Peer{}, xerrors.Errorf("get %q: %w", id, err)
	}
	if resp.Count < 1 || len(resp.Kvs) < 1 {
		return storage.Peer{}, storage.ErrPeerNotFound
	}

	var b storage.Peer
	if err := json.Unmarshal(resp.Kvs[0].Value, &b); err != nil {
		return storage.Peer{}, xerrors.Errorf("unmarshal: %w", err)
	}

	return b, nil
}
