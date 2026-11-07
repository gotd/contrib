package pebble_test

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/pebble"
	"github.com/gotd/contrib/storage"
)

// syncFS counts file syncs.
type syncFS struct {
	vfs.FS
	syncs *atomic.Int64
}

type syncFile struct {
	vfs.File
	syncs *atomic.Int64
}

func (f syncFile) Sync() error {
	f.syncs.Add(1)
	return f.File.Sync()
}

func (f syncFile) SyncData() error {
	f.syncs.Add(1)
	return f.File.SyncData()
}

func (f syncFile) SyncTo(length int64) (bool, error) {
	f.syncs.Add(1)
	return f.File.SyncTo(length)
}

func (s syncFS) wrap(f vfs.File, err error) (vfs.File, error) {
	if err != nil {
		return nil, err
	}
	return syncFile{File: f, syncs: s.syncs}, nil
}

func (s syncFS) Create(name string) (vfs.File, error) {
	return s.wrap(s.FS.Create(name))
}

func (s syncFS) ReuseForWrite(oldname, newname string) (vfs.File, error) {
	return s.wrap(s.FS.ReuseForWrite(oldname, newname))
}

func (s syncFS) OpenReadWrite(name string, opts ...vfs.OpenOption) (vfs.File, error) {
	return s.wrap(s.FS.OpenReadWrite(name, opts...))
}

func TestPeerStorageSyncs(t *testing.T) {
	ctx := context.Background()
	update := &tg.Updates{}
	for i := 1; i <= 10; i++ {
		update.Users = append(update.Users, &tg.User{ID: int64(i), AccessHash: int64(i)})
	}
	next := telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil })

	for _, tt := range []struct {
		name  string
		opts  *pebbledb.WriteOptions
		syncs int64
	}{
		// One synced commit per update, not per peer.
		{"Default", nil, 1},
		{"Sync", pebbledb.Sync, 1},
		{"NoSync", pebbledb.NoSync, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			a := require.New(t)
			var syncs atomic.Int64
			db, err := pebbledb.Open("pebble.db", &pebbledb.Options{
				FS: syncFS{FS: vfs.NewMem(), syncs: &syncs},
			})
			a.NoError(err)
			defer func() { _ = db.Close() }()

			s := pebble.NewPeerStorage(db).WithWriteOptions(tt.opts)
			h := storage.UpdateHook(next, s)

			before := syncs.Load()
			a.NoError(h.Handle(ctx, update))
			a.Equal(tt.syncs, syncs.Load()-before)

			for i := 1; i <= 10; i++ {
				p, err := s.Find(ctx, storage.PeerKey{Kind: dialogs.User, ID: int64(i)})
				a.NoError(err)
				a.Equal(int64(i), p.Key.AccessHash)
			}
		})
	}
}

func TestPeerStorageAddManyEncodeError(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	db, err := pebbledb.Open("pebble.db", &pebbledb.Options{FS: vfs.NewMem()})
	a.NoError(err)
	defer func() { _ = db.Close() }()
	s := pebble.NewPeerStorage(db)

	var peers []storage.Peer
	for i := 1; i <= 3; i++ {
		var p storage.Peer
		a.True(p.FromUser(&tg.User{ID: int64(i), AccessHash: int64(i)}))
		peers = append(peers, p)
	}
	// Metadata that cannot be encoded fails only this peer.
	peers[1].Metadata = map[string]any{"bad": func() {}}

	err = s.AddMany(ctx, peers)
	a.Error(err, "encoding error must be reported")

	for _, id := range []int64{1, 3} {
		p, err := s.Find(ctx, storage.PeerKey{Kind: dialogs.User, ID: id})
		a.NoError(err, "other peers must be written")
		a.Equal(id, p.Key.AccessHash)
	}
	_, err = s.Find(ctx, storage.PeerKey{Kind: dialogs.User, ID: 2})
	a.ErrorIs(err, storage.ErrPeerNotFound)
}

