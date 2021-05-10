package updates

import (
	"errors"
	"sync"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/contrib/updates/internal/sequence"
	"github.com/gotd/td/tg"
)

type channelState struct {
	pts        *sequence.Box
	recovering atomic.Bool
	timeout    time.Time
}

// Engine deals with gaps.
type Engine struct {
	raw RawClient

	pts  *sequence.Box
	qts  *sequence.Box
	seq  *sequence.Box
	date int

	recovering     bool
	recoverWaiters []chan error
	recoverMux     sync.Mutex

	channels map[int]*channelState
	chanMux  sync.Mutex
	hashes   map[int]int64
	hashMux  sync.Mutex

	uchan chan tg.UpdatesClass

	storage       Storage     // immutable
	selfID        int         // immutable
	chanDiffLimit int         // immutable
	concurrency   int         // immutable
	handler       Handler     // immutable
	forget        bool        // immutable
	log           *zap.Logger // immutable
}

// New creates new engine.
func New(cfg Config) *Engine {
	cfg.setDefaults()
	e := &Engine{
		raw: cfg.RawClient,

		channels: map[int]*channelState{},
		hashes:   make(map[int]int64),
		uchan:    make(chan tg.UpdatesClass, 10),

		storage:       cfg.Storage,
		selfID:        cfg.SelfID,
		chanDiffLimit: 100,
		concurrency:   cfg.Concurrency,
		handler:       cfg.Handler,
		forget:        cfg.Forget,
		log:           cfg.Logger,
	}

	if cfg.IsBot {
		e.chanDiffLimit = 100000
	}

	return e
}

func (e *Engine) handleUpdates(u tg.UpdatesClass) error {
	if err := func() error {
		switch u := u.(type) {
		case *tg.Updates:
			return e.handleSeq(&tg.UpdatesCombined{
				Updates:  u.Updates,
				Users:    u.Users,
				Chats:    u.Chats,
				Date:     u.Date,
				Seq:      u.Seq,
				SeqStart: u.Seq,
			}, u.Seq, u.Seq)

		case *tg.UpdatesCombined:
			return e.handleSeq(u, u.Seq, u.SeqStart)

		case *tg.UpdateShort:
			return e.handleUpdate(u.Update)

		case *tg.UpdateShortMessage:
			return e.handleUpdates(e.convertShortMessage(u))

		case *tg.UpdateShortChatMessage:
			return e.handleUpdates(e.convertShortChatMessage(u))

		case *tg.UpdateShortSentMessage:
			return e.handleUpdates(e.convertShortSentMessage(u))

		case *tg.UpdatesTooLong:
			return e.recoverState(false)

		default:
			return xerrors.Errorf("unexpected tg.UpdatesClass type: %T", u)
		}
	}(); err != nil {
		if errors.Is(err, sequence.ErrGap) {
			e.log.Debug("Gap, recovering state")
			return e.recoverState(false)
		}

		if errors.Is(err, errPtsChanged) {
			e.log.Debug("Pts changed, recovering state")
			return e.recoverState(false)
		}

		return err
	}

	return nil
}

func (e *Engine) handleUpdate(u tg.UpdateClass) error {
	if _, ok := u.(*tg.UpdatePtsChanged); ok {
		return e.recoverState(false)
	}

	if pts, ptsCount, ok := isCommonPtsUpdate(u); ok {
		return e.handlePts(update{Value: u}, pts, ptsCount)
	}

	if qts, ok := isCommonQtsUpdate(u); ok {
		return e.handleQts(update{Value: u}, qts)
	}

	if channelID, pts, ptsCount, ok, err := isChannelPtsUpdate(u); ok {
		if err != nil {
			e.log.Warn("Invalid channel update", zap.Error(err))
			return nil
		}

		return e.handleChannelPts(update{Value: u}, channelID, pts, ptsCount)
	}

	return e.handler.HandleUpdates(Updates{Updates: []tg.UpdateClass{u}})
}

func (e *Engine) handlePts(u update, pts, ptsCount int) error {
	if err := validatePts(pts, ptsCount); err != nil {
		e.log.Warn("Pts validation failed", zap.Error(err))
		return nil
	}

	return e.pts.Handle(sequence.Update{
		State: pts,
		Count: ptsCount,
		Value: u,
	})
}

