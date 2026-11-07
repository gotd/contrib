package pebble

import (
	"bytes"
	"context"
	"encoding/binary"

	"github.com/cockroachdb/pebble"
	"github.com/go-faster/errors"
	"go.uber.org/multierr"

	"github.com/gotd/td/telegram/updates"
)

var (
	_ updates.StateStorage        = (*StateStorage)(nil)
	_ updates.ChannelAccessHasher = (*AccessHasher)(nil)
	_ updates.UserAccessHasher    = (*AccessHasher)(nil)
)

// Key prefixes of updates storage.
const (
	prefixState      = "updates_state_"
	prefixChannelPts = "updates_channel_pts_"
	prefixChanHash   = "updates_channel_hash_"
	prefixUserHash   = "updates_user_hash_"
)

// ErrStateNotFound is returned by StateStorage setters if user state does
// not exist, as updates.StateStorage requires.
var ErrStateNotFound = errors.New("state not found")

func userKey(prefix string, userID int64) []byte {
	b := make([]byte, 0, len(prefix)+8)
	b = append(b, prefix...)
	return binary.BigEndian.AppendUint64(b, uint64(userID))
}

func userPairKey(prefix string, userID, id int64) []byte {
	return binary.BigEndian.AppendUint64(userKey(prefix, userID), uint64(id))
}

func encodeInts(v ...int) []byte {
	b := make([]byte, 0, 8*len(v))
	for _, x := range v {
		b = binary.BigEndian.AppendUint64(b, uint64(int64(x)))
	}
	return b
}

func decodeInt(b []byte, i int) int {
	return int(int64(binary.BigEndian.Uint64(b[8*i:])))
}

// get returns copy of value by key.
func get(db *pebble.DB, key []byte) ([]byte, bool, error) {
	v, closer, err := db.Get(key)
	if errors.Is(err, pebble.ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	r := bytes.Clone(v)
	if err := closer.Close(); err != nil {
		return nil, false, err
	}
	return r, true, nil
}

// StateStorage is updates.StateStorage implementation using pebble.
type StateStorage struct {
	db        *pebble.DB
	writeOpts *pebble.WriteOptions
}

// NewStateStorage creates new updates state storage using pebble.
//
// Caller is responsible for db.Close() invocation.
func NewStateStorage(db *pebble.DB) *StateStorage {
	return &StateStorage{db: db}
}

// WithWriteOptions sets pebble's write options for write operations.
//
// Default is nil, which pebble treats as pebble.Sync.
func (s *StateStorage) WithWriteOptions(writeOpts *pebble.WriteOptions) *StateStorage {
	s.writeOpts = writeOpts
	return s
}

// GetState implements updates.StateStorage.
func (s *StateStorage) GetState(_ context.Context, userID int64) (updates.State, bool, error) {
	v, ok, err := get(s.db, userKey(prefixState, userID))
	if err != nil || !ok {
		return updates.State{}, false, err
	}
	if len(v) != 4*8 {
		return updates.State{}, false, errors.Errorf("invalid state length %d", len(v))
	}
	return updates.State{
		Pts:  decodeInt(v, 0),
		Qts:  decodeInt(v, 1),
		Date: decodeInt(v, 2),
		Seq:  decodeInt(v, 3),
	}, true, nil
}

// SetState implements updates.StateStorage.
func (s *StateStorage) SetState(_ context.Context, userID int64, state updates.State) error {
	return s.db.Set(userKey(prefixState, userID), encodeInts(state.Pts, state.Qts, state.Date, state.Seq), s.writeOpts)
}

func (s *StateStorage) update(ctx context.Context, userID int64, f func(state *updates.State)) error {
	state, ok, err := s.GetState(ctx, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrStateNotFound
	}
	f(&state)
	return s.SetState(ctx, userID, state)
}

// SetPts implements updates.StateStorage.
func (s *StateStorage) SetPts(ctx context.Context, userID int64, pts int) error {
	return s.update(ctx, userID, func(state *updates.State) { state.Pts = pts })
}

// SetQts implements updates.StateStorage.
func (s *StateStorage) SetQts(ctx context.Context, userID int64, qts int) error {
	return s.update(ctx, userID, func(state *updates.State) { state.Qts = qts })
}

// SetDate implements updates.StateStorage.
func (s *StateStorage) SetDate(ctx context.Context, userID int64, date int) error {
	return s.update(ctx, userID, func(state *updates.State) { state.Date = date })
}

// SetSeq implements updates.StateStorage.
func (s *StateStorage) SetSeq(ctx context.Context, userID int64, seq int) error {
	return s.update(ctx, userID, func(state *updates.State) { state.Seq = seq })
}

// SetDateSeq implements updates.StateStorage.
func (s *StateStorage) SetDateSeq(ctx context.Context, userID int64, date, seq int) error {
	return s.update(ctx, userID, func(state *updates.State) {
		state.Date = date
		state.Seq = seq
	})
}

// GetChannelPts implements updates.StateStorage.
func (s *StateStorage) GetChannelPts(_ context.Context, userID, channelID int64) (int, bool, error) {
	v, ok, err := get(s.db, userPairKey(prefixChannelPts, userID, channelID))
	if err != nil || !ok {
		return 0, false, err
	}
	return decodeInt(v, 0), true, nil
}

// SetChannelPts implements updates.StateStorage.
func (s *StateStorage) SetChannelPts(_ context.Context, userID, channelID int64, pts int) error {
	return s.db.Set(userPairKey(prefixChannelPts, userID, channelID), encodeInts(pts), s.writeOpts)
}

// ForEachChannels implements updates.StateStorage.
func (s *StateStorage) ForEachChannels(
	ctx context.Context,
	userID int64,
	f func(ctx context.Context, channelID int64, pts int) error,
) (rerr error) {
	prefix := userKey(prefixChannelPts, userID)
	iter, err := s.db.NewIter(prefixIterOptions(prefix))
	if err != nil {
		return errors.Wrap(err, "new iter")
	}
	defer func() {
		multierr.AppendInto(&rerr, iter.Close())
	}()

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		channelID := int64(binary.BigEndian.Uint64(key[len(prefix):]))
		if err := f(ctx, channelID, decodeInt(iter.Value(), 0)); err != nil {
			return err
		}
	}
	return iter.Error()
}

