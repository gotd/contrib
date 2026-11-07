package pebble_test

import (
	"context"
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
