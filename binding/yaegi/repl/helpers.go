package main

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func helpers(ctx context.Context, client *telegram.Client) map[string]reflect.Value {
	invoker := tg.NewClient(client)
	return map[string]reflect.Value{
		"Client": reflect.ValueOf(client),
		"RPC":    reflect.ValueOf(invoker),
		"Ctx":    reflect.ValueOf(ctx),
		"Send":   reflect.ValueOf(send(ctx, invoker, client)),
	}
}

func resolve(ctx context.Context, invoker *tg.Client, domain string) (tg.InputPeerClass, error) {
	peer, err := invoker.ContactsResolveUsername(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("resolve %q failed: %w", domain, err)
	}

	users := make(map[int]*tg.User, len(peer.Users))
	for _, class := range peer.Users {
		user, ok := class.(*tg.User)
		if !ok {
			continue
		}
		users[user.ID] = user
	}

	chats := make(map[int]*tg.Chat, len(peer.Chats))
	channels := make(map[int]*tg.Channel, len(peer.Chats))
	for _, class := range peer.Chats {
		switch chat := class.(type) {
		case *tg.Chat:
			chats[chat.ID] = chat
		case *tg.Channel:
			channels[chat.ID] = chat
		}
	}

	switch v := peer.Peer.(type) {
	case *tg.PeerUser: // peerUser#9db1bc6d
		p, ok := users[v.UserID]
		if !ok {
			return nil, fmt.Errorf("invalid user ID %d received", v.UserID)
		}

		return &tg.InputPeerUser{
			UserID:     p.ID,
			AccessHash: p.AccessHash,
		}, nil
	case *tg.PeerChat: // peerChat#bad0e5bb
		p, ok := chats[v.ChatID]
		if !ok {
			return nil, fmt.Errorf("invalid chat ID %d received", v.ChatID)
		}

		return &tg.InputPeerChat{
			ChatID: p.ID,
		}, nil
	case *tg.PeerChannel: // peerChannel#bddde532
		p, ok := channels[v.ChannelID]
		if !ok {
			return nil, fmt.Errorf("invalid channel ID %d received", v.ChannelID)
		}

		return &tg.InputPeerChannel{
			ChannelID:  p.ID,
			AccessHash: p.AccessHash,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected type %T", peer.Peer)
	}
}

func send(ctx context.Context, invoker *tg.Client, client *telegram.Client) func(msg, to string) error {
	return func(msg, to string) error {
		peer, err := resolve(ctx, invoker, to)
		if err != nil {
			return err
		}

		return client.SendMessage(ctx, &tg.MessagesSendMessageRequest{
			Peer:    peer,
			Message: msg,
		})
	}
}
