package proto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestPeer(t *testing.T) {
	a := require.New(t)
	user := &tg.InputPeerUser{
		UserID:     10,
		AccessHash: 10,
	}

	p, err := FromInputPeer(user)
	a.NoError(err)

	data, err := json.Marshal(p)
	a.NoError(err)

	var p2 Peer
	a.NoError(json.Unmarshal(data, &p2))
	a.Equal(p, p2)

	_, ok := p2.AsInputUser()
	a.True(ok)

	u := p2.AsInputPeer()
	a.Equal(user, u)
}
