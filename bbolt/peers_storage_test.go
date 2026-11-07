package bbolt

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"

	"github.com/gotd/td/telegram/peers"

	"github.com/gotd/contrib/internal/tests"
)

func openTemp(t *testing.T) *bolt.DB {
	t.Helper()
	db, err := bolt.Open(filepath.Join(t.TempDir(), "bbolt.db"), 0o600, &bolt.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestPeersStorage(t *testing.T) {
	tests.TestPeersStorage(t, NewPeersStorage(openTemp(t)))
}

func TestPeersStorageCorruptValues(t *testing.T) {
	ctx := context.Background()
	db := openTemp(t)
	key := peers.Key{Prefix: "users_", ID: 1}
	s := NewPeersStorage(db)
	require.NoError(t, s.put(bucketPeersKeys, encodePeersKey(key), []byte{1}))
	require.NoError(t, s.put(bucketPeersPhones, []byte("phone"), []byte{1}))
	require.NoError(t, s.put(bucketPeersMeta, keyContactsHash, []byte{1}))

	_, _, err := s.Find(ctx, key)
	require.Error(t, err)
	_, _, _, err = s.FindPhone(ctx, "phone") // nolint:dogsled
	require.Error(t, err)
	_, err = s.GetContactsHash(ctx)
	require.Error(t, err)
}
