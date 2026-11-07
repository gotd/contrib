package bbolt

import (
	"context"
	"path"
	"testing"
	"time"

	"github.com/gotd/td/telegram/updates"

	"github.com/stretchr/testify/require"
	bboltdb "go.etcd.io/bbolt"
)

func TestState(t *testing.T) {
	db, err := bboltdb.Open(path.Join(t.TempDir(), "bbolt.db"), 0666, &bboltdb.Options{}) // nolint:gocritic
	require.NoError(t, err)

	state := NewStateStorage(db)

	ctx := context.Background()
	cb := func(ctx context.Context, channelID int64, pts int) error {
		return nil
	}
	require.NoError(t, state.ForEachChannels(ctx, 0, cb))
	require.NoError(t, db.Close())
}

func TestStateForEachChannelsReadOnly(t *testing.T) {
	db, err := bboltdb.Open(path.Join(t.TempDir(), "bbolt.db"), 0666, &bboltdb.Options{}) // nolint:gocritic
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	ctx := context.Background()
	state := NewStateStorage(db)
	hasher := NewAccessHasher(db)
	require.NoError(t, state.SetState(ctx, 1, updates.State{Pts: 1}))
	require.NoError(t, state.SetChannelPts(ctx, 1, 10, 100))
	require.NoError(t, hasher.SetChannelAccessHash(ctx, 1, 10, 42))

	var seen int
	require.NoError(t, state.ForEachChannels(ctx, 1, func(ctx context.Context, channelID int64, pts int) error {
		seen++
		require.Equal(t, int64(10), channelID)
		require.Equal(t, 100, pts)

		// The updates manager reads access hashes from the same db while
		// iterating (loadChannels).
		hash, found, err := hasher.GetChannelAccessHash(ctx, 1, channelID)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, int64(42), hash)

		// Other writers must not be blocked for the whole iteration.
		done := make(chan error, 1)
		go func() { done <- hasher.SetUserAccessHash(ctx, 1, 7, 7) }()
		select {
		case err := <-done:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent write blocked by ForEachChannels")
		}
		return nil
	}))
	require.Equal(t, 1, seen)

	// Unknown user: no buckets are created, nothing is iterated.
	require.NoError(t, state.ForEachChannels(ctx, 2, func(context.Context, int64, int) error {
		t.Fatal("unexpected channel")
		return nil
	}))
}
