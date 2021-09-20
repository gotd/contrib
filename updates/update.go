package updates

import "github.com/gotd/td/tg"

type update struct {
	Value interface{}
	Ents  *Entities
}

// Entities contains Telegram's entities which comes with updates.
type Entities struct {
	Users    map[int]*tg.User
	Chats    map[int]*tg.Chat
	Channels map[int]*tg.Channel
}

func newEntities() *Entities {
	return &Entities{
		Users:    map[int]*tg.User{},
		Chats:    map[int64]*tg.Chat{},
		Channels: map[int]*tg.Channel{},
	}
}

func (e *Entities) merge(from *Entities) {
	if from == nil {
		return
	}

	for userID, user := range from.Users {
		e.Users[userID] = user
	}

	for chanID, chat := range from.Chats {
		e.Chats[chanID] = chat
	}

	for channelID, channel := range from.Channels {
		e.Channels[channelID] = channel
	}
}

// AsUsers returns user entities as tg.UserClass slice.
func (e *Entities) AsUsers() []tg.UserClass {
	if e == nil {
		return nil
	}

	var users []tg.UserClass
	for _, user := range e.Users {
		users = append(users, user)
	}
	return users
}

// AsChats returns chat entities as tg.ChatClass slice.
func (e *Entities) AsChats() []tg.ChatClass {
	if e == nil {
		return nil
	}

	var chats []tg.ChatClass
	for _, chat := range e.Chats {
		chats = append(chats, chat)
	}
	for _, channel := range e.Channels {
		chats = append(chats, channel)
	}
	return chats
}
