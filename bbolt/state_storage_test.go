package bbolt

import (
	"context"
	"path"
	"testing"

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
