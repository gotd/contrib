package upload

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"log"
	"time"

	"github.com/tdakkota/tgcontrib/creator"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func ExampleUpload() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err := creator.UserFromEnvironment(ctx, telegram.Options{}, nil,
		func(ctx context.Context, client *telegram.Client) error {
			img := image.NewRGBA(image.Rect(0, 0, 512, 512))
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

			buf := bytes.NewBuffer(nil)
			if err := png.Encode(buf, img); err != nil {
				return err
			}

			raw := tg.NewClient(client)
			f, err := NewUploader(raw).Upload(ctx, NewUpload("abc.jpg", buf))
			if err != nil {
				return err
			}

			req := &tg.PhotosUploadProfilePhotoRequest{}
			req.SetFile(f)
			_, err = raw.PhotosUploadProfilePhoto(ctx, req)
			if err != nil {
				return err
			}

			return nil
		})

	if err != nil {
		log.Fatal(err)
	}
}
