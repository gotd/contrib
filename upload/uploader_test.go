package upload

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"testing"
	"testing/iotest"
	"time"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"

	"github.com/tdakkota/tgcontrib/creator"
)

type Image func() *image.RGBA

func testUploader(gen Image) func(t *testing.T) {
	return func(t *testing.T) {
		a := require.New(t)

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		err := creator.TestClient(ctx, telegram.Options{}, func(ctx context.Context, client *telegram.Client) error {
			if _, err := client.Self(ctx); err != nil {
				return xerrors.Errorf("self: %w", err)
			}

			img := bytes.NewBuffer(nil)
			if err := png.Encode(img, gen()); err != nil {
				return xerrors.Errorf("png encode: %w", err)
			}
			t.Log("size of image", img.Len(), "bytes")

			raw := tg.NewClient(client)
			upld := NewUpload("abc.jpg", iotest.HalfReader(img))
			f, err := NewUploader(raw).WithPartSize(2048).Upload(ctx, upld)
			if err != nil {
				return xerrors.Errorf("upload: %w", err)
			}

			req := &tg.PhotosUploadProfilePhotoRequest{}
			req.SetFile(f)
			res, err := raw.PhotosUploadProfilePhoto(ctx, req)
			if err != nil {
				return xerrors.Errorf("change profile photo: %w", err)
			}

			_, ok := res.Photo.(*tg.Photo)
			a.Truef(ok, "unexpected type %T", res.Photo)
			return nil
		})

		a.NoError(err)
	}
}

func generateImage(x, y int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, x, y))
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			if (x+y)%2 == 0 || (x%2 != 0 && y%2 != 0) {
				img.SetRGBA(x, y, color.RGBA{
					R: 255,
					G: 255,
					B: 255,
					A: 255,
				})
			}
		}
	}
	return img
}

func TestUploader(t *testing.T) {
	t.Run("small", testUploader(func() *image.RGBA {
		return generateImage(255, 255)
	}))

	t.Run("big", testUploader(func() *image.RGBA {
		return generateImage(1024, 1024)
	}))
}
