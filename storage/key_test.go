package storage

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/query/dialogs"

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

	b := k.Bytes(nil)
	a.NoError(k.Parse(b))

	s := k.String()
	sBytes := []byte(s)
	a.Equal(b, sBytes)
	a.NoError(k.Parse(sBytes))
}

func TestKey_Parse(t *testing.T) {
	tests := []struct {
		fields  PeerKey
		arg     string
		wantErr bool
	}{
		{PeerKey{}, "_10", true},
		{PeerKey{}, "10", true},
		{PeerKey{}, "10_", true},
		{PeerKey{}, "10_1", true},
		{PeerKey{}, "peer10_1", true},
		{PeerKey{
			Kind: dialogs.Channel,
			ID:   1,
		}, "peer" + strconv.Itoa(int(dialogs.Channel)) + "_1", false},
	}
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			k := PeerKey{}

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
