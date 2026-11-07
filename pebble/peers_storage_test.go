package pebble

import (
	"context"
	"testing"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/peers"

	"github.com/gotd/contrib/internal/tests"
)

func TestPeersStorage(t *testing.T) {
	tests.TestPeersStorage(t, NewPeersStorage(openMem(t)).WithWriteOptions(pebbledb.NoSync))
}

func TestPeersStorageSharedDB(t *testing.T) {
	// PeerStorage iterates over keys with storage.PeerKeyPrefix, so it must
	// not see keys of PeersStorage in the same db.
	ctx := context.Background()
	db := openMem(t)
	s := NewPeersStorage(db).WithWriteOptions(pebbledb.NoSync)
	require.NoError(t, s.Save(ctx, peers.Key{Prefix: "users_", ID: 1}, peers.Value{AccessHash: 1}))
	require.NoError(t, s.SavePhone(ctx, "phone", peers.Key{Prefix: "users_", ID: 1}))
	require.NoError(t, s.SaveContactsHash(ctx, 1))

	iter, err := NewPeerStorage(db).Iterate(ctx)
	require.NoError(t, err)
	defer func() { require.NoError(t, iter.Close()) }()
	require.False(t, iter.Next(ctx))
	require.NoError(t, iter.Err())
}

func TestPeersStorageCorruptValues(t *testing.T) {
	ctx := context.Background()
	db := openMem(t)
	key := peers.Key{Prefix: "users_", ID: 1}
	require.NoError(t, db.Set(encodePeersKey([]byte(prefixPeersKey), key), []byte{1}, pebbledb.NoSync))
	require.NoError(t, db.Set([]byte(prefixPeersPhone+"phone"), []byte{1}, pebbledb.NoSync))
	require.NoError(t, db.Set([]byte(keyPeersContactHash), []byte{1}, pebbledb.NoSync))

	s := NewPeersStorage(db)
	_, _, err := s.Find(ctx, key)
	require.Error(t, err)
	_, _, _, err = s.FindPhone(ctx, "phone") // nolint:dogsled
	require.Error(t, err)
	_, err = s.GetContactsHash(ctx)
	require.Error(t, err)
}
