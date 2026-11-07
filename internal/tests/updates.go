package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/updates"
)

// AccessHasher is a storage of channel and user access hashes.
type AccessHasher interface {
	updates.ChannelAccessHasher
	updates.UserAccessHasher
}

// TestStateStorage runs different tests for given updates state storage
// implementation.
func TestStateStorage(t *testing.T, s updates.StateStorage) {
	ctx := context.Background()
	a := require.New(t)
	const userID, otherUserID = 10, 11

	_, found, err := s.GetState(ctx, userID)
	a.NoError(err)
	a.False(found)
	a.Error(s.SetPts(ctx, userID, 1), "setter without state must fail")

	state := updates.State{Pts: 1, Qts: 2, Date: 3, Seq: 4}
	a.NoError(s.SetState(ctx, userID, state))
	got, found, err := s.GetState(ctx, userID)
	a.NoError(err)
	a.True(found)
	a.Equal(state, got)

	a.NoError(s.SetPts(ctx, userID, 10))
	a.NoError(s.SetQts(ctx, userID, 20))
	a.NoError(s.SetDate(ctx, userID, 30))
	a.NoError(s.SetSeq(ctx, userID, 40))
	got, _, err = s.GetState(ctx, userID)
	a.NoError(err)
	a.Equal(updates.State{Pts: 10, Qts: 20, Date: 30, Seq: 40}, got)
	a.NoError(s.SetDateSeq(ctx, userID, 50, 60))
	got, _, err = s.GetState(ctx, userID)
	a.NoError(err)
	a.Equal(updates.State{Pts: 10, Qts: 20, Date: 50, Seq: 60}, got)

	_, found, err = s.GetChannelPts(ctx, userID, 100)
	a.NoError(err)
	a.False(found)
	a.NoError(s.SetChannelPts(ctx, userID, 100, 1000))
	a.NoError(s.SetChannelPts(ctx, userID, 200, 2000))
	a.NoError(s.SetChannelPts(ctx, otherUserID, 300, 3000))
	pts, found, err := s.GetChannelPts(ctx, userID, 100)
	a.NoError(err)
	a.True(found)
	a.Equal(1000, pts)

	channels := map[int64]int{}
	a.NoError(s.ForEachChannels(ctx, userID, func(ctx context.Context, channelID int64, pts int) error {
		channels[channelID] = pts
		return nil
	}))
	a.Equal(map[int64]int{100: 1000, 200: 2000}, channels)
	a.NoError(s.ForEachChannels(ctx, 12345, func(context.Context, int64, int) error {
		a.Fail("unexpected channel")
		return nil
	}))
}

// TestAccessHasher runs different tests for given access hasher
// implementation.
func TestAccessHasher(t *testing.T, h AccessHasher) {
	ctx := context.Background()
	a := require.New(t)
	const userID, otherUserID = 10, 11

	_, found, err := h.GetChannelAccessHash(ctx, userID, 100)
	a.NoError(err)
	a.False(found)
	a.NoError(h.SetChannelAccessHash(ctx, userID, 100, -42))
	a.NoError(h.SetChannelAccessHash(ctx, userID, 100, -42))
	hash, found, err := h.GetChannelAccessHash(ctx, userID, 100)
	a.NoError(err)
	a.True(found)
	a.Equal(int64(-42), hash)
	_, found, err = h.GetChannelAccessHash(ctx, otherUserID, 100)
	a.NoError(err)
	a.False(found)

	_, found, err = h.GetUserAccessHash(ctx, userID, 100)
	a.NoError(err)
	a.False(found, "user and channel hashes are separate")
	a.NoError(h.SetUserAccessHash(ctx, userID, 100, 7))
	a.NoError(h.SetUserAccessHash(ctx, userID, 100, 8))
	hash, found, err = h.GetUserAccessHash(ctx, userID, 100)
	a.NoError(err)
	a.True(found)
	a.Equal(int64(8), hash)
}
