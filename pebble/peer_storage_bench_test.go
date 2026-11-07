package pebble_test

import (
	"context"
	"testing"

	pebbledb "github.com/cockroachdb/pebble"

	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/pebble"
	"github.com/gotd/contrib/storage"
)

func BenchmarkPeerStorageAddUnchanged(b *testing.B) {
	db, err := pebbledb.Open(b.TempDir(), &pebbledb.Options{})
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	s := pebble.NewPeerStorage(db)
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
