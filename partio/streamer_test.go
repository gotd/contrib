package partio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPadding(t *testing.T) {
	for _, tt := range []struct {
		padding, offset, result int64
	}{
		{},
		{
			padding: 8,
			offset:  2,
			result:  0,
		},
		{
			padding: 8,
			offset:  9,
			result:  8,
		},
		{
			padding: 8,
			offset:  8,
			result:  8,
		},
		{
			padding: 1024,
			offset:  413,
			result:  0,
		},
		{
			offset: 514,
			result: 514,
		},
	} {
		t.Run(fmt.Sprintf("%d_%d", tt.padding, tt.offset), func(t *testing.T) {
			require.Equal(t, tt.result, nearestOffset(tt.padding, tt.offset))
		})
	}
}

type BytesReader struct {
	Data  []byte
	Align int64
}

func (r BytesReader) Chunk(ctx context.Context, offset int64, b []byte) (int64, error) {
	if offset%r.Align != 0 {
		return 0, errors.New("unaligned")
	}
	if int64(len(b)) != r.Align {
		return 0, errors.New("invalid chunk size")
	}

	if offset > int64(len(r.Data)) {
		return 0, io.EOF
	}
	buf := r.Data[offset:]
	n := int64(copy(b, buf))
	if n != r.Align {
		return n, io.EOF
	}

	return n, nil
}

type StreamReader struct {
	Align int64
	Total int64
}

func (r StreamReader) Chunk(ctx context.Context, offset int64, b []byte) (int64, error) {
	if offset%r.Align != 0 {
		return 0, errors.New("unaligned")
	}
	if int64(len(b)) != r.Align {
		return 0, errors.New("invalid chunk size")
	}
	if offset > r.Total {
		return 0, io.EOF
	}

	n := r.Align
	if (offset + n) > r.Total {
		// Last chunk.
		n = r.Total - offset
	}
	for i := range b {
		b[i] = byte(i)
	}
	if n != r.Align {
		return n, io.EOF
	}
	return n, nil
}

func TestReaderFrom_Stream(t *testing.T) {
	ctx := context.Background()
	t.Run("Simple", func(t *testing.T) {
		const chunkSize = 3
		s := NewStreamer(&BytesReader{
			Align: chunkSize,
			Data:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}, chunkSize)
		t.Run("Equal", func(t *testing.T) {
			out := new(bytes.Buffer)
			require.NoError(t, s.StreamAt(ctx, 2, out))
			require.Equal(t, []byte{3, 4, 5, 6, 7, 8}, out.Bytes())
		})
		t.Run("Discard", func(t *testing.T) {
			require.NoError(t, s.Stream(ctx, io.Discard))
		})
	})
	t.Run("Stream", func(t *testing.T) {
		const (
			chunkSize = 1024
			total     = chunkSize*100 + 56
		)
		s := NewStreamer(&StreamReader{
			Align: chunkSize,
			Total: total,
		}, chunkSize)
		t.Run("Equal", func(t *testing.T) {
			buf := new(bytes.Buffer)
			require.NoError(t, s.StreamAt(ctx, total-chunkSize, buf))
			require.Equal(t, byte(56), buf.Bytes()[0])
			require.Equal(t, 1024, buf.Len())
		})
		t.Run("Discard", func(t *testing.T) {
			require.NoError(t, s.Stream(ctx, io.Discard))
		})
	})
}

func BenchmarkReaderFrom_StreamAt(b *testing.B) {
	const (
		chunkSize = 1024
		total     = chunkSize*100 + 56
		offset    = chunkSize/2 + 10
	)
	s := NewStreamer(&StreamReader{
		Align: chunkSize,
		Total: total,
	}, chunkSize)

	b.ReportAllocs()
	b.SetBytes(total - offset)

	for i := 0; i < b.N; i++ {
		if err := s.StreamAt(context.Background(), offset, io.Discard); err != nil {
			b.Fatal()
		}
	}
}
