package bbolt

import (
	"context"
	"path"
	"testing"

	bboltdb "go.etcd.io/bbolt"

	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/storage"
)

func BenchmarkPeerStorageAddUnchanged(b *testing.B) {
	db, err := bboltdb.Open(path.Join(b.TempDir(), "bbolt.db"), 0666, &bboltdb.Options{}) // nolint:gocritic
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	s := NewPeerStorage(db, []byte("peers"))
	ctx := context.Background()

	var p storage.Peer
	if !p.FromUser(&tg.User{ID: 1, AccessHash: 1, Username: "foo", FirstName: "Foo"}) {
		b.Fatal("from user")
	}
	if err := s.Add(ctx, p); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		if err := s.Add(ctx, p); err != nil {
			b.Fatal(err)
		}
	}
}
