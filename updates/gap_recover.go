package updates

import (
	"context"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

func (e *Engine) recoverGap() error {
	var (
		localDate = e.date
		localSeq  = e.seq.GetState()
		localPts  = e.pts.GetState()
		localQts  = e.qts.GetState()
	)

	e.log.Debug("Recovering common state",
		zap.Int("local_date", localDate),
		zap.Int("local_seq", localSeq),
		zap.Int("local_pts", localPts),
		zap.Int("local_qts", localQts),
	)

	setState := func(state tg.UpdatesState) error {
		e.log.Debug("Set common state",
			zap.Int("new_date", state.Date),
			zap.Int("new_seq", state.Seq),
			zap.Int("new_pts", state.Pts),
			zap.Int("new_qts", state.Qts),
		)

		if err := e.storage.SetState((State{}).fromRemote(&state)); err != nil {
			return err
		}

		e.pts.SetState(state.Pts)
		e.qts.SetState(state.Qts)
		e.seq.SetState(state.Seq)
		e.date = state.Date
		return nil
	}

	fillPendingData := func(d *DiffUpdate) {
		updates, ents := asUpdateClasses(e.pts.ExtractBuffer())
		d.Pending = append(d.Pending, updates...)
		d.PendingEnts.merge(ents)

		updates, ents = asUpdateClasses(e.qts.ExtractBuffer())
		d.Pending = append(d.Pending, updates...)
		d.PendingEnts.merge(ents)

		updates, ents = extractUpdatesFromCombs(asCombined(e.seq.ExtractBuffer()))
		d.Pending = append(d.Pending, updates...)
		d.PendingEnts.merge(ents)
	}

	diff, err := e.raw.UpdatesGetDifference(context.TODO(), &tg.UpdatesGetDifferenceRequest{
		Pts:  localPts,
		Qts:  localQts,
		Date: localDate,
	})
	if err != nil {
		return xerrors.Errorf("get difference: %w", err)
	}

	switch diff := diff.(type) {
	case *tg.UpdatesDifference:
		e.saveChannelHashes(diff.Chats)
		if err := e.handleChannelTooLong(diff.OtherUpdates); err != nil {
			return xerrors.Errorf("handle channelTooLong: %w", err)
		}

		if err := setState(diff.State); err != nil {
			return err
		}

		d := DiffUpdate{
			NewMessages:          diff.NewMessages,
			NewEncryptedMessages: diff.NewEncryptedMessages,
			Users:                diff.Users,
			Chats:                diff.Chats,
			OtherUpdates:         diff.OtherUpdates,
			PendingEnts:          newEntities(),
		}

		fillPendingData(&d)
		return e.handler.HandleDiff(d)

	// No events.
	case *tg.UpdatesDifferenceEmpty:
		e.date = diff.Date
		e.seq.SetState(diff.Seq)
		if err := e.storage.SetDateSeq(diff.Date, diff.Seq); err != nil {
			return err
		}

		return nil

	// Incomplete list of occurred events.
	case *tg.UpdatesDifferenceSlice:
		e.saveChannelHashes(diff.Chats)
		if err := e.handleChannelTooLong(diff.OtherUpdates); err != nil {
			return xerrors.Errorf("handle channelTooLong: %w", err)
		}

		if err := setState(diff.IntermediateState); err != nil {
			return err
		}

		d := DiffUpdate{
			NewMessages:          diff.NewMessages,
			NewEncryptedMessages: diff.NewEncryptedMessages,
			Users:                diff.Users,
			Chats:                diff.Chats,
			OtherUpdates:         diff.OtherUpdates,
			PendingEnts:          newEntities(),
		}

		fillPendingData(&d)
		if err := e.handler.HandleDiff(d); err != nil {
			return err
		}

		return e.recoverGap()

	// The difference is too long, and the specified state must be used to refetch updates.
	case *tg.UpdatesDifferenceTooLong:
		e.pts.SetState(diff.Pts)
		return e.recoverGap()

	default:
		return xerrors.Errorf("unexpected diff type: %T", diff)
	}
}

func (e *Engine) recoverChannelGap(channelID int, accessHash int64, state *channelState) error {
	if now := time.Now(); now.Before(state.timeout) {
		time.Sleep(state.timeout.Sub(now))
	}

	var (
		localPts = state.pts.GetState()
		log      = e.log.With(zap.Int("channel_id", channelID))
	)

	log.Debug("Recovering channel state", zap.Int("local_pts", localPts))
	diff, err := e.raw.UpdatesGetChannelDifference(context.TODO(), &tg.UpdatesGetChannelDifferenceRequest{
		Channel: &tg.InputChannel{
			ChannelID:  channelID,
			AccessHash: accessHash,
		},
		Filter: &tg.ChannelMessagesFilterEmpty{},
		Pts:    localPts,
		Limit:  e.chanDiffLimit,
	})
	if err != nil {
		return xerrors.Errorf("get channel difference: %w", err)
	}

	switch diff := diff.(type) {
	case *tg.UpdatesChannelDifference:
		e.saveChannelHashes(diff.Chats)
		d := DiffUpdate{
			NewMessages:  diff.NewMessages,
			OtherUpdates: diff.OtherUpdates,
			Users:        diff.Users,
			Chats:        diff.Chats,
			PendingEnts:  newEntities(),
		}

		updates, ents := asUpdateClasses(state.pts.ExtractBuffer())
		d.Pending = append(d.Pending, updates...)
		d.PendingEnts.merge(ents)

		if err := e.storage.SetChannelPts(channelID, diff.Pts); err != nil {
			return err
		}

		if err := e.handler.HandleDiff(d); err != nil {
			return err
		}

		state.pts.SetState(diff.Pts)
		if seconds, ok := diff.GetTimeout(); ok {
			state.timeout = time.Now().Add(time.Second * time.Duration(seconds+1))
		}

		if !diff.Final {
			return e.recoverChannelGap(channelID, accessHash, state)
		}

		return nil

	case *tg.UpdatesChannelDifferenceEmpty:
		if err := e.storage.SetChannelPts(channelID, diff.Pts); err != nil {
			return err
		}

		state.pts.SetState(diff.Pts)
		if seconds, ok := diff.GetTimeout(); ok {
			state.timeout = time.Now().Add(time.Second * time.Duration(seconds+1))
		}
		return nil

	case *tg.UpdatesChannelDifferenceTooLong:
		e.saveChannelHashes(diff.Chats)
		// Reset channel state.
		e.chanMux.Lock()
		delete(e.channels, channelID)
		e.chanMux.Unlock()

		e.handler.ChannelTooLong(channelID)
		if seconds, ok := diff.GetTimeout(); ok {
			state.timeout = time.Now().Add(time.Second * time.Duration(seconds+1))
		}
		return nil

	default:
		return xerrors.Errorf("unexpected channel diff type: %T", diff)
	}
}

func (e *Engine) handleChannelTooLong(others []tg.UpdateClass) error {
	var g errgroup.Group
	for _, u := range others {
		long, ok := u.(*tg.UpdateChannelTooLong)
		if !ok {
			continue
		}

		g.Go(func() error {
			e.chanMux.Lock()
			state, ok := e.channels[long.ChannelID]
			e.chanMux.Unlock()
			if !ok {
				// Note:
				// If this channel is not in the state,
				// most likely we also do not have its access hash.
				// Notify handler about unrecoverable gap.
				//
				// TODO: Reset channel pts?
				// TODO: Trigger common state gap recover?
				e.handler.ChannelTooLong(long.ChannelID)
				return nil
			}

			if err := e.recoverChannelState(long.ChannelID, state); err != nil {
				return xerrors.Errorf("recover channelTooLong(id: %d): %w", long.ChannelID, err)
			}

			return nil
		})
	}

	return g.Wait()
}

func (e *Engine) saveChannelHashes(chats []tg.ChatClass) {
	e.hashMux.Lock()
	defer e.hashMux.Unlock()

	for _, c := range chats {
		switch c := c.(type) {
		case *tg.Channel:
			if hash, ok := c.GetAccessHash(); ok && !c.Min {
				if oldHash, haveOld := e.hashes[c.ID]; haveOld {
					if hash == oldHash {
						continue
					}
					e.log.Info("Channel access hash changed",
						zap.Int("channel_id", c.ID),
						zap.String("channel_name", c.Username),
					)
				} else {
					e.log.Info("New channel access hash",
						zap.Int("channel_id", c.ID),
						zap.String("channel_name", c.Username),
					)
				}

				e.hashes[c.ID] = hash
			}
		case *tg.ChannelForbidden:
			if oldHash, haveOld := e.hashes[c.ID]; haveOld {
				if c.AccessHash == oldHash {
					continue
				}
				e.log.Info("Forbidden channel access hash changed",
					zap.Int("channel_id", c.ID),
					zap.String("channel_title", c.Title),
				)
			} else {
				e.log.Info("New forbidden channel access hash",
					zap.Int("channel_id", c.ID),
					zap.String("channel_title", c.Title),
				)
			}

			e.hashes[c.ID] = c.AccessHash
		}
	}
}
