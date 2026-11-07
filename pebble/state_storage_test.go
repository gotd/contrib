package pebble

import (
	"context"
	"sync"
	"testing"

	pebbledb "github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/updates"
)

func openMem(t *testing.T) *pebbledb.DB {
	t.Helper()
	db, err := pebbledb.Open("pebble.db", &pebbledb.Options{FS: vfs.NewMem()})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestStateStorageCorruptValues(t *testing.T) {
	ctx := context.Background()
	db := openMem(t)
	const userID, channelID = 10, 20
	corrupt := []byte{1, 2, 3}

	require.NoError(t, db.Set(userPairKey(prefixChannelPts, userID, channelID), corrupt, pebbledb.NoSync))
	require.NoError(t, db.Set(userPairKey(prefixChanHash, userID, channelID), corrupt, pebbledb.NoSync))
	require.NoError(t, db.Set(userKey(prefixState, userID), corrupt, pebbledb.NoSync))

	s := NewStateStorage(db)
	_, _, err := s.GetChannelPts(ctx, userID, channelID)
	require.Error(t, err)
	err = s.ForEachChannels(ctx, userID, func(context.Context, int64, int) error { return nil })
	require.Error(t, err)
	_, _, err = s.GetState(ctx, userID)
	require.Error(t, err)

	_, _, err = NewAccessHasher(db).GetChannelAccessHash(ctx, userID, channelID)
	require.Error(t, err)
}

func TestStateStorageConcurrentSetters(t *testing.T) {
	ctx := context.Background()
	s := NewStateStorage(openMem(t)).WithWriteOptions(pebbledb.NoSync)
	const userID = 1
	require.NoError(t, s.SetState(ctx, userID, updates.State{}))

	// Each setter changes its own field: none of them may be lost.
	const n = 100
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 1; i <= n; i++ {
			require.NoError(t, s.SetPts(ctx, userID, i))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 1; i <= n; i++ {
			require.NoError(t, s.SetQts(ctx, userID, i))
		}
	}()
	wg.Wait()

	state, ok, err := s.GetState(ctx, userID)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, n, state.Pts)
	require.Equal(t, n, state.Qts)
}
