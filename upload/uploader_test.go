package upload

import (
	"bytes"
	"context"
	"crypto/rand"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
)

type Image func() *image.RGBA

func testUploader(creator Image) func(t *testing.T) {
	return func(t *testing.T) {
		a := require.New(t)

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
			Addr: telegram.AddrTest,
		})

		a.NoError(client.Connect(ctx))
		defer func() {
			_ = client.Close()
		}()

		a.NoError(telegram.NewAuth(
			telegram.TestAuth(rand.Reader, 2),
			telegram.SendCodeOptions{},
		).Run(ctx, client))

		_, err := client.Self(ctx)
		a.NoError(err)

		img := bytes.NewBuffer(nil)
		a.NoError(png.Encode(img, creator()))
		t.Log("size of image", img.Len(), "bytes")

		raw := tg.NewClient(client)
		f, err := NewUploader(raw).Upload(ctx, NewUpload("abc.jpg", img))
		a.NoError(err)

		req := &tg.PhotosUploadProfilePhotoRequest{}
		req.SetFile(f)
		res, err := raw.PhotosUploadProfilePhoto(ctx, req)
		a.NoError(err)

		_, ok := res.Photo.(*tg.Photo)
		a.Truef(ok, "unexpected type %T", res.Photo)
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
