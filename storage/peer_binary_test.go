package storage

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

func testPeers(t *testing.T) []Peer {
	t.Helper()
	user := Peer{}
	require.True(t, user.FromUser(&tg.User{
		ID: 10, AccessHash: 20, Username: "foo", FirstName: "Foo", Phone: "123",
	}))
	user.Metadata = map[string]any{"k": "v", "n": 1.5}

	chat := Peer{}
	require.True(t, chat.FromChat(&tg.Chat{ID: 11, Title: "chat", Photo: &tg.ChatPhotoEmpty{}}))

	channel := Peer{}
	require.True(t, channel.FromChat(&tg.Channel{ID: 12, AccessHash: 22, Title: "channel", Username: "ch", Photo: &tg.ChatPhotoEmpty{}}))

	return []Peer{user, chat, channel}
}

func TestEncodePeer(t *testing.T) {
	for _, p := range testPeers(t) {
		t.Run(p.String(), func(t *testing.T) {
			a := require.New(t)
			p.CreatedAt = time.Unix(p.CreatedAt.Unix(), 0)

			data, err := EncodePeer(p)
			a.NoError(err)
			a.Equal(peerFormatBinary, data[0])

			var got Peer
			a.NoError(DecodePeer(data, &got))
			a.Equal(p.Key, got.Key)
			a.Equal(p.Version, got.Version)
			a.True(p.CreatedAt.Equal(got.CreatedAt))
			a.Equal(p.Metadata, got.Metadata)
			a.Equal(p.User, got.User)
			a.Equal(p.Chat, got.Chat)
			a.Equal(p.Channel, got.Channel)

			// Values stored in the JSON format are still readable.
			legacy, err := json.Marshal(p)
			a.NoError(err)
			var fromJSON Peer
			a.NoError(DecodePeer(legacy, &fromJSON))
			a.Equal(got.Key, fromJSON.Key)
			a.Equal(got.User, fromJSON.User)
			a.Equal(got.Chat, fromJSON.Chat)
			a.Equal(got.Channel, fromJSON.Channel)

			a.Less(len(data), len(legacy), "binary encoding must be smaller")
		})
	}
}

func TestDecodePeerInvalid(t *testing.T) {
	a := require.New(t)
	p := testPeers(t)[0]

	outdated := p
	outdated.Version = LatestVersion - 1
	data, err := EncodePeer(outdated)
	a.NoError(err)
	var got Peer
	a.ErrorIs(DecodePeer(data, &got), ErrPeerUnmarshalMustInvalidate)

	data, err = EncodePeer(p)
	a.NoError(err)
	a.Error(DecodePeer(data[:len(data)-3], &got), "truncated")
	a.Error(DecodePeer(nil, &got), "empty")
}

func BenchmarkPeerEncoding(b *testing.B) {
	var p Peer
	p.FromUser(&tg.User{ID: 10, AccessHash: 20, Username: "foo", FirstName: "Foo", LastName: "Bar", Phone: "123"})
	jsonData, err := json.Marshal(p)
	if err != nil {
		b.Fatal(err)
	}
	binData, err := EncodePeer(p)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Encode/JSON", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, err := json.Marshal(p); err != nil {
				b.Fatal(err)
			}
		}
		b.ReportMetric(float64(len(jsonData)), "bytes")
	})
	b.Run("Encode/Binary", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, err := EncodePeer(p); err != nil {
				b.Fatal(err)
			}
		}
		b.ReportMetric(float64(len(binData)), "bytes")
	})
	b.Run("Decode/JSON", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var v Peer
			if err := DecodePeer(jsonData, &v); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("Decode/Binary", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var v Peer
			if err := DecodePeer(binData, &v); err != nil {
				b.Fatal(err)
			}
		}
	})
}
