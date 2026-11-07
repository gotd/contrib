package storage

import (
	"context"

	"go.uber.org/multierr"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type updateHook struct {
	next    telegram.UpdateHandler
	storage PeerStorage
}

type updatesWithPeers interface {
	GetUsers() []tg.UserClass
	GetChats() []tg.ChatClass
	tg.UpdatesClass
}

func (h updateHook) Handle(ctx context.Context, u tg.UpdatesClass) error {
	updates, ok := u.(updatesWithPeers)
	if !ok {
		return h.next.Handle(ctx, u)
	}

	chats, users := updates.GetChats(), updates.GetUsers()
	values := make([]Peer, 0, len(chats)+len(users))
	for _, chat := range chats {
		if value := (Peer{}); value.FromChat(chat) {
			values = append(values, value)
		}
	}
	for _, user := range users {
		if value := (Peer{}); value.FromUser(user) {
			values = append(values, value)
		}
	}

	rerr := addMany(ctx, h.storage, values)
	return multierr.Append(rerr, h.next.Handle(ctx, u))
}

// UpdateHook creates update hook, to collect peer data from updates.
func UpdateHook(next telegram.UpdateHandler, storage PeerStorage) telegram.UpdateHandler {
	return updateHook{
		next:    next,
		storage: storage,
	}
}
