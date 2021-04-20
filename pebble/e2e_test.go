package pebble_test

import (
	"testing"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"

	"github.com/gotd/contrib/internal/tests"
	"github.com/gotd/contrib/pebble"
)

func TestE2E(t *testing.T) {
	db, err := pebbledb.Open("pebble.db", &pebbledb.Options{
		FS: vfs.NewMem(),
	})
	if err != nil {
		t.Fatal(err)
	}

	tests.TestSessionStorage(t, pebble.NewSessionStorage(db, "testsession"))
	tests.TestCredentials(t, pebble.NewCredentials(db))
	tests.TestPeerStorage(t, pebble.NewPeerStorage(db))
}
