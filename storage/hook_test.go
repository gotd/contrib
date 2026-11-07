package storage

import (
	"context"
	"testing"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tg"
)

type testHandler struct {
	returnErr error
}

func (t testHandler) Handle(ctx context.Context, u tg.UpdatesClass) error {
	return t.returnErr
}

func TestUpdateHook(t *testing.T) {
	ctx := context.Background()
	testData := &tg.Updates{
		Chats: []tg.ChatClass{
			&tg.Channel{
				ID:         10,
				AccessHash: 10,
				Username:   "channel",
			},
			&tg.ChannelForbidden{
				ID:         11,
				AccessHash: 11,
			},
		},
		Users: []tg.UserClass{
			&tg.User{
				ID:         10,
				AccessHash: 10,
				Username:   "username",
			},
		},
	}

	t.Run("Good", func(t *testing.T) {
		a := require.New(t)
		storage := newMemStorage()
		h := UpdateHook(testHandler{}, storage)

		a.NoError(h.Handle(ctx, testData))

		p, err := storage.Resolve(ctx, "channel")
		a.NoError(err)
		a.NotNil(p.Channel)

		p, err = storage.Resolve(ctx, "username")
		a.NoError(err)
		a.NotNil(p.User)
	})

	t.Run("Error", func(t *testing.T) {
		a := require.New(t)
		storage := newMemStorage()
		h := UpdateHook(testHandler{
			returnErr: errors.New("testErr"),
		}, storage)

		a.Error(h.Handle(ctx, testData))

		p, err := storage.Resolve(ctx, "channel")
		a.NoError(err)
		a.NotNil(p.Channel)

		p, err = storage.Resolve(ctx, "username")
		a.NoError(err)
		a.NotNil(p.User)
	})
}

func TestUpdateHookMinUser(t *testing.T) {
	ctx := context.Background()
	a := require.New(t)
	storage := newMemStorage()
	h := UpdateHook(testHandler{}, storage)

	a.NoError(h.Handle(ctx, &tg.Updates{Users: []tg.UserClass{
		&tg.User{ID: 42, AccessHash: 1111, Username: "alice", Phone: "123"},
	}}))
	// Min user (e.g. from a group message) must not overwrite full one.
	a.NoError(h.Handle(ctx, &tg.Updates{Users: []tg.UserClass{
		&tg.User{ID: 42, Min: true, Username: "alice"},
	}}))

	p, err := storage.Resolve(ctx, "alice")
	a.NoError(err)
	a.False(p.User.Min)
	a.Equal(int64(1111), p.Key.AccessHash)
	a.Equal("123", p.User.Phone)

	// Min user alone is not stored.
	a.NoError(h.Handle(ctx, &tg.Updates{Users: []tg.UserClass{
		&tg.User{ID: 43, Min: true, Username: "bob"},
	}}))
	_, err = storage.Resolve(ctx, "bob")
	a.ErrorIs(err, ErrPeerNotFound)
}
