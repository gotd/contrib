package redis

import (
	"context"
	"strconv"

	"github.com/go-faster/errors"
	"github.com/go-redis/redis/v8"

	"github.com/gotd/td/telegram/updates"
)

var (
	_ updates.StateStorage        = (*StateStorage)(nil)
	_ updates.ChannelAccessHasher = (*AccessHasher)(nil)
	_ updates.UserAccessHasher    = (*AccessHasher)(nil)
)

// Key prefixes of updates storage, followed by user ID.
//
// They must not start with storage.PeerKeyPrefix ("peer"), which PeerStorage
// scans.
const (
	prefixState        = "updates:state:"
	prefixChannelPts   = "updates:channel_pts:"
	prefixChannelHash  = "updates:channel_hash:"
	prefixUserHash     = "updates:user_hash:"
	fieldPts, fieldQts = "pts", "qts"
	fieldDate          = "date"
	fieldSeq           = "seq"
)

// ErrStateNotFound is returned by StateStorage setters if user state does
// not exist, as updates.StateStorage requires.
var ErrStateNotFound = errors.New("state not found")

// luaSetIfExists sets hash fields only if the hash exists, atomically.
//
// Returns 0 if the hash does not exist.
const luaSetIfExists = `
if redis.call("EXISTS", KEYS[1]) == 0 then
	return 0
end
redis.call("HSET", KEYS[1], unpack(ARGV))
return 1
`

var setIfExists = redis.NewScript(luaSetIfExists) // nolint:gochecknoglobals

func userKey(prefix string, userID int64) string {
	return prefix + strconv.FormatInt(userID, 10)
}

func formatID(id int64) string { return strconv.FormatInt(id, 10) }

// StateStorage is updates.StateStorage implementation using redis.
//
// Common state of a user is stored in a hash, channel pts in another hash
// keyed by channel ID.
type StateStorage struct {
	redis *redis.Client
}

// NewStateStorage creates new updates state storage using redis.
func NewStateStorage(client *redis.Client) *StateStorage {
	return &StateStorage{redis: client}
}

// GetState implements updates.StateStorage.
func (s *StateStorage) GetState(ctx context.Context, userID int64) (updates.State, bool, error) {
	vals, err := s.redis.HMGet(ctx, userKey(prefixState, userID),
		fieldPts, fieldQts, fieldDate, fieldSeq,
	).Result()
	if err != nil {
		return updates.State{}, false, err
	}
	var (
		ints    [4]int
		missing int
	)
	for i, v := range vals {
		if v == nil {
			missing++
			continue
		}
		str, ok := v.(string)
		if !ok {
			return updates.State{}, false, errors.Errorf("unexpected state value %T", v)
		}
		n, err := strconv.Atoi(str)
		if err != nil {
			return updates.State{}, false, errors.Wrap(err, "parse state")
		}
		ints[i] = n
	}
	switch missing {
	case 0:
	case len(vals):
		return updates.State{}, false, nil
	default:
		return updates.State{}, false, errors.New("incomplete state")
	}
	return updates.State{
		Pts:  ints[0],
		Qts:  ints[1],
		Date: ints[2],
		Seq:  ints[3],
	}, true, nil
}

// SetState implements updates.StateStorage.
func (s *StateStorage) SetState(ctx context.Context, userID int64, state updates.State) error {
	return s.redis.HSet(ctx, userKey(prefixState, userID),
		fieldPts, state.Pts,
		fieldQts, state.Qts,
		fieldDate, state.Date,
		fieldSeq, state.Seq,
	).Err()
}

func (s *StateStorage) update(ctx context.Context, userID int64, fields ...interface{}) error {
	set, err := setIfExists.Run(ctx, s.redis, []string{userKey(prefixState, userID)}, fields...).Int()
	if err != nil {
		return err
	}
	if set == 0 {
		return ErrStateNotFound
	}
	return nil
}

// SetPts implements updates.StateStorage.
func (s *StateStorage) SetPts(ctx context.Context, userID int64, pts int) error {
	return s.update(ctx, userID, fieldPts, pts)
}

// SetQts implements updates.StateStorage.
func (s *StateStorage) SetQts(ctx context.Context, userID int64, qts int) error {
	return s.update(ctx, userID, fieldQts, qts)
}

