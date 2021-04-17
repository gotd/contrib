package redis

import (
	"github.com/go-redis/redis/v8"

	"github.com/gotd/td/session"

	"github.com/gotd/contrib/auth/kv"
)

var _ session.Storage = SessionStorage{}

// SessionStorage is a MTProto session Redis storage.
type SessionStorage struct {
	kv.Session
}

// NewSessionStorage creates new SessionStorage.
func NewSessionStorage(client *redis.Client, key string) SessionStorage {
	s := redisClient{client: client}
	return SessionStorage{
		Session: kv.NewSession(s, key),
	}
}
