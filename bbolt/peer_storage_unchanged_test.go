package bbolt

import (
	"context"
	"encoding/json"
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

func TestPeerStorageEncodingFormats(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	db, err := bboltdb.Open(path.Join(t.TempDir(), "bbolt.db"), 0666, &bboltdb.Options{}) // nolint:gocritic
	a.NoError(err)
	t.Cleanup(func() { a.NoError(db.Close()) })
	bucket := []byte("peers")
	s := NewPeerStorage(db, bucket)

	var legacy storage.Peer
	a.True(legacy.FromUser(&tg.User{ID: 1, AccessHash: 1}))
	data, err := json.Marshal(legacy)
	a.NoError(err)
	a.NoError(db.Update(func(tx *bboltdb.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return b.Put(storage.KeyFromPeer(legacy).Bytes(nil), data)
	}))
	p, err := s.Find(ctx, storage.KeyFromPeer(legacy))
	a.NoError(err, "old JSON values must be readable")
	a.Equal(int64(1), p.Key.AccessHash)

	firstByte := func(p storage.Peer) (b byte) {
		a.NoError(db.View(func(tx *bboltdb.Tx) error {
			b = tx.Bucket(bucket).Get(storage.KeyFromPeer(p).Bytes(nil))[0]
			return nil
		}))
		return b
	}

	// JSON stays the default, so older versions can read new values.
	var fresh storage.Peer
	a.True(fresh.FromUser(&tg.User{ID: 2, AccessHash: 2}))
	a.NoError(s.Add(ctx, fresh))
	a.Equal(byte('{'), firstByte(fresh), "default must be JSON")

	// The compact format is opt-in, and both formats stay readable.
	bin := NewPeerStorage(db, bucket).WithBinaryEncoding()
	var compact storage.Peer
	a.True(compact.FromUser(&tg.User{ID: 3, AccessHash: 3}))
	a.NoError(bin.Add(ctx, compact))
	a.NotEqual(byte('{'), firstByte(compact), "opt-in values must not be JSON")
	for _, st := range []*PeerStorage{s, bin} {
		for _, want := range []storage.Peer{fresh, compact} {
			p, err := st.Find(ctx, storage.KeyFromPeer(want))
			a.NoError(err)
			a.Equal(want.Key.AccessHash, p.Key.AccessHash)
		}
	}
}
