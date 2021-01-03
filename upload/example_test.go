package upload

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func ExampleUpload() {
	// Reading app id from env (never hardcode it!).
	appID, err := strconv.Atoi(os.Getenv("APP_ID"))
	if err != nil {
		log.Fatal(err)
	}

	appHash := os.Getenv("APP_HASH")
	if appHash == "" {
		log.Fatal("no APP_HASH provided")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	client := telegram.NewClient(appID, appHash, telegram.Options{})

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = client.Close()
	}()

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
		log.Fatal(err)
	}

	raw := tg.NewClient(client)
	f, err := NewUploader(raw).Upload(ctx, NewUpload("abc.jpg", buf))
	if err != nil {
		log.Fatal(err)
	}

	req := &tg.PhotosUploadProfilePhotoRequest{}
	req.SetFile(f)
	_, err = raw.PhotosUploadProfilePhoto(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
}