// AccessHasher is updates.ChannelAccessHasher and updates.UserAccessHasher
// implementation using pebble.
type AccessHasher struct {
	db        *pebble.DB
	writeOpts *pebble.WriteOptions
}

// NewAccessHasher creates new access hasher using pebble.
//
// Caller is responsible for db.Close() invocation.
func NewAccessHasher(db *pebble.DB) *AccessHasher {
	return &AccessHasher{db: db}
}

// WithWriteOptions sets pebble's write options for write operations.
//
// Default is nil, which pebble treats as pebble.Sync.
func (s *AccessHasher) WithWriteOptions(writeOpts *pebble.WriteOptions) *AccessHasher {
	s.writeOpts = writeOpts
	return s
}

func (s *AccessHasher) set(prefix string, userID, id, hash int64) error {
	key := userPairKey(prefix, userID, id)
	value := binary.BigEndian.AppendUint64(nil, uint64(hash))
	if old, ok, err := get(s.db, key); err == nil && ok && bytes.Equal(old, value) {
		// Skip unchanged: access hashes are set on every update.
		return nil
	}
	return s.db.Set(key, value, s.writeOpts)
}

func (s *AccessHasher) get(prefix string, userID, id int64) (int64, bool, error) {
	v, ok, err := get(s.db, userPairKey(prefix, userID, id))
	if err != nil || !ok {
		return 0, false, err
	}
	return int64(binary.BigEndian.Uint64(v)), true, nil
}

// SetChannelAccessHash implements updates.ChannelAccessHasher.
func (s *AccessHasher) SetChannelAccessHash(_ context.Context, userID, channelID, accessHash int64) error {
	return s.set(prefixChanHash, userID, channelID, accessHash)
}

// GetChannelAccessHash implements updates.ChannelAccessHasher.
func (s *AccessHasher) GetChannelAccessHash(_ context.Context, userID, channelID int64) (int64, bool, error) {
	return s.get(prefixChanHash, userID, channelID)
}

// SetUserAccessHash implements updates.UserAccessHasher.
func (s *AccessHasher) SetUserAccessHash(_ context.Context, userID, targetUserID, accessHash int64) error {
	return s.set(prefixUserHash, userID, targetUserID, accessHash)
}

// GetUserAccessHash implements updates.UserAccessHasher.
func (s *AccessHasher) GetUserAccessHash(_ context.Context, userID, targetUserID int64) (int64, bool, error) {
	return s.get(prefixUserHash, userID, targetUserID)
}
