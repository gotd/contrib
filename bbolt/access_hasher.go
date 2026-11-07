package bbolt

import (
	"context"

	bolt "go.etcd.io/bbolt"

	"github.com/gotd/td/telegram/updates"
)

var (
	_ updates.ChannelAccessHasher = (*AccessHasher)(nil)
	_ updates.UserAccessHasher    = (*AccessHasher)(nil)
)

// AccessHasher is updates.ChannelAccessHasher and updates.UserAccessHasher
// implementation using bbolt.
//
// It uses the same per-user buckets as State, so both can share one db.
type AccessHasher struct {
	db *bolt.DB
}

// NewAccessHasher creates new access hasher over bbolt.
//
// Caller is responsible for db.Close() invocation.
func NewAccessHasher(db *bolt.DB) *AccessHasher { return &AccessHasher{db: db} }

var (
	bucketChannelHashes = []byte("channel_hashes")
	bucketUserHashes    = []byte("user_hashes")
)

func (s *AccessHasher) set(bucket []byte, userID, id, hash int64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		user, err := tx.CreateBucketIfNotExists(i642b(userID))
		if err != nil {
			return err
		}
		b, err := user.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return b.Put(i642b(id), i642b(hash))
	})
}

func (s *AccessHasher) get(bucket []byte, userID, id int64) (hash int64, found bool, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		user := tx.Bucket(i642b(userID))
		if user == nil {
			return nil
		}
		b := user.Bucket(bucket)
		if b == nil {
			return nil
		}
		v := b.Get(i642b(id))
		if v == nil {
			return nil
		}
		hash, found = b2i64(v), true
		return nil
	})
	return hash, found, err
}

// SetChannelAccessHash implements updates.ChannelAccessHasher.
func (s *AccessHasher) SetChannelAccessHash(_ context.Context, userID, channelID, accessHash int64) error {
	return s.set(bucketChannelHashes, userID, channelID, accessHash)
}

// GetChannelAccessHash implements updates.ChannelAccessHasher.
func (s *AccessHasher) GetChannelAccessHash(_ context.Context, userID, channelID int64) (int64, bool, error) {
	return s.get(bucketChannelHashes, userID, channelID)
}

// SetUserAccessHash implements updates.UserAccessHasher.
func (s *AccessHasher) SetUserAccessHash(_ context.Context, userID, targetUserID, accessHash int64) error {
	return s.set(bucketUserHashes, userID, targetUserID, accessHash)
}

// GetUserAccessHash implements updates.UserAccessHasher.
func (s *AccessHasher) GetUserAccessHash(_ context.Context, userID, targetUserID int64) (int64, bool, error) {
	return s.get(bucketUserHashes, userID, targetUserID)
}
