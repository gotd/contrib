package s3_test

import (
	"context"
	"os"
	"testing"

	"github.com/minio/minio-go/v7"

	"github.com/gotd/contrib/internal/tests"
	"github.com/gotd/contrib/s3"
)

func TestE2E(t *testing.T) {
	addr := os.Getenv("S3_ADDR")
	if addr == "" {
		t.Skip("Set S3_ADDR to run E2E test")
	}

	db, err := minio.New(addr, &minio.Options{})
	if err != nil {
		t.Fatal(err)
	}
	tests.RetryUntilAvailable(t, "s3", addr, func(ctx context.Context) error {
		_, err := db.ListBuckets(ctx)
		return err
	})

	tests.TestSessionStorage(t, s3.NewSessionStorage(db, "testsession", "session"))
}
