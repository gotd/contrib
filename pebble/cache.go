package pebble

import (
	"context"
	"encoding/json"

	"github.com/cockroachdb/pebble"
	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/internal/bytesconv"
	"github.com/gotd/contrib/internal/proto"
)

// ResolverCache is a peer resolver cache.
type ResolverCache struct {
	next   peer.Resolver
	pebble *pebble.DB
}

// NewResolverCache creates new resolver cache using pebble.
func NewResolverCache(next peer.Resolver, db *pebble.DB) *ResolverCache {
	return &ResolverCache{next: next, pebble: db}
}

func (r ResolverCache) notFound(
	ctx context.Context,
	key string, k []byte,
	f func(context.Context, string) (tg.InputPeerClass, error),
) (_ tg.InputPeerClass, rerr error) {
	// If key not found, try to resolve.
	resolved, err := f(ctx, key)
	if err != nil {
		return nil, err
	}

	// Create proto.Peer object.
	value, err := proto.FromInputPeer(resolved)
	if err != nil {
		return nil, xerrors.Errorf("extract object: %w", err)
	}

	data, err := json.Marshal(value)
	if err != nil {
		return nil, xerrors.Errorf("marshal: %w", err)
	}

	b := r.pebble.NewBatch()
	defer func() {
		multierr.AppendInto(&rerr, b.Close())
	}()

	id := proto.KeyFromPeer(value).Bytes(nil)
	if err := b.Set(id, data, nil); err != nil {
		return nil, xerrors.Errorf("set id <-> data: %w", err)
	}

	if err := b.Set(k, id, nil); err != nil {
		return nil, xerrors.Errorf("set key <-> id: %w", err)
	}

	if err := b.Commit(nil); err != nil {
		return nil, xerrors.Errorf("commit changes: %w", err)
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
	k := bytesconv.S2B(key)

	// Find id by domain.
	id, idCloser, err := r.pebble.Get(k)
	if xerrors.Is(err, pebble.ErrNotFound) {
		return r.notFound(ctx, key, k, f)
	}
	if err != nil {
		return nil, xerrors.Errorf("get %q: %w", key, err)
	}
	defer func() {
		multierr.AppendInto(&rerr, idCloser.Close())
	}()

	// Find object by id.
	data, dataCloser, err := r.pebble.Get(id)
	if err != nil {
		return nil, xerrors.Errorf("get %q: %w", id, err)
	}
	defer func() {
		multierr.AppendInto(&rerr, dataCloser.Close())
	}()

	var b proto.Peer
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, xerrors.Errorf("unmarshal: %w", err)
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