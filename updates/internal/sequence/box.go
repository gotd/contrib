package sequence

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

// Box helps to apply updates in correct order.
type Box struct {
	state int
	mode  mode

	// Pending updates.
	buffer []Update
	mux    sync.Mutex

	// Fastgap.
	fgCollector *rangeCollector
	fgTimeout   time.Duration
	fgWaiters   []chan struct{}
	fgForceBuf  bool

	apply func(state int, updates []interface{}) error
	log   *zap.Logger

	hooks hooks
}

// New creates new sequence box.
func New(cfg Config) *Box {
	cfg.setDefaults()
	cfg.Logger.Info("Initialized",
		zap.Int("state", cfg.InitialState),
		zap.Duration("fastgap_timeout", cfg.FastgapTimeout),
	)

	return &Box{
		state:     cfg.InitialState,
		mode:      modeNormal,
		apply:     cfg.Apply,
		fgTimeout: cfg.FastgapTimeout,
		log:       cfg.Logger,
		hooks:     cfg.hooks,
	}
}

// Handle handles update.
// Transitions:
//  * normal -> fastgap
func (b *Box) Handle(u Update) error {
	b.mux.Lock()
	log := b.log.With(
		zap.Int("upd_from", u.start()),
		zap.Int("upd_to", u.end()),
	)

	// Ignore outdated updates.
	if checkGap(b.state, u.State, u.Count) == gapIgnore {
		b.mux.Unlock()
		log.Debug("Outdated update, skip")
		return nil
	}

	switch b.mode {
	case modeNormal:
		switch checkGap(b.state, u.State, u.Count) {
		case gapApply:
			defer b.mux.Unlock()
			return b.applyUpdates(u.State, []interface{}{u.Value})

		case gapRefetch:
			b.setMode(modeFastgap)

			gapFrom, gapTo := b.state+1, u.start()-1
			b.fgCollector = newRangeCollector(gapFrom, gapTo, b.log.Named("fgc"))

			// Check if we already have acceptable updates in buffer.
			for _, u := range b.buffer {
				// TODO: Remove accepted updates from buffer?
				_ = b.fgCollector.Consume(u)
			}
			b.buffer = append(b.buffer, u)
			b.mux.Unlock()

			log.Debug("Gap detected",
				zap.Int("gap_from", gapFrom),
				zap.Int("gap_to", gapTo),
				zap.Duration("timeout", b.fgTimeout),
			)

			return b.handleFastgap(u)

		default:
			panic("unreachable")
		}

	case modeFastgap:
		defer b.mux.Unlock()
		b.buffer = append(b.buffer, u)
		if b.fgCollector.Consume(u) {
			log.Debug("Fastgap accepted")
			return nil
		}

		return nil

	case modeBuffer:
		defer b.mux.Unlock()
		log.Debug("Send update to buffer")
		b.buffer = append(b.buffer, u)
		return nil
	default:
		panic("unreachable")
	}
}

// Transitions:
//  * fastgap -> normal
//  * fastgap -> buffer
func (b *Box) handleFastgap(head Update) error {
	ctx, cancel := context.WithTimeout(context.Background(), b.fgTimeout)
	defer cancel()

	b.hooks.FastgapBegin()
	resolvedUpdates, err := b.fgCollector.Wait(ctx)
	b.hooks.FastgapCollectorFinished()

	b.mux.Lock()
	defer func() {
		for _, c := range b.fgWaiters {
			close(c)
		}
		b.fgWaiters = nil
		b.fgForceBuf = false
		b.fgCollector = nil
		b.mux.Unlock()
		b.hooks.FastgapEnd()
	}()

	if err != nil {
		b.setMode(modeBuffer)

		if errors.Is(err, context.DeadlineExceeded) {
			b.log.Debug("Fastgap deadline exceeded, need to fetch updates manually")
			return ErrGap
		}

		b.log.Warn("Unexpected collector error", zap.Error(err))
		return err
	}

	newMode := modeNormal
	if b.fgForceBuf {
		newMode = modeBuffer
		b.log.Debug("Force buffer mode")
	}

	b.setMode(newMode)

	b.log.Debug("Gap was resolved by waiting. Applying updates.")
	return b.applyUpdates(head.State, append(resolvedUpdates, head.Value))
}