// SetDate implements updates.StateStorage.
func (s *StateStorage) SetDate(ctx context.Context, userID int64, date int) error {
	return s.update(ctx, userID, fieldDate, date)
}

// SetSeq implements updates.StateStorage.
func (s *StateStorage) SetSeq(ctx context.Context, userID int64, seq int) error {
	return s.update(ctx, userID, fieldSeq, seq)
}

// SetDateSeq implements updates.StateStorage.
func (s *StateStorage) SetDateSeq(ctx context.Context, userID int64, date, seq int) error {
	return s.update(ctx, userID, fieldDate, date, fieldSeq, seq)
}

// GetChannelPts implements updates.StateStorage.
func (s *StateStorage) GetChannelPts(ctx context.Context, userID, channelID int64) (pts int, found bool, err error) {
	v, err := s.redis.HGet(ctx, userKey(prefixChannelPts, userID), formatID(channelID)).Result()
	if errors.Is(err, redis.Nil) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	pts, err = strconv.Atoi(v)
	if err != nil {
		return 0, false, errors.Wrapf(err, "channel %d pts", channelID)
	}
	return pts, true, nil
}

// SetChannelPts implements updates.StateStorage.
func (s *StateStorage) SetChannelPts(ctx context.Context, userID, channelID int64, pts int) error {
	return s.redis.HSet(ctx, userKey(prefixChannelPts, userID), formatID(channelID), pts).Err()
}

// ForEachChannels implements updates.StateStorage.
func (s *StateStorage) ForEachChannels(
	ctx context.Context,
	userID int64,
	f func(ctx context.Context, channelID int64, pts int) error,
) error {
	all, err := s.redis.HGetAll(ctx, userKey(prefixChannelPts, userID)).Result()
	if err != nil {
		return err
	}
	for k, v := range all {
		channelID, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			return errors.Wrap(err, "channel key")
		}
		pts, err := strconv.Atoi(v)
		if err != nil {
			return errors.Wrapf(err, "channel %d pts", channelID)
		}
		if err := f(ctx, channelID, pts); err != nil {
			return err
		}
	}
	return nil
}

// AccessHasher is updates.ChannelAccessHasher and updates.UserAccessHasher
// implementation using redis.
type AccessHasher struct {
	redis *redis.Client
}

// NewAccessHasher creates new access hasher using redis.
func NewAccessHasher(client *redis.Client) *AccessHasher {
	return &AccessHasher{redis: client}
}

func (s *AccessHasher) set(ctx context.Context, prefix string, userID, id, hash int64) error {
	return s.redis.HSet(ctx, userKey(prefix, userID), formatID(id), formatID(hash)).Err()
}

func (s *AccessHasher) get(ctx context.Context, prefix string, userID, id int64) (hash int64, found bool, err error) {
	v, err := s.redis.HGet(ctx, userKey(prefix, userID), formatID(id)).Result()
	if errors.Is(err, redis.Nil) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	hash, err = strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false, errors.Wrapf(err, "access hash of %d", id)
	}
	return hash, true, nil
}

// SetChannelAccessHash implements updates.ChannelAccessHasher.
func (s *AccessHasher) SetChannelAccessHash(ctx context.Context, userID, channelID, accessHash int64) error {
	return s.set(ctx, prefixChannelHash, userID, channelID, accessHash)
}

// GetChannelAccessHash implements updates.ChannelAccessHasher.
func (s *AccessHasher) GetChannelAccessHash(ctx context.Context, userID, channelID int64) (accessHash int64, found bool, err error) {
	return s.get(ctx, prefixChannelHash, userID, channelID)
}

// SetUserAccessHash implements updates.UserAccessHasher.
func (s *AccessHasher) SetUserAccessHash(ctx context.Context, userID, targetUserID, accessHash int64) error {
	return s.set(ctx, prefixUserHash, userID, targetUserID, accessHash)
}

// GetUserAccessHash implements updates.UserAccessHasher.
func (s *AccessHasher) GetUserAccessHash(ctx context.Context, userID, targetUserID int64) (accessHash int64, found bool, err error) {
	return s.get(ctx, prefixUserHash, userID, targetUserID)
}
