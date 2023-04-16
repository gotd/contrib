package tg_io

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/http_io"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/contrib/partio"
)

const (
	chunk1kb = 1024
)

func TestE2E(t *testing.T) {
	if os.Getenv("TG_IO_E2E") != "1" {
		t.Skip("TG_IO_E2E not set")
	}
	logger := zaptest.NewLogger(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	floodWaiter := floodwait.NewWaiter()

	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
		DC:     2,
		DCList: dcs.Test(),
		Logger: logger.Named("client"),
		Middlewares: []telegram.Middleware{
			ratelimit.New(rate.Every(100*time.Millisecond), 5),
			floodWaiter,
		},
	})
	api := tg.NewClient(client)

	handler := func(ctx context.Context) error {
		authClient := auth.NewClient(api, rand.Reader, telegram.TestAppID, telegram.TestAppHash)
		if err := auth.NewFlow(
			auth.Test(rand.Reader, 2),
			auth.SendCodeOptions{},
		).Run(ctx, authClient); err != nil {
			return err
		}

		const size = chunk1kb*5 + 100
		f, err := uploader.NewUploader(api).FromBytes(ctx, "upload.bin", make([]byte, size))
		if err != nil {
			return errors.Errorf("upload: %w", err)
		}

		mc, err := message.NewSender(api).Self().UploadMedia(ctx, message.File(f))
		if err != nil {
			return errors.Errorf("create media: %w", err)
		}

		media, ok := mc.(*tg.MessageMediaDocument)
		if !ok {
			return errors.Errorf("unexpected type: %T", media)
		}

		doc, ok := media.Document.AsNotEmpty()
		if !ok {
			return errors.Errorf("unexpected type: %T", media.Document)
		}

		t.Log("Streaming")
		u := partio.NewStreamer(NewDownloader(api).ChunkSource(doc.Size, doc.AsInputDocumentFileLocation()), chunk1kb)
		buf := new(bytes.Buffer)

		const offset = chunk1kb / 2
		if err := u.StreamAt(ctx, offset, buf); err != nil {
			return errors.Errorf("stream at %d: %w", offset, err)
		}

		t.Log(buf.Len())
		assert.Equal(t, doc.Size-offset, int64(buf.Len()))

		ln, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			return errors.Errorf("listen: %w", err)
		}
		defer func() {
			_ = ln.Close()
		}()
		s := http.Server{
			Handler: http_io.NewHandler(u, doc.Size).
				WithContentType(doc.MimeType).
				WithLog(logger.Named("httpio")),
		}
		g, ctx := errgroup.WithContext(ctx)
		done := make(chan struct{})
		g.Go(func() error {
			select {
			case <-ctx.Done():
			case <-done:
			}
			return s.Close()
		})
		g.Go(func() error {
			if err := s.Serve(ln); err != nil && err != http.ErrServerClosed {
				return errors.Errorf("server: %w", err)
			}
			return nil
		})
		g.Go(func() error {
			defer close(done)

			requestURL := &url.URL{
				Scheme: "http",
				Host:   ln.Addr().String(),
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), http.NoBody)
			if err != nil {
				return err
			}
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return errors.Errorf("send GET %q: %w", requestURL, err)
			}
			defer func() { _ = res.Body.Close() }()
			t.Log(res.Status)

			outBuf := new(bytes.Buffer)
			if _, err := io.Copy(outBuf, res.Body); err != nil {
				return errors.Errorf("read response: %w", err)
			}

			t.Log(outBuf.Len())

			return nil
		})

		return g.Wait()
	}
	run := func(ctx context.Context) error {
		return client.Run(ctx, handler)
	}
	require.NoError(t, floodWaiter.Run(ctx, run))
}
