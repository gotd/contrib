package s3

import (
	"bytes"
	"context"
	"io"

	"github.com/go-faster/errors"
	"github.com/minio/minio-go/v7"

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
		return nil, errors.Wrapf(err, "get %q/%q", s.bucketName, s.objectName)
	}
	return io.ReadAll(obj)
}

// StoreSession implements session.Storage.
func (s SessionStorage) StoreSession(ctx context.Context, data []byte) error {
	if err := s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{}); err != nil {
		return errors.Wrapf(err, "create bucket %q", s.bucketName)
	}

	_, err := s.client.PutObject(ctx, s.bucketName, s.objectName,
		bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{
			ContentType: "application/json",
			NumThreads:  1,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "put %q/%q", s.bucketName, s.objectName)
	}

	return nil
}
