package redis

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"go.uber.org/multierr"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/internal/bytesconv"
	"github.com/gotd/contrib/internal/proto"
)

// ResolverCache is a peer resolver cache.
type ResolverCache struct {
	next  peer.Resolver
	redis *redis.Client
}

// NewResolverCache creates new resolver cache using redis.
func NewResolverCache(next peer.Resolver, client *redis.Client) *ResolverCache {
	return &ResolverCache{next: next, redis: client}
}

func (r ResolverCache) notFound(
	ctx context.Context,
	key string,
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

	id := proto.KeyFromPeer(value).Bytes(nil)
	tx := r.redis.TxPipeline()
	defer func() {
		multierr.AppendInto(&rerr, tx.Close())
	}()

	if err := tx.Set(ctx, bytesconv.B2S(id), data, 0).Err(); err != nil {
		return nil, xerrors.Errorf("set id <-> data: %w", err)
	}

	if err := tx.Set(ctx, key, id, 0).Err(); err != nil {
		return nil, xerrors.Errorf("set key <-> id: %w", err)
	}

	if _, err := tx.Exec(ctx); err != nil {
		return nil, xerrors.Errorf("exec: %w", err)
	}

	return resolved, nil
}

func (r ResolverCache) tryResolve(
	ctx context.Context,
	key string,
	f func(context.Context, string) (tg.InputPeerClass, error),
) (tg.InputPeerClass, error) {
	// Find id by domain.
	id, err := r.redis.Get(ctx, key).Result()
	if xerrors.Is(err, redis.Nil) {
		return r.notFound(ctx, key, f)
	}
	if err != nil {
		return nil, xerrors.Errorf("get %q: %w", key, err)
	}

	// Find object by id.
	data, err := r.redis.Get(ctx, id).Bytes()
	if err != nil {
		return nil, xerrors.Errorf("get %q: %w", id, err)
	}

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
