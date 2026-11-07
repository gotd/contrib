package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestPeerUnchanged(t *testing.T) {
	for _, enc := range []PeerEncoding{PeerEncodingJSON, PeerEncodingBinary} {
		a := require.New(t)
		var p Peer
		a.True(p.FromUser(&tg.User{ID: 1, AccessHash: 2, Username: "foo"}))
		stored, err := enc.Encode(p)
		a.NoError(err)

		same := p
		same.CreatedAt = p.CreatedAt.Add(time.Hour)
		a.True(PeerUnchanged(stored, same, enc), "CreatedAt is ignored")

		var changed Peer
		a.True(changed.FromUser(&tg.User{ID: 1, AccessHash: 2, Username: "bar"}))
		a.False(PeerUnchanged(stored, changed, enc))

		a.False(PeerUnchanged([]byte("garbage"), p, enc))

		// Peer stored in the other format is rewritten once in this one.
		other := PeerEncodingBinary
		if enc == PeerEncodingBinary {
			other = PeerEncodingJSON
		}
		a.False(PeerUnchanged(stored, p, other))
	}
}
