package s3

import (
	"bytes"
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"golang.org/x/xerrors"

	"github.com/gotd/td/session"
)

var _ session.Storage = SessionStorage{}

// SessionStorage is a MTProto session S3 storage.
type SessionStorage struct {
	client                 *minio.Client
	bucketName, objectName string
}

// NewSessionStorage creates new SessionStorage.
func NewSessionStorage(client *minio.Client, bucketName, objectName string) SessionStorage {
	return SessionStorage{
		client:     client,
		bucketName: bucketName,
		objectName: objectName,
	}
}

// LoadSession implements session.Storage.
func (s SessionStorage) LoadSession(ctx context.Context) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucketName, s.objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, xerrors.Errorf("get %q/%q: %w", s.bucketName, s.objectName, err)
	}
	return io.ReadAll(obj)
}

// StoreSession implements session.Storage.
func (s SessionStorage) StoreSession(ctx context.Context, data []byte) error {
	if err := s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{}); err != nil {
		return xerrors.Errorf("create bucket %q: %w", s.bucketName, err)
	}

	_, err := s.client.PutObject(ctx, s.bucketName, s.objectName,
		bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{
			ContentType: "application/json",
			NumThreads:  1,
		},
	)
	if err != nil {
		return xerrors.Errorf("put %q/%q: %w", s.bucketName, s.objectName, err)
	}

	return nil
}
