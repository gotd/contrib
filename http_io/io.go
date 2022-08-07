// Package http_io implements http handlers based on partial input/output primitives.
package http_io

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/gotd/contrib/http_range"
)

// StreamerAt implements streaming with offset.
type StreamerAt interface {
	StreamAt(ctx context.Context, skip int64, w io.Writer) error
}

// Handler implements ranged http requests on top of StreamerAt interface.
type Handler struct {
	log         *zap.Logger
	size        int64
	contentType string
	streamer    StreamerAt
}

// WithLog sets logger of handler.
func (h *Handler) WithLog(log *zap.Logger) *Handler {
	h.log = log
	return h
}

// WithContentType sets contentType header.
func (h *Handler) WithContentType(contentType string) *Handler {
	h.contentType = contentType
	return h
}

// ServeHTTP implements http.Handler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ranges, err := http_range.ParseRange(r.Header.Get("Range"), h.size)
	if err == http_range.ErrNoOverlap {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", h.size))
		http.Error(w, http_range.ErrNoOverlap.Error(), http.StatusRequestedRangeNotSatisfiable)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(ranges) > 1 {
		http.Error(w, "multiple ranges are not supported", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// Preparing response.
	code := http.StatusOK
	sendSize := h.size

	var offset int64
	if len(ranges) > 0 {
		r := ranges[0]
		offset = r.Start
		sendSize = r.Length
		code = http.StatusPartialContent
		w.Header().Set("Content-Range", r.ContentRange(h.size))
	}
	if h.contentType != "" {
		w.Header().Set("Content-Type", h.contentType)
	}
	w.Header().Set("Accept-Ranges", "bytes")
	if w.Header().Get("Content-Encoding") == "" {
		w.Header().Set("Content-Length", strconv.FormatInt(sendSize, 10))
	}

	// Writing response.
	w.WriteHeader(code)
	if r.Method == http.MethodHead {
		// Not writing body on HEAD.
		return
	}
	h.log.Info("Serving", zap.Int64("offset", offset))
	// TODO: handle case of partial writes (e.g. not until the end of file).
	if err := h.streamer.StreamAt(r.Context(), offset, w); err != nil && !errors.Is(err, context.Canceled) {
		h.log.Error("Failed to stream", zap.Error(err))
		return
	}
}

// NewHandler initializes and returns http handler for ranged requests using
// provided StreamerAt as file source and total file size.
func NewHandler(s StreamerAt, size int64) *Handler {
	return &Handler{
		log:      zap.NewNop(),
		size:     size,
		streamer: s,
	}
}
