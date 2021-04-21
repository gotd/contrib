package storage

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"go.uber.org/multierr"
)

type updateHook struct {
	next    telegram.UpdateHandler
	storage PeerStorage
}

func (h updateHook) Handle(ctx context.Context, u *tg.Updates) error {
	rerr := h.next.Handle(ctx, u)

	for _, chat := range u.Chats {
		if value := (Peer{}); value.FromChat(chat) {
			multierr.AppendInto(&rerr, h.storage.Add(ctx, value))
		}
	}

	for _, user := range u.Users {
		if value := (Peer{}); value.FromUser(user) {
			multierr.AppendInto(&rerr, h.storage.Add(ctx, value))
		}
	}

	return rerr
}

func (h updateHook) HandleShort(ctx context.Context, u *tg.UpdateShort) error {
	// Short does not contain peer data.
	return h.next.HandleShort(ctx, u)
}

// UpdateHook creates update hook, to collect peer data from updates.
func UpdateHook(next telegram.UpdateHandler, storage PeerStorage) telegram.UpdateHandler {
	return updateHook{
		next:    next,
		storage: storage,
	}
}
