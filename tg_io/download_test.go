package tg_io

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/http_io"
	"github.com/gotd/contrib/partio"
)

const (
	chunk1kb = 1024
)

func TestE2E(t *testing.T) {
	if os.Getenv("TG_IO_E2E") != "1" {
		t.Skip("TG_IO_E2E not set")
	}

	client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, telegram.Options{
		DC:     2,
		DCList: dcs.StagingDCs(),
	})
	api := tg.NewClient(client)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	require.NoError(t, client.Run(ctx, func(ctx context.Context) error {
		if err := telegram.NewAuth(
			telegram.TestAuth(rand.Reader, 2),
			telegram.SendCodeOptions{},
		).Run(ctx, client); err != nil {
			return err
		}

		const size = chunk1kb*5 + 100
		f, err := uploader.NewUploader(api).FromBytes(ctx, "upload.bin", make([]byte, size))
		if err != nil {
			return err
		}

		mc, err := message.NewSender(api).Self().UploadMedia(ctx, message.File(f))
		if err != nil {
			return xerrors.Errorf("file: %w", err)
		}
		doc, ok := mc.(*tg.MessageMediaDocument).Document.AsNotEmpty()
		if !ok {
			return xerrors.New("bad doc")
		}

		t.Log("Streaming")
		u := partio.NewStreamer(NewDownloader(api).ChunkSource(doc.Size, doc.AsInputDocumentFileLocation()), chunk1kb)
		buf := new(bytes.Buffer)

		const offset = chunk1kb / 2
		if err := u.StreamAt(ctx, offset, buf); err != nil {
			return err
		}

		t.Log(buf.Len())
		assert.Equal(t, doc.Size-offset, buf.Len())

		ln, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			return err
		}
		defer func() {
			_ = ln.Close()
		}()
		s := http.Server{
			Handler: http_io.NewHandler(u, doc.Size).WithContentType(doc.MimeType),
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
				return err
			}
			return nil
		})
		g.Go(func() error {
			defer close(done)

			req, err := http.NewRequest(http.MethodGet, "http://"+ln.Addr().String(), nil)
			if err != nil {
				return err
			}
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer func() { _ = res.Body.Close() }()
			t.Log(res.Status)

			outBuf := new(bytes.Buffer)
			if _, err := io.Copy(outBuf, res.Body); err != nil {
				return err
			}

			t.Log(outBuf.Len())

			return nil
		})

		return g.Wait()
	}))
}
