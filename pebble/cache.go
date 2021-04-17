package pebble

import (
	"context"
	"encoding/json"

	"github.com/cockroachdb/pebble"
	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/internal/proto"
)

// ResolverCache is a resolver cache.
type ResolverCache struct {
	next   peer.Resolver
	pebble *pebble.DB
}

// NewResolverCache creates new resolver cache using pebble.
func NewResolverCache(next peer.Resolver, db *pebble.DB) *ResolverCache {
	return &ResolverCache{next: next, pebble: db}
}

// Evict deletes record from cache.
func (r ResolverCache) Evict(ctx context.Context, s string) error {
	return r.pebble.Delete(s2b(s), nil)
}

func (r ResolverCache) notFound(
	ctx context.Context,
	key string, k []byte,
	f func(context.Context, string) (tg.InputPeerClass, error),
) (tg.InputPeerClass, error) {
	// If key not found, try to resolve.
	resolved, err := f(ctx, key)
	if err != nil {
		return nil, err
	}

	// Create proto.Peer object.
	value, err := proto.FromInputPeer(resolved)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	if err := r.pebble.Set(k, data, nil); err != nil {
		return nil, err
	}

	return resolved, nil
}

func (r ResolverCache) tryResolve(
	ctx context.Context,
	key string,
	f func(context.Context, string) (tg.InputPeerClass, error),
) (_ tg.InputPeerClass, rerr error) {
	// Convert key string to the byte slice.
	// Pebble copies key, so we can use unsafe conversion here.
	k := s2b(key)

	data, closer, err := r.pebble.Get(k)
	if xerrors.Is(err, pebble.ErrNotFound) {
		return r.notFound(ctx, key, k, f)
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		multierr.AppendInto(&rerr, closer.Close())
	}()

	var b proto.Peer
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}

	return b.AsInputPeer(), nil
}

// ResolveDomain resolves given domain using cache, if value not found, it uses next resolver in chain.
func (r ResolverCache) ResolveDomain(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	return r.tryResolve(ctx, domain, r.next.ResolveDomain)
}

// ResolvePhone resolves given phone using cache, if value not found, it uses next resolver in chain.
func (r ResolverCache) ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error) {
	return r.tryResolve(ctx, phone, r.next.ResolvePhone)
}
