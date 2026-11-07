package pebble

import (
	"context"
	"testing"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/internal/tests"
	"github.com/gotd/contrib/storage"
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

func TestPeersStorageUsernameCollision(t *testing.T) {
	// PeerStorage writes usernames and phones as raw keys into the same db, so
	// keys of PeersStorage must contain a byte usernames and phones can't.
	ctx := context.Background()
	db := openMem(t)
	s := NewPeersStorage(db).WithWriteOptions(pebbledb.NoSync)
	key := peers.Key{Prefix: "users_", ID: 1}
	require.NoError(t, s.SaveContactsHash(ctx, 42))
	require.NoError(t, s.Save(ctx, key, peers.Value{AccessHash: 1}))
	require.NoError(t, s.SavePhone(ctx, "79991234567", key))

	// Usernames (and the phone) that equal old-style keys of PeersStorage.
	ps := NewPeerStorage(db).WithWriteOptions(pebbledb.NoSync)
	for i, name := range []string{"tdpeers_contacts_hash", "tdpeers_phone_79991234567"} {
		var p storage.Peer
		require.True(t, p.FromUser(&tg.User{ID: int64(10 + i), AccessHash: 1, Username: name}))
		require.NoError(t, ps.Add(ctx, p))
	}

	hash, err := s.GetContactsHash(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(42), hash)
	got, _, found, err := s.FindPhone(ctx, "79991234567")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, key, got)

	for _, k := range []string{prefixPeersKey, prefixPeersPhone, keyPeersContactHash} {
		require.Contains(t, k, ":", "%q must contain a byte usernames can't", k)
	}
}
