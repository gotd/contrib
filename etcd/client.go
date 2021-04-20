package etcd

import (
	"context"

	"go.etcd.io/etcd/client/v3"

	"github.com/gotd/contrib/auth/kv"
)

type etcdClient struct {
	client *clientv3.Client
}

func (r etcdClient) Set(ctx context.Context, k, v string) error {
	_, err := r.client.Put(ctx, k, v)
	if err != nil {
		return err
	}
	return nil
}

func (r etcdClient) Get(ctx context.Context, k string) (string, error) {
	resp, err := r.client.Get(ctx, k)
	if err != nil {
		return "", err
	}

	if resp.Count < 1 || len(resp.Kvs) < 1 {
		return "", kv.ErrKeyNotFound
	}

	return string(resp.Kvs[0].Value), nil
}
