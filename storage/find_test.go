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
