package bbolt

import (
	"context"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	bboltdb "go.etcd.io/bbolt"

	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/storage"
)

func TestPeerStorageSkipUnchanged(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	db, err := bboltdb.Open(path.Join(t.TempDir(), "bbolt.db"), 0666, &bboltdb.Options{}) // nolint:gocritic
	a.NoError(err)
	t.Cleanup(func() { a.NoError(db.Close()) })
	s := NewPeerStorage(db, []byte("peers"))

	peer := func(username string) storage.Peer {
		var p storage.Peer
		a.True(p.FromUser(&tg.User{ID: 1, AccessHash: 1, Username: username}))
		return p
	}

	a.NoError(s.Add(ctx, peer("foo")))
	writes := dbWrites(db)
	a.NoError(s.Add(ctx, peer("foo")))
	a.Equal(writes, dbWrites(db), "unchanged peer must not be written")

	a.NoError(s.Add(ctx, peer("bar")))
	a.Greater(dbWrites(db), writes)
	p, err := s.Resolve(ctx, "bar")
	a.NoError(err)
	a.Equal(int64(1), p.Key.ID)
}

// dbWrites returns the number of writes performed by bbolt.
func dbWrites(db *bboltdb.DB) int64 {
	stats := db.Stats()
	return stats.TxStats.GetWrite()
}
