package storage

import (
	"strconv"
	"testing"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func TestKey(t *testing.T) {
	a := require.New(t)
	var p Peer
	a.NoError(p.FromInputPeer(&tg.InputPeerUser{
		UserID:     10,
		AccessHash: 10,
	}))
	k := KeyFromPeer(p)

	data := k.Bytes(nil)
	a.NoError(k.Parse(data))
}

func TestKey_Parse(t *testing.T) {
	tests := []struct {
		fields  Key
		arg     string
		wantErr bool
	}{
		{Key{}, "_10", true},
		{Key{}, "10", true},
		{Key{}, "10_", true},
		{Key{}, "10_1", true},
		{Key{}, "peer10_1", true},
		{Key{
			Kind: peer.Channel,
			ID:   1,
		}, "peer" + strconv.Itoa(int(peer.Channel)) + "_1", false},
	}
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			k := Key{}

			err := k.Parse([]byte(tt.arg))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.fields, k)
			}
		})
	}
}
