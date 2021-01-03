package upload

import (
	"encoding/hex"
	"io"

	"github.com/gotd/td/tg"
)

// NewUpload creates new Upload struct using given
// name and reader.
func NewUpload(name string, from io.Reader) *Upload {
	return &Upload{
		name: name,
		from: from,
	}
}

// File is file abstraction.
type File interface {
	Name() string
	io.Reader
}

// FromFile creates new Upload struct using
// given File.
func FromFile(f File) *Upload {
	return NewUpload(f.Name(), f)
}

// Upload represents Telegram file upload.
type Upload struct {
	id       int64
	uploaded int
	name     string
	from     io.Reader
}

func (u *Upload) nextPart(data []byte) *tg.UploadSaveFilePartRequest {
	return &tg.UploadSaveFilePartRequest{
		FileID:   u.id,
		FilePart: u.uploaded,
		Bytes:    data,
	}
}

func (u *Upload) fileObject(hash []byte) (tg.InputFileClass, error) {
	return &tg.InputFile{
		ID:          u.id,
		Parts:       u.uploaded,
		Name:        u.name,
		Md5Checksum: hex.EncodeToString(hash),
	}, nil
}