func (e *Engine) handleQts(u update, qts int) error {
	if err := validateQts(qts); err != nil {
		e.log.Warn("Qts validation failed", zap.Error(err))
		return nil
	}

	return e.qts.Handle(sequence.Update{
		State: qts,
		Count: 1,
		Value: u,
	})
}

func (e *Engine) handleSeq(u *tg.UpdatesCombined, seq, seqStart int) error {
	if err := validateSeq(seq, seqStart); err != nil {
		e.log.Warn("Seq validation failed", zap.Error(err))
		return nil
	}

	// Special case.
	if seq == 0 {
		return e.applyUpdatesCombined(u)
	}

	return e.seq.Handle(sequence.Update{
		State: seq,
		Count: seq - seqStart + 1,
		Value: u,
	})
}

func (e *Engine) handleChannelPts(u update, channelID, pts, ptsCount int) error {
	if err := validatePts(pts, ptsCount); err != nil {
		e.log.Warn("Pts validation failed", zap.Error(err), zap.Int("channel_id", channelID))
		return nil
	}

	log := e.log.With(zap.Int("channel_id", channelID))

	e.chanMux.Lock()
	state, ok := e.channels[channelID]
	if !ok {
		if err := e.storage.SetChannelPts(channelID, pts-ptsCount); err != nil {
			e.chanMux.Unlock()
			return err
		}

		state = &channelState{
			pts: sequence.New(sequence.Config{
				InitialState: pts - ptsCount,
				Apply:        e.channelUpdateApplyFunc(channelID),
				Logger:       log.Named("channel_pts"),
			}),
		}

		e.channels[channelID] = state
	}
	e.chanMux.Unlock()

	err := state.pts.Handle(sequence.Update{
		State: pts,
		Count: ptsCount,
		Value: u,
	})

	if errors.Is(err, sequence.ErrGap) {
		log.Debug("Channel pts gap, recovering state")
		return e.recoverChannelState(channelID, state)
	}

	return err
}

func (e *Engine) recoverState(wait bool) (err error) {
	e.recoverMux.Lock()
	if e.recovering {
		if wait {
			c := make(chan error)
			e.recoverWaiters = append(e.recoverWaiters, c)
			e.recoverMux.Unlock()
			return <-c
		}

		e.recoverMux.Unlock()
		return nil
	}
	e.recovering = true
	e.recoverMux.Unlock()

	defer func() {
		e.recoverMux.Lock()
		defer e.recoverMux.Unlock()
		for _, waiter := range e.recoverWaiters {
			waiter <- err
		}
		e.recoverWaiters = nil
		e.recovering = false
	}()

	e.pts.EnableBuffering()
	e.qts.EnableBuffering()
	e.seq.EnableBuffering()

	defer func() {
		e.pts.DisableBuffering()
		e.qts.DisableBuffering()
		e.seq.DisableBuffering()
	}()

	err = e.recoverGap()
	return err
}

func (e *Engine) recoverChannelState(channelID int, state *channelState) error {
	// Prevent parallel execution.
	if !state.recovering.CAS(false, true) {
		return nil
	}
	defer state.recovering.Store(false)

	accessHash, ok := e.getChannelAccessHash(channelID)
	if !ok {
		e.log.Info("No channel access hash. GAP.", zap.Int("channel_id", channelID))
		// Try to restore missing hash by updates.getDifference.
		if err := e.recoverState(true); err != nil {
			return err
		}

		// Check if the required hash has arrived.
		accessHash, ok = e.getChannelAccessHash(channelID)
		if !ok {
			return xerrors.Errorf("server did not return required channel hash")
		}
	}

	state.pts.EnableBuffering()
	defer state.pts.DisableBuffering()

	if err := e.recoverChannelGap(channelID, accessHash, state); err != nil {
		return xerrors.Errorf("recover channel state(id: %d): %w", channelID, err)
	}

	return nil
}

func (e *Engine) getChannelAccessHash(channelID int) (int64, bool) {
	e.hashMux.Lock()
	defer e.hashMux.Unlock()

	accessHash, ok := e.hashes[channelID]
	return accessHash, ok
}
