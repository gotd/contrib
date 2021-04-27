package storage

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/query/channels/participants"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
)

// PeerCollector is a simple helper to collect peers from different sources.
type PeerCollector struct {
	storage PeerStorage
}

// Dialogs collects peers from dialog iterator.
func (c PeerCollector) Dialogs(ctx context.Context, iter *dialogs.Iterator) error {
	for iter.Next(ctx) {
		var (
			p     Peer
			value = iter.Value()
		)
		if err := p.FromInputPeer(value.Peer); err != nil {
			return err
		}
		if err := c.storage.Add(ctx, p); err != nil {
			return xerrors.Errorf("add: %w", err)
		}
	}

	return iter.Err()
}

// Participants collects peers from participants iterator.
func (c PeerCollector) Participants(ctx context.Context, iter *participants.Iterator) error {
	for iter.Next(ctx) {
		var (
			p     Peer
			value = iter.Value()
		)
		user, ok := value.User()
		if !ok {
			continue
		}

		if !p.FromUser(user) {
			continue
		}
		if err := c.storage.Add(ctx, p); err != nil {
			return xerrors.Errorf("add: %w", err)
		}
	}

	return iter.Err()
}

// Contacts collects peers from contacts iterator.
func (c PeerCollector) Contacts(ctx context.Context, contacts *tg.ContactsContacts) error {
	for _, user := range contacts.Users {
		var p Peer
		if !p.FromUser(user) {
			continue
		}
		if err := c.storage.Add(ctx, p); err != nil {
			return xerrors.Errorf("add: %w", err)
		}
	}
	return nil
}

// CollectPeers creates new PeerCollector.
func CollectPeers(storage PeerStorage) PeerCollector {
	return PeerCollector{storage: storage}
}
