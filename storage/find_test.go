package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestFindPeer(t *testing.T) {
	a := require.New(t)
	ctx := context.Background()
	s := newMemStorage()

	var p Peer
	a.NoError(p.FromInputPeer(&tg.InputPeerUser{
		UserID:     10,
		AccessHash: 10,
	}))
	a.NoError(s.Assign(ctx, "domain", p))

	p2, err := FindPeer(ctx, s, &tg.PeerUser{
		UserID: 10,
	})
	a.NoError(err)
	a.Equal(p, p2)
}

type testIterator struct {
	buf    []Peer
	cursor int
}

func (t *testIterator) Next(ctx context.Context) bool {
	if t.cursor < len(t.buf) {
		t.cursor++
		return true
	}

	return false
}

func (t *testIterator) Err() error {
	return nil
}

func (t *testIterator) Value() Peer {
	return t.buf[t.cursor-1]
}

func (t *testIterator) Close() error {
	return nil
}

func TestForEach(t *testing.T) {
	a := require.New(t)

	buf := func() (r []Peer) {
		for i := range [5]struct{}{} {
			var p Peer
			a.NoError(p.FromInputPeer(&tg.InputPeerUser{
				UserID:     int64(i) + 11,
				AccessHash: int64(i) + 11,
			}))
			r = append(r, p)
		}

		return
	}()

	iter := &testIterator{
		buf:    buf,
		cursor: 0,
	}

	i := 0
	a.NoError(ForEach(context.Background(), iter, func(peer Peer) error {
		a.Equal(buf[i], peer)
		i++
		return nil
	}))
}
