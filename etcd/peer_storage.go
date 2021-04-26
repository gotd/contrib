package etcd

import (
	"context"
	"encoding/json"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/xerrors"

	"github.com/gotd/contrib/storage"
)

var _ storage.PeerStorage = PeerStorage{}

// PeerStorage is a peer storage based on etcd.
type PeerStorage struct {
	etcd      *clientv3.Client
	iterLimit int64
}

// NewPeerStorage creates new peer storage using etcd.
func NewPeerStorage(etcd *clientv3.Client) *PeerStorage {
	return &PeerStorage{etcd: etcd, iterLimit: 25}
}

// WithIterLimit sets limit of buffer for used for iteration.
func (s *PeerStorage) WithIterLimit(iterLimit int64) *PeerStorage {
	s.iterLimit = iterLimit
	return s
}

type etcdIterator struct {
	etcd      *clientv3.Client
	iterLimit int64

	// Iterator state.
	lastKey string
	lastErr error
	wasLast bool

	// Buffer state.
	cursor int
	buf    []storage.Peer
}

func (p *etcdIterator) Close() error {
	return nil
}

func (p *etcdIterator) Next(ctx context.Context) bool {
	switch {
	case p.cursor < len(p.buf):
		p.cursor++
		return true
	case p.wasLast:
		return false
	}

	r, err := p.etcd.Get(ctx, p.lastKey,
		clientv3.WithFromKey(),
		clientv3.WithPrefix(),
		clientv3.WithLimit(p.iterLimit),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
	)
	if err != nil {
		p.lastErr = xerrors.Errorf("get from %q: %w", p.lastKey, err)
		return false
	}
	p.wasLast = !r.More
	if r.Count < 1 || len(r.Kvs) < 1 {
		return false
	}

	for _, pair := range r.Kvs {
		var value storage.Peer
		if err := json.Unmarshal(pair.Value, &value); err != nil {
			p.lastErr = xerrors.Errorf("unmarshal: %w", err)
			return false
		}

		p.buf = append(p.buf, value)
	}
	p.lastKey = string(r.Kvs[len(r.Kvs)-1].Key)

	return true
}

func (p *etcdIterator) Err() error {
	return p.lastErr
}

func (p *etcdIterator) Value() storage.Peer {
	return p.buf[p.cursor-1]
}

// Iterate creates and returns new PeerIterator
func (s PeerStorage) Iterate(ctx context.Context) (storage.PeerIterator, error) {
	return &etcdIterator{
		etcd:      s.etcd,
		iterLimit: s.iterLimit,
		lastKey:   string(storage.KeyPrefix),
		buf:       make([]storage.Peer, 0, s.iterLimit),
	}, nil
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

	ops := make([]clientv3.Op, 0, len(associated)+1)
	ops = append(ops, clientv3.OpPut(id, buf.String()))
	for _, key := range associated {
		ops = append(ops, clientv3.OpPut(key, id))
	}

	if _, err := s.etcd.Txn(ctx).Then(ops...).Commit(); err != nil {
		return xerrors.Errorf("commit: %w", err)
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
