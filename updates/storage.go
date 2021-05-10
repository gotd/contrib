package updates

import (
	"sync"

	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

// State contains common sequence state.
type State struct {
	Pts  int
	Qts  int
	Date int
	Seq  int
}

func (s State) fromRemote(remote *tg.UpdatesState) State {
	s.Pts = remote.Pts
	s.Qts = remote.Qts
	s.Date = remote.Date
	s.Seq = remote.Seq
	return s
}

// ErrStateNotFound says that we do not have a local state.
var ErrStateNotFound = xerrors.Errorf("state not found")

// Storage interface.
type Storage interface {
	GetState() (State, error)
	SetState(s State) error
	SetPts(pts int) error
	SetQts(qts int) error
	SetDateSeq(date, seq int) error

	SetChannelPts(channelID, pts int) error
	Channels(iter func(channelID, pts int)) error
	ForgetAll() error
}

type memStorage struct {
	state    State
	channels map[int]int
	mux      sync.Mutex
}

func newMemStorage() *memStorage {
	return &memStorage{
		channels: map[int]int{},
	}
}

func (ms *memStorage) GetState() (State, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	return State{}, ErrStateNotFound
}

func (ms *memStorage) SetState(s State) error {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	ms.state = s
	return nil
}

func (ms *memStorage) SetPts(pts int) error {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	ms.state.Pts = pts
	return nil
}

func (ms *memStorage) SetQts(qts int) error {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	ms.state.Qts = qts
	return nil
}

func (ms *memStorage) SetDateSeq(date, seq int) error {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	ms.state.Date = date
	ms.state.Seq = seq
	return nil
}

func (ms *memStorage) SetChannelPts(channelID, pts int) error {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	ms.channels[channelID] = pts
	return nil
}

func (ms *memStorage) Channels(iter func(channelID, pts int)) error { return nil }

func (ms *memStorage) ForgetAll() error { return nil }
