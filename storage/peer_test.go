package storage

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
)

func TestPeer_MarshalJSON(t *testing.T) {
	t.Run("Latest", func(t *testing.T) {
		// Prepare data.
		chat := &tg.Chat{
			Photo: &tg.ChatPhotoEmpty{},
		}
		chat.SetFlags()

		user := &tg.User{
			Username:   "foo",
			ID:         100,
			AccessHash: 200,
			Photo:      &tg.UserProfilePhotoEmpty{},
		}
		user.SetFlags()

		channel := &tg.Channel{
			Photo:             &tg.ChatPhotoEmpty{},
			ParticipantsCount: 200,
		}
		channel.SetFlags()

		key := dialogs.DialogKey{
			ID:         1,
			AccessHash: 3,
			Kind:       dialogs.User,
		}
		meta := map[string]any{
			"foo": "bar",
			"v":   true,
		}

		for i, p := range []Peer{
			{
				Version:   LatestVersion,
				Key:       key,
				User:      user,
				Chat:      chat,
				Channel:   channel,
				Metadata:  meta,
				CreatedAt: time.Unix(1681541743, 0),
			},
			{
				Version:   LatestVersion,
				Key:       key,
				CreatedAt: time.Unix(1682525712, 0),
			},
		} {
			t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
				{
					// Just print.
					var e jx.Encoder
					e.SetIdent(2)
					require.NoError(t, p.Marshal(&e))
					t.Log(e.String())
				}

				data, err := json.Marshal(p)
				require.NoError(t, err)

				var out Peer
				require.NoError(t, json.Unmarshal(data, &out))

				assert.Equal(t, p.CreatedAt, out.CreatedAt, "CreatedAt")
				assert.Equal(t, p.Version, out.Version, "Version")
				assert.Equal(t, p.Key, out.Key, "Key")
				assert.Equal(t, p.Metadata, out.Metadata, "Metadata")
				if assert.True(t, (p.User == nil) == (out.User == nil), "User nil") && p.User != nil {
					assert.Equal(t, *p.User, *out.User, "User")
				}
				if assert.True(t, (p.Chat == nil) == (out.Chat == nil), "Chat nil") && p.Chat != nil {
					assert.Equal(t, *p.Chat, *out.Chat, "Chat")
				}
				if assert.True(t, (p.Channel == nil) == (out.Channel == nil), "Channel nil") && p.Channel != nil {
					assert.Equal(t, *p.Channel, *out.Channel, "Channel")
				}
			})
		}
	})
	t.Run("Outdated", func(t *testing.T) {
		var d jx.Encoder
		d.SetIdent(2)
		d.Obj(func(e *jx.Encoder) {
			e.Field("Version", func(e *jx.Encoder) {
				e.Int(1)
			})
		})
		err := json.Unmarshal(d.Bytes(), &Peer{})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrPeerUnmarshalMustInvalidate)
	})
}
