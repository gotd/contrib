package etcd

import (
	"context"
	"encoding/json"
	"strings"

	"go.etcd.io/etcd/client/v3"
	"golang.org/x/xerrors"

	"github.com/gotd/contrib/internal/bytesconv"
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

// Add adds given peer to the storage.
func (s PeerStorage) Add(ctx context.Context, value storage.Peer) error {
	data, err := json.Marshal(value)
	if err != nil {
		return xerrors.Errorf("marshal: %w", err)
	}

	id := storage.KeyFromPeer(value).Bytes(nil)
	if _, err := s.etcd.Put(ctx, bytesconv.B2S(id), bytesconv.B2S(data)); err != nil {
		return xerrors.Errorf("set id <-> data: %w", err)
	}

	return nil
}

// Find finds peer using given key.
func (s PeerStorage) Find(ctx context.Context, key storage.Key) (storage.Peer, error) {
	id := bytesconv.B2S(key.Bytes(nil))

	resp, err := s.etcd.Get(ctx, id)
	if err != nil {
		return storage.Peer{}, xerrors.Errorf("get %q: %w", key, err)
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
	var b strings.Builder
	if err := json.NewEncoder(&b).Encode(value); err != nil {
		return xerrors.Errorf("marshal: %w", err)
	}
	id := bytesconv.B2S(storage.KeyFromPeer(value).Bytes(nil))

	_, err := s.etcd.Txn(ctx).Then(
		clientv3.OpPut(key, id),
		clientv3.OpPut(id, b.String()),
	).Commit()
	if err != nil {
		return xerrors.Errorf("commit: %w", err)
	}

	return nil
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

	id := bytesconv.B2S(resp.Kvs[0].Value)
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
