package storage

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestPeerUnchanged(t *testing.T) {
	a := require.New(t)
	var p Peer
	a.True(p.FromUser(&tg.User{ID: 1, AccessHash: 2, Username: "foo"}))
	stored, err := EncodePeer(p)
	a.NoError(err)

	same := p
	same.CreatedAt = p.CreatedAt.Add(time.Hour)
	a.True(PeerUnchanged(stored, same), "CreatedAt is ignored")

	var changed Peer
	a.True(changed.FromUser(&tg.User{ID: 1, AccessHash: 2, Username: "bar"}))
	a.False(PeerUnchanged(stored, changed))

	a.False(PeerUnchanged([]byte("garbage"), p))

	// Peer stored in the old JSON format is rewritten once in the new one.
	legacy, err := json.Marshal(p)
	a.NoError(err)
	a.False(PeerUnchanged(legacy, p))
}
