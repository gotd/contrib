package pebble_test

import (
	"testing"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"

	"github.com/gotd/contrib/auth/internal/tests"
	"github.com/gotd/contrib/auth/pebble"
)

func TestE2E(t *testing.T) {
	db, err := pebbledb.Open("pebble.db", &pebbledb.Options{
		FS: vfs.NewMem(),
	})
	if err != nil {
		t.Fatal(err)
	}

	tests.TestStorage(
		t,
		pebble.NewSessionStorage(db, "testsession"),
		pebble.NewCredentials(db),
	)
}
