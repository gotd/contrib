package http_io

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/contrib/partio"
)

// bytesSource pretends to be Telegram: it serves fixed-size chunks
// from an in-memory slice, like upload.getFile does.
type bytesSource struct {
	data []byte
}

func (s bytesSource) Chunk(_ context.Context, offset int64, b []byte) (int64, error) {
	if offset >= int64(len(s.data)) {
		return 0, io.EOF
	}
	n := int64(copy(b, s.data[offset:]))
	if offset+n >= int64(len(s.data)) {
		return n, io.EOF // last chunk
	}
	return n, nil
}

func TestHandler_Range(t *testing.T) {
	const (
		chunkSize = 1024
		size      = chunkSize*10 + 100 // not a multiple of chunkSize on purpose
	)
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 251) // non-repeating-ish pattern so wrong offsets are caught
	}

	h := NewHandler(partio.NewStreamer(bytesSource{data: data}, chunkSize), size)
	// Real server, not ResponseRecorder: only the real one enforces Content-Length.
	srv := httptest.NewServer(h)
	defer srv.Close()

	for _, tt := range []struct {
		name       string
		rangeHdr   string
		start, end int64 // inclusive, like HTTP ranges
		code       int
	}{
		{name: "Full", start: 0, end: size - 1, code: http.StatusOK},
		{name: "SmallerThanChunk", rangeHdr: "bytes=0-99", start: 0, end: 99, code: http.StatusPartialContent},
		{name: "UnalignedWithinChunk", rangeHdr: "bytes=100-500", start: 100, end: 500, code: http.StatusPartialContent},
		{name: "AcrossChunks", rangeHdr: "bytes=1000-3100", start: 1000, end: 3100, code: http.StatusPartialContent},
		{name: "ExactChunk", rangeHdr: "bytes=1024-2047", start: 1024, end: 2047, code: http.StatusPartialContent},
		{name: "SingleByte", rangeHdr: "bytes=5000-5000", start: 5000, end: 5000, code: http.StatusPartialContent},
		{name: "OpenEnded", rangeHdr: "bytes=9000-", start: 9000, end: size - 1, code: http.StatusPartialContent},
		{name: "Suffix", rangeHdr: "bytes=-150", start: size - 150, end: size - 1, code: http.StatusPartialContent},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
			require.NoError(t, err)
			if tt.rangeHdr != "" {
				req.Header.Set("Range", tt.rangeHdr)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t, tt.code, resp.StatusCode)
			require.Equal(t, data[tt.start:tt.end+1], body)
			if tt.code == http.StatusPartialContent {
				require.Equal(t,
					fmt.Sprintf("bytes %d-%d/%d", tt.start, tt.end, size),
					resp.Header.Get("Content-Range"),
				)
			}
		})
	}
}
