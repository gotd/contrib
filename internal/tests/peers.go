package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gotd/td/telegram/peers"
)

// TestPeersStorage runs different tests for given peers.Storage
// implementation.
func TestPeersStorage(t *testing.T, s peers.Storage) {
	ctx := context.Background()
	a := require.New(t)

	user := peers.Key{Prefix: "users_", ID: 1}
	channel := peers.Key{Prefix: "channel_", ID: 1}

	_, found, err := s.Find(ctx, user)
	a.NoError(err)
	a.False(found)

	a.NoError(s.Save(ctx, user, peers.Value{AccessHash: 10}))
	v, found, err := s.Find(ctx, user)
	a.NoError(err)
	a.True(found)
	a.Equal(peers.Value{AccessHash: 10}, v)

	_, found, err = s.Find(ctx, channel)
	a.NoError(err)
	a.False(found, "keys with same ID and different prefix are separate")

	a.NoError(s.Save(ctx, channel, peers.Value{AccessHash: -20}))
	a.NoError(s.Save(ctx, user, peers.Value{AccessHash: 11}))
	v, _, err = s.Find(ctx, user)
	a.NoError(err)
	a.Equal(peers.Value{AccessHash: 11}, v, "value is overwritten")
	v, _, err = s.Find(ctx, channel)
	a.NoError(err)
	a.Equal(peers.Value{AccessHash: -20}, v)

	_, _, found, err = s.FindPhone(ctx, "phone")
	a.NoError(err)
	a.False(found)

	a.NoError(s.SavePhone(ctx, "phone", user))
	k, v, found, err := s.FindPhone(ctx, "phone")
	a.NoError(err)
	a.True(found)
	a.Equal(user, k)
	a.Equal(peers.Value{AccessHash: 11}, v)

	// Phone of a key without value is not found, as in peers.InmemoryStorage.
	a.NoError(s.SavePhone(ctx, "other", peers.Key{Prefix: "users_", ID: 2}))
	_, _, found, err = s.FindPhone(ctx, "other")
	a.NoError(err)
	a.False(found)

	hash, err := s.GetContactsHash(ctx)
	a.NoError(err)
	a.Zero(hash)
	a.NoError(s.SaveContactsHash(ctx, 1))
	a.NoError(s.SaveContactsHash(ctx, -5))
	hash, err = s.GetContactsHash(ctx)
	a.NoError(err)
	a.Equal(int64(-5), hash)
}
