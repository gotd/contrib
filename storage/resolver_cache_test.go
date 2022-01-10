package storage

import (
	"context"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

type memStorage struct {
	peers map[PeerKey]Peer
	keys  map[string]PeerKey
}

func newMemStorage() memStorage {
	return memStorage{
		peers: map[PeerKey]Peer{},
		keys:  map[string]PeerKey{},
	}
}

func (m memStorage) Iterate(ctx context.Context) (PeerIterator, error) {
	return nil, errors.New("unimplemented")
}

func (m memStorage) add(keys []string, p Peer) {
	id := KeyFromPeer(p)
	m.peers[id] = p
	for _, key := range keys {
		m.keys[key] = id
	}
}

func (m memStorage) Add(ctx context.Context, p Peer) error {
	m.add(p.Keys(), p)
	return nil
}

func (m memStorage) Find(ctx context.Context, key PeerKey) (Peer, error) {
	v, ok := m.peers[key]
	if !ok {
		return Peer{}, ErrPeerNotFound
	}
	return v, nil
}

func (m memStorage) Assign(ctx context.Context, key string, p Peer) error {
	m.add(append(p.Keys(), key), p)
	return nil
}

func (m memStorage) Resolve(ctx context.Context, key string) (Peer, error) {
	id, ok := m.keys[key]
	if !ok {
		return Peer{}, ErrPeerNotFound
	}

	v, ok := m.peers[id]
	if !ok {
		return Peer{}, ErrPeerNotFound
	}
	return v, nil
}

type resolverFunc func(ctx context.Context, domain string) (tg.InputPeerClass, error)

func (r resolverFunc) ResolveDomain(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	return r(ctx, domain)
}

func (r resolverFunc) ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error) {
	return r(ctx, phone)
}

func TestResolverCache(t *testing.T) {
	t.Run("Domain", func(t *testing.T) {
		a := require.New(t)
		ctx := context.Background()
		expected := &tg.InputPeerUser{
			UserID:     10,
			AccessHash: 10,
		}
		expectedKey := "abc"
		counter := 0

		r := func(ctx context.Context, k string) (tg.InputPeerClass, error) {
			a.Equal(expectedKey, k)
			a.Zero(counter)
			counter++
			return expected, nil
		}
		c := NewResolverCache(resolverFunc(r), newMemStorage())

		result, err := c.ResolveDomain(ctx, "abc")
		a.NoError(err)
		a.Equal(expected, result)

		result, err = c.ResolveDomain(ctx, "abc")
		a.NoError(err)
		a.Equal(expected, result)
	})

	t.Run("Phone", func(t *testing.T) {
		a := require.New(t)
		ctx := context.Background()
		expected := &tg.InputPeerUser{
			UserID:     10,
			AccessHash: 10,
		}
		expectedKey := "abc"
		counter := 0

		r := func(ctx context.Context, k string) (tg.InputPeerClass, error) {
			a.Equal(expectedKey, k)
			a.Zero(counter)
			counter++
			return expected, nil
		}
		c := NewResolverCache(resolverFunc(r), newMemStorage())

		result, err := c.ResolvePhone(ctx, "abc")
		a.NoError(err)
		a.Equal(expected, result)

		result, err = c.ResolvePhone(ctx, "abc")
		a.NoError(err)
		a.Equal(expected, result)
	})
}
