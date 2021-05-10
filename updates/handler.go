package updates

import "github.com/gotd/td/tg"

// DiffUpdate contains updates received by getDifference.
type DiffUpdate struct {
	NewMessages          []tg.MessageClass
	NewEncryptedMessages []tg.EncryptedMessageClass
	OtherUpdates         []tg.UpdateClass
	Users                []tg.UserClass
	Chats                []tg.ChatClass

	Pending     []tg.UpdateClass
	PendingEnts *Entities
}

// Updates contains Telegram updates.
type Updates struct {
	Updates []tg.UpdateClass
	Ents    *Entities
}

// Handler interface.
type Handler interface {
	HandleDiff(DiffUpdate) error
	HandleUpdates(Updates) error
	ChannelTooLong(channelID int)
}
