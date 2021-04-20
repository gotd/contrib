package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/gotd/td/session"

	"github.com/gotd/contrib/auth/kv"
)

var _ session.Storage = SessionStorage{}

// SessionStorage is a MTProto session Redis storage.
type SessionStorage struct {
	kv.Session
}

// NewSessionStorage creates new SessionStorage.
func NewSessionStorage(client *clientv3.Client, key string) SessionStorage {
	s := etcdClient{client: client}
	return SessionStorage{
		Session: kv.NewSession(s, key),
	}
}
