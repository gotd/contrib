package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/query/channels/participants"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
)

func testUser() *tg.User {
	return &tg.User{
		ID:         10,
		AccessHash: 10,
		FirstName:  "Жак",
		LastName:   "Фреско",
		Username:   "zagadka1337",
	}
}

type dialogQuery struct {
	result *tg.MessagesDialogs
}

func (d dialogQuery) Query(ctx context.Context, req dialogs.Request) (tg.MessagesDialogsClass, error) {
	return d.result, nil
}

func TestPeerCollector_Dialogs(t *testing.T) {
	a := require.New(t)
	mem := newMemStorage()
	collector := CollectPeers(mem)
	ctx := context.Background()

	user := testUser()
	iter := dialogs.NewIterator(dialogQuery{
		result: &tg.MessagesDialogs{
			Dialogs: []tg.DialogClass{
				&tg.Dialog{
					Pinned:     false,
					UnreadMark: false,
					Peer:       &tg.PeerUser{UserID: 10},
					TopMessage: 1,
				},
			},
			Messages: []tg.MessageClass{
				&tg.Message{
					ID:      1,
					PeerID:  &tg.PeerUser{UserID: 10},
					Message: "бебебе с бабаба",
				},
			},
			Users: []tg.UserClass{user},
		},
	}, 1)
	a.NoError(collector.Dialogs(ctx, iter))

	p, err := mem.Resolve(ctx, user.Username)
	a.NoError(err)
	a.NotNil(p.User)
	a.Equal(user.FirstName, p.User.FirstName)
}

func testChannel() *tg.Channel {
	return &tg.Channel{
		ID:         20,
		AccessHash: 20,
		Title:      "Загадка",
		Username:   "zagadka_channel",
	}
}

// td v0.161 added DialogFolder and DialogCommunity under the DialogClass
// interface and dropped GetPeer from it. The collector must read peers from
// dialogs that have one (Dialog, DialogFolder) and skip the peerless
// DialogCommunity without failing.
func TestPeerCollector_DialogsMixedTypes(t *testing.T) {
	a := require.New(t)
	mem := newMemStorage()
	collector := CollectPeers(mem)
	ctx := context.Background()

	user := testUser()
	channel := testChannel()
	iter := dialogs.NewIterator(dialogQuery{
		result: &tg.MessagesDialogs{
			Dialogs: []tg.DialogClass{
				&tg.Dialog{Peer: &tg.PeerUser{UserID: user.ID}, TopMessage: 1},
				&tg.DialogFolder{Peer: &tg.PeerChannel{ChannelID: channel.ID}, TopMessage: 2},
				&tg.DialogCommunity{CommunityID: 999}, // peerless: must be skipped, not panic
			},
			Users: []tg.UserClass{user},
			Chats: []tg.ChatClass{channel},
		},
	}, 10)
	a.NoError(collector.Dialogs(ctx, iter))

	pu, err := mem.Resolve(ctx, user.Username)
	a.NoError(err)
	a.NotNil(pu.User)
	a.Equal(user.FirstName, pu.User.FirstName)

	pc, err := mem.Resolve(ctx, channel.Username)
	a.NoError(err)
	a.NotNil(pc.Channel)
	a.Equal(channel.Title, pc.Channel.Title)
}

type participantsQuery struct {
	result *tg.ChannelsChannelParticipants
}

func (p *participantsQuery) Query(ctx context.Context, req participants.Request) (tg.ChannelsChannelParticipantsClass, error) {
	p.result.Participants = p.result.Participants[req.Offset:]
	return p.result, nil
}

func TestPeerCollector_Participants(t *testing.T) {
	a := require.New(t)
	mem := newMemStorage()
	collector := CollectPeers(mem)
	ctx := context.Background()

	user := testUser()
	iter := participants.NewIterator(&participantsQuery{
		result: &tg.ChannelsChannelParticipants{
			Count: 1,
			Participants: []tg.ChannelParticipantClass{
				&tg.ChannelParticipantCreator{
					UserID:      user.ID,
					AdminRights: tg.ChatAdminRights{},
					Rank:        "фреска",
				},
			},
			Users: []tg.UserClass{user},
		},
	}, 1)
	a.NoError(collector.Participants(ctx, iter))

	p, err := mem.Resolve(ctx, "zagadka1337")
	a.NoError(err)
	a.NotNil(p.User)
	a.Equal("Жак", p.User.FirstName)
}

func TestPeerCollector_Contacts(t *testing.T) {
	a := require.New(t)
	mem := newMemStorage()
	collector := CollectPeers(mem)
	ctx := context.Background()

	user := testUser()
	a.NoError(collector.Contacts(ctx, &tg.ContactsContacts{
		Users: []tg.UserClass{user},
	}))

	p, err := mem.Resolve(ctx, "zagadka1337")
	a.NoError(err)
	a.NotNil(p.User)
	a.Equal("Жак", p.User.FirstName)
}