func TestPeerStorageSkipUnchanged(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	var syncs atomic.Int64
	db, err := pebbledb.Open("pebble.db", &pebbledb.Options{
		FS: syncFS{FS: vfs.NewMem(), syncs: &syncs},
	})
	a.NoError(err)
	defer func() { _ = db.Close() }()

	s := pebble.NewPeerStorage(db)
	next := telegram.UpdateHandlerFunc(func(context.Context, tg.UpdatesClass) error { return nil })
	h := storage.UpdateHook(next, s)
	update := func(username string) *tg.Updates {
		return &tg.Updates{Users: []tg.UserClass{
			&tg.User{ID: 1, AccessHash: 1, Username: username},
			&tg.User{ID: 2, AccessHash: 2},
		}}
	}

	a.NoError(h.Handle(ctx, update("foo")))
	before := syncs.Load()
	// Same peers again (CreatedAt differs): nothing is written.
	a.NoError(h.Handle(ctx, update("foo")))
	a.Equal(before, syncs.Load(), "unchanged peers must not be written")
	a.NoError(s.Add(ctx, mustUser(t, &tg.User{ID: 2, AccessHash: 2})))
	a.Equal(before, syncs.Load(), "unchanged peer must not be written")

	// Changed peer is written.
	a.NoError(h.Handle(ctx, update("bar")))
	a.Greater(syncs.Load(), before)
	p, err := s.Resolve(ctx, "bar")
	a.NoError(err)
	a.Equal(int64(1), p.Key.ID)
}

func mustUser(t *testing.T, u *tg.User) storage.Peer {
	t.Helper()
	var p storage.Peer
	require.True(t, p.FromUser(u))
	return p
}

func TestPeerStorageAddManyLastWins(t *testing.T) {
	// A stored peer repeated after a changed copy in one batch must win, even
	// though it equals the stored value.
	a := require.New(t)
	ctx := context.Background()
	db, err := pebbledb.Open("pebble.db", &pebbledb.Options{FS: vfs.NewMem()})
	a.NoError(err)
	defer func() { _ = db.Close() }()

	s := pebble.NewPeerStorage(db)
	stored := &tg.User{ID: 1, AccessHash: 1, FirstName: "A"}
	a.NoError(s.Add(ctx, mustUser(t, stored)))
	a.NoError(s.AddMany(ctx, []storage.Peer{
		mustUser(t, &tg.User{ID: 1, AccessHash: 1, FirstName: "A'"}),
		mustUser(t, stored),
	}))

	var key storage.PeerKey
	key.Kind = dialogs.User
	key.ID = 1
	p, err := s.Find(ctx, key)
	a.NoError(err)
	a.Equal("A", p.User.FirstName)
}

func TestPeerStorageEncodingFormats(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	db, err := pebbledb.Open("pebble.db", &pebbledb.Options{FS: vfs.NewMem()})
	a.NoError(err)
	defer func() { _ = db.Close() }()
	s := pebble.NewPeerStorage(db)

	// Peer stored in the old JSON format is still readable.
	legacy := mustUser(t, &tg.User{ID: 1, AccessHash: 1})
	data, err := json.Marshal(legacy)
	a.NoError(err)
	id := storage.KeyFromPeer(legacy).Bytes(nil)
	a.NoError(db.Set(id, data, pebbledb.Sync))
	p, err := s.Find(ctx, storage.KeyFromPeer(legacy))
	a.NoError(err)
	a.Equal(int64(1), p.Key.AccessHash)

	// New writes use the compact format.
	fresh := mustUser(t, &tg.User{ID: 2, AccessHash: 2})
	a.NoError(s.Add(ctx, fresh))
	stored, closer, err := db.Get(storage.KeyFromPeer(fresh).Bytes(nil))
	a.NoError(err)
	a.NotEqual(byte('{'), stored[0], "new values must not be JSON")
	a.NoError(closer.Close())
	p, err = s.Find(ctx, storage.KeyFromPeer(fresh))
	a.NoError(err)
	a.Equal(int64(2), p.Key.AccessHash)
}
