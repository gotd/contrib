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
	rerr := h.next.Handle(ctx, u)

	updates, ok := u.(updatesWithPeers)
	if !ok {
		return rerr
	}

	for _, chat := range updates.GetChats() {
		if value := (Peer{}); value.FromChat(chat) {
			multierr.AppendInto(&rerr, h.storage.Add(ctx, value))
		}
	}

	for _, user := range updates.GetUsers() {
		if value := (Peer{}); value.FromUser(user) {
			multierr.AppendInto(&rerr, h.storage.Add(ctx, value))
		}
	}

	return rerr
}

// UpdateHook creates update hook, to collect peer data from updates.
func UpdateHook(next telegram.UpdateHandler, storage PeerStorage) telegram.UpdateHandler {
	return updateHook{
		next:    next,
		storage: storage,
	}
}
