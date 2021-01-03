package upload

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func randInt64(randSource io.Reader) (int64, error) {
	var buf [bin.Word * 2]byte
	if _, err := io.ReadFull(randSource, buf[:]); err != nil {
		return 0, err
	}
	b := &bin.Buffer{Buf: buf[:]}
	return b.Long()
}

// Uploader is Telegram file uploader.
type Uploader struct {
	rpc      *tg.Client
	id       func() (int64, error)
	partSize int
}

// NewUploader creates new Uploader.
func NewUploader(rpc *tg.Client) *Uploader {
	return &Uploader{
		rpc: rpc,
		id: func() (int64, error) {
			return randInt64(rand.Reader)
		},
		partSize: 1024,
	}
}

// WithIDGenerator sets id generator.
func (u *Uploader) WithIDGenerator(cb func() (int64, error)) *Uploader {
	u.id = cb
	return u
}

// WithPartSize sets part size.
// Should be divisible by 1024.
// 524288 should be divisible by partSize.
//
// See https://core.telegram.org/api/files#uploading-files.
func (u *Uploader) WithPartSize(partSize int) *Uploader {
	u.partSize = partSize
	return u
}

// Upload uploads data from Upload object.
func (u *Uploader) Upload(ctx context.Context, upld *Upload) (tg.InputFileClass, error) {
	if upld.id == 0 {
		id, err := u.id()
		if err != nil {
			return nil, xerrors.Errorf("id generation: %w", err)
		}

		upld.id = id
	}

	buf := make([]byte, u.partSize)
	hash := md5.New()
	for {
		_, err := upld.from.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, xerrors.Errorf("read source: %w", err)
		}
		_, _ = hash.Write(buf)

		r, err := u.rpc.UploadSaveFilePart(ctx, upld.nextPart(buf))
		if err != nil {
			return nil, xerrors.Errorf("send upload RPC: %w", err)
		}

		if !r {
			continue
		}

		upld.uploaded++
	}

	return upld.fileObject(hash.Sum(buf[:0]))
}
