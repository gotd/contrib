package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/gotd/contrib/auth/kv"
)

// Credentials stores user credentials to Pebble.
type Credentials struct {
	kv.Credentials
}

// NewCredentials creates new Credentials.
func NewCredentials(client *clientv3.Client) Credentials {
	s := etcdClient{client: client}
	return Credentials{
		Credentials: kv.NewCredentials(s),
	}
}
