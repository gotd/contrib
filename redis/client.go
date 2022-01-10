package redis

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/go-redis/redis/v8"

	"github.com/gotd/contrib/auth/kv"
)

type redisClient struct {
	client *redis.Client
}

func (r redisClient) Set(ctx context.Context, k, v string) error {
	return r.client.Set(ctx, k, v, 0).Err()
}

func (r redisClient) Get(ctx context.Context, k string) (string, error) {
	v, err := r.client.Get(ctx, k).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", kv.ErrKeyNotFound
		}
		return "", err
	}

	return v, nil
}
