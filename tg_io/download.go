// Package tg_io implements partial i/o using telegram.
package tg_io

import (
	"context"
	"io"

	"github.com/go-faster/errors"

	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/partio"
)

// Downloader implements streamable file downloads of Telegram files.
type Downloader struct {
	api *tg.Client
}

// NewDownloader creates new Downloader.
func NewDownloader(api *tg.Client) *Downloader {
	return &Downloader{
		api: api,
	}
}

// ChunkSource creates new chunk source for provided file.
func (d *Downloader) ChunkSource(size int64, loc tg.InputFileLocationClass) partio.ChunkSource {
	return &chunkSource{
		loc:  loc,
		api:  d.api,
		size: size,
	}
}

type chunkSource struct {
	loc  tg.InputFileLocationClass
	api  *tg.Client
	size int64
}

// Chunk implements partio.ChunkSource.
func (s chunkSource) Chunk(ctx context.Context, offset int64, b []byte) (int64, error) {
	req := &tg.UploadGetFileRequest{
		Offset:   offset,
		Limit:    len(b),
		Location: s.loc,
	}
	req.SetPrecise(true)

	r, err := s.api.UploadGetFile(ctx, req)
	if err != nil {
		return 0, err
	}

	switch result := r.(type) {
	case *tg.UploadFile:
		n := int64(copy(b, result.Bytes))

		var err error
		if req.Offset+n >= s.size {
			// No more data.
			err = io.EOF
		}

		return n, err
	default:
		return 0, errors.Errorf("unexpected type %T", r)
	}
}
