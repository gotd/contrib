package updates

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/gotd/contrib/updates/internal/sequence"
	"github.com/gotd/td/tg"
)

// Handle update.
func (e *Engine) Handle(u tg.UpdatesClass) { e.uchan <- u }

// Run starts update handling.
func (e *Engine) Run(ctx context.Context) error {
	if err := e.init(ctx); err != nil {
		return xerrors.Errorf("init: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return e.recoverState(false) })

	e.chanMux.Lock()
	for channelID, state := range e.channels {
		channelID, state := channelID, state
		g.Go(func() error {
			return e.recoverChannelState(channelID, state)
		})
	}
	e.chanMux.Unlock()

	for i := 0; i < e.concurrency; i++ {
		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()

				case u := <-e.uchan:
					if err := e.handleUpdates(u); err != nil {
						return err
					}
				}
			}
		})
	}

	return g.Wait()
}

func (e *Engine) init(ctx context.Context) error {
	state, err := e.initState(ctx)
	if err != nil {
		return xerrors.Errorf("init state: %w", err)
	}

	if err := e.storage.Channels(func(channelID, pts int) {
		e.channels[channelID] = &channelState{
			pts: sequence.New(sequence.Config{
				InitialState: pts,
				Apply:        e.channelUpdateApplyFunc(channelID),
				Logger:       e.log.Named("channel_pts").With(zap.Int("channel_id", channelID)),
			}),
		}
	}); err != nil {
		return xerrors.Errorf("restora local channels state: %w", err)
	}

	e.date = state.Date
	e.seq = sequence.New(sequence.Config{
		InitialState: state.Seq,
		Apply:        e.applySeqUpdates,
		Logger:       e.log.Named("seq"),
	})
	e.pts = sequence.New(sequence.Config{
		InitialState: state.Pts,
		Apply:        e.applyPtsUpdates,
		Logger:       e.log.Named("pts"),
	})
	e.qts = sequence.New(sequence.Config{
		InitialState: state.Qts,
		Apply:        e.applyQtsUpdates,
		Logger:       e.log.Named("qts"),
	})

	return nil
}

func (e *Engine) initState(ctx context.Context) (State, error) {
	if e.forget {
		if err := e.storage.ForgetAll(); err != nil {
			return State{}, err
		}

		remote, err := e.raw.UpdatesGetState(ctx)
		if err != nil {
			return State{}, xerrors.Errorf("get remote state: %w", err)
		}

		state := State{}.fromRemote(remote)
		if err := e.storage.SetState(state); err != nil {
			return State{}, xerrors.Errorf("save remote state: %w", err)
		}

		return state, nil
	}

	state, err := e.storage.GetState()
	if err != nil {
		if errors.Is(err, ErrStateNotFound) {
			remote, err := e.raw.UpdatesGetState(ctx)
			if err != nil {
				return State{}, xerrors.Errorf("get remote state: %w", err)
			}

			state = state.fromRemote(remote)
			if err := e.storage.SetState(state); err != nil {
				return State{}, xerrors.Errorf("save remote state: %w", err)
			}
		} else {
			return State{}, xerrors.Errorf("restore local state: %w", err)
		}
	}

	return state, nil
}