func (b *Box) applyUpdates(state int, updates []interface{}) error {
	// Check for acceptable updates in buffer.
	if newState, newUpdates, ok := b.haveNewerInBuffer(state); ok {
		b.log.Debug("Have newer updates in buffer",
			zap.Int("old_state", state),
			zap.Int("new_state", newState),
			zap.Int("new_updates_count", len(newUpdates)),
		)
		state = newState
		updates = append(updates, newUpdates...)
	}

	b.hooks.Apply(state, updates)
	err := b.apply(state, updates)
	if err != nil {
		var result *ResultError
		if !errors.As(err, &result) {
			b.log.Error("Apply function returned error", zap.Error(err))
			return err
		}

		err = result.Err
	}

	b.log.Debug("New state", zap.Int("new_state", state))
	b.state = state
	return err
}

// haveNewerInBuffer checks buffer for acceptable updates starting from 'state'.
// Also removes outdated updates.
func (b *Box) haveNewerInBuffer(state int) (newState int, updates []interface{}, ok bool) {
	sort.SliceStable(b.buffer, func(i, j int) bool {
		return b.buffer[i].start() < b.buffer[j].start()
	})

	cursor := 0

loop:
	for i, u := range b.buffer {
		cursor = i
		switch checkGap(state, u.State, u.Count) {
		case gapApply:
			updates = append(updates, u.Value)
			state = u.State
			continue
		case gapIgnore:
			continue
		case gapRefetch:
			break loop
		default:
			panic("unreachable")
		}
	}

	// Erase outdated updates.
	b.buffer = b.buffer[cursor:]
	if cursor > 0 {
		b.log.Debug("Outdated updates erased from buffer",
			zap.Int("erased", cursor),
			zap.Int("new_buffer_size", len(b.buffer)),
		)
	}

	if len(updates) == 0 {
		return 0, nil, false
	}

	return state, updates, true
}

// GetState returns current box state.
func (b *Box) GetState() int {
	b.mux.Lock()
	defer b.mux.Unlock()

	if b.mode != modeBuffer {
		panic(xerrors.Errorf("bad mode: %s", b.mode))
	}

	return b.state
}

// SetState manually sets new box state.
func (b *Box) SetState(state int) {
	b.mux.Lock()
	defer b.mux.Unlock()

	if b.mode != modeBuffer {
		panic(xerrors.Errorf("bad mode: %s", b.mode))
	}

	b.log.Debug("Set state", zap.Int("new_state", state))
	b.state = state
}

// EnableBuffering enables buffer mode.
// Should be called before fetching diffs.
//
// Transitions:
//  * normal  -> buffer
//  * fastgap -> buffer
func (b *Box) EnableBuffering() {
	b.mux.Lock()

	switch b.mode {
	case modeNormal:
		b.setMode(modeBuffer)
		b.log.Debug("Buffer mode enabled manually")
		b.mux.Unlock()

	case modeFastgap:
		b.log.Debug("Enabling buffer mode manually during fastgap")
		b.fgForceBuf = true
		waitChan := make(chan struct{})
		b.fgWaiters = append(b.fgWaiters, waitChan)
		b.mux.Unlock()

		<-waitChan

		// TODO: Unnecessary check.
		b.mux.Lock()
		if b.mode != modeBuffer {
			panic(xerrors.Errorf("bad mode: %s", b.mode))
		}
		b.mux.Unlock()
		b.log.Debug("Buffer mode enabled manually after fastgap stage")

	case modeBuffer:
		// Do nothing.
		b.mux.Unlock()
	}
}

// DisableBuffering disables buffer mode.
// Should be called after the difference has been received.
//
// Transitions:
//  * buffer -> normal
func (b *Box) DisableBuffering() {
	b.mux.Lock()
	defer b.mux.Unlock()
	if b.mode != modeBuffer {
		panic(xerrors.Errorf("bad mode: %s", b.mode))
	}

	b.setMode(modeNormal)
}

// ExtractBuffer returns buffered updates.
func (b *Box) ExtractBuffer() []interface{} {
	b.mux.Lock()
	defer b.mux.Unlock()
	var updates []interface{}
	for _, u := range b.buffer {
		updates = append(updates, u.Value)
	}

	b.log.Debug("Extract updates from buffer", zap.Int("updates_count", len(updates)))
	b.buffer = nil
	return updates
}

func (b *Box) setMode(m mode) {
	prev := b.mode
	b.mode = m
	if prev != m {
		b.log.Debug("Transition",
			zap.Stringer("prev_mode", prev),
			zap.Stringer("new_mode", m),
		)
		b.hooks.Transition(prev, m)
	}
}
