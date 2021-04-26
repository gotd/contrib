package bbolt_test

import (
	"os"
	"testing"

	bboltdb "go.etcd.io/bbolt"

	"github.com/gotd/contrib/bbolt"
	"github.com/gotd/contrib/internal/tests"
)

func TestE2E(t *testing.T) {
	db, err := bboltdb.Open("bbolt.db", 0, &bboltdb.Options{
		OpenFile: func(s string, flag int, mode os.FileMode) (*os.File, error) {
			return os.CreateTemp("", "*"+s)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	bucket := []byte("test")

	tests.TestSessionStorage(t, bbolt.NewSessionStorage(db, "testsession", bucket))
	tests.TestCredentials(t, bbolt.NewCredentials(db, bucket))
	tests.TestPeerStorage(t, bbolt.NewPeerStorage(db, bucket))
}
