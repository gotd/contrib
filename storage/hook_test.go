package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

type testHandler struct {
	returnErr error
}

func (t testHandler) Handle(ctx context.Context, u *tg.Updates) error {
	return t.returnErr
}

func (t testHandler) HandleShort(ctx context.Context, u *tg.UpdateShort) error {
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
			returnErr: xerrors.New("testErr"),
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
