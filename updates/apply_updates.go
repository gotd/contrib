package updates

import (
	"sort"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"

	"github.com/gotd/contrib/updates/internal/sequence"
	"github.com/gotd/td/tg"
)

var errPtsChanged = xerrors.Errorf("pts changed")

func (e *Engine) applySeqUpdates(_ int, combs []interface{}) error {
	return e.applyUpdatesCombined(asCombined(combs)...)
}

//nolint:gocognit
func (e *Engine) applyUpdatesCombined(combs ...*tg.UpdatesCombined) error {
	type ptsUpdate struct {
		Pts, PtsCount int
		Update        tg.UpdateClass
	}

	type qtsUpdate struct {
		Qts    int
		Update tg.UpdateClass
	}

	// Data extracted from updates.
	var (
		otherUpdates   []tg.UpdateClass
		ptsUpdates     []ptsUpdate
		qtsUpdates     []qtsUpdate
		channelUpdates = make(map[int][]ptsUpdate)
		ents           = newEntities()
		ptsChanged     bool
		date, seq      int
	)

	// Sort updates.
	for _, comb := range combs {
		for _, u := range comb.Updates {
			if _, ok := u.(*tg.UpdatePtsChanged); ok {
				ptsChanged = true
				continue
			}

			if pts, ptsCount, ok := isCommonPtsUpdate(u); ok {
				ptsUpdates = append(ptsUpdates, ptsUpdate{
					Pts:      pts,
					PtsCount: ptsCount,
					Update:   u,
				})
				continue
			}

			if channelID, pts, ptsCount, ok, err := isChannelPtsUpdate(u); ok {
				if err != nil {
					return err
				}

				channelUpdates[channelID] = append(channelUpdates[channelID], ptsUpdate{
					Pts:      pts,
					PtsCount: ptsCount,
					Update:   u,
				})
				continue
			}

			if qts, ok := isCommonQtsUpdate(u); ok {
				qtsUpdates = append(qtsUpdates, qtsUpdate{
					Qts:    qts,
					Update: u,
				})
				continue
			}

			otherUpdates = append(otherUpdates, u)
		}

		// Fill entities.
		comb.MapUsers().FillUserMap(ents.Users)
		comb.MapChats().FillChatMap(ents.Chats)
		comb.MapChats().FillChannelMap(ents.Channels)
		date, seq = comb.Date, comb.Seq
	}

	// Handle all state-sensitive updates in their own sequence boxes.
	var g errgroup.Group
	g.Go(func() error {
		sort.SliceStable(ptsUpdates, func(i, j int) bool {
			l, r := ptsUpdates[i], ptsUpdates[j]
			return (l.Pts - l.PtsCount) < (r.Pts - r.PtsCount)
		})

		for _, u := range ptsUpdates {
			if err := e.handlePts(update{
				Value: u.Update,
				Ents:  ents,
			}, u.Pts, u.PtsCount); err != nil {
				return err
			}
		}
		return nil
	})

	g.Go(func() error {
		sort.SliceStable(qtsUpdates, func(i, j int) bool {
			l, r := qtsUpdates[i], qtsUpdates[j]
			return l.Qts < r.Qts
		})

		for _, u := range qtsUpdates {
			if err := e.handleQts(update{
				Value: u.Update,
				Ents:  ents,
			}, u.Qts); err != nil {
				return err
			}
		}
		return nil
	})

	for channelID, updates := range channelUpdates {
		channelID, updates := channelID, updates
		g.Go(func() error {
			sort.SliceStable(updates, func(i, j int) bool {
				l, r := updates[i], updates[j]
				return (l.Pts - l.PtsCount) < (r.Pts - r.PtsCount)
			})

			for _, u := range updates {
				if err := e.handleChannelPts(update{
					Value: u.Update,
					Ents:  ents,
				}, channelID, u.Pts, u.PtsCount); err != nil {
					return err
				}
			}
			return nil
		})
	}

	g.Go(func() error {
		if err := e.storage.SetDateSeq(date, seq); err != nil {
			return err
		}
		return e.handler.HandleUpdates(Updates{
			Updates: otherUpdates,
			Ents:    ents,
		})
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if ptsChanged {
		// Success, but we should notify e.handleUpdates about pts change.
		return &sequence.ResultError{
			Err: errPtsChanged,
		}
	}

	return nil
}

func (e *Engine) applyPtsUpdates(pts int, updatesIface []interface{}) error {
	var (
		updates []tg.UpdateClass
		ents    = newEntities()
	)

	for _, u := range updatesIface {
		update := u.(update) //nolint:govet
		updates = append(updates, update.Value.(tg.UpdateClass))
		ents.merge(update.Ents)
	}

	if err := e.storage.SetPts(pts); err != nil {
		return err
	}
	return e.handler.HandleUpdates(Updates{
		Updates: updates,
		Ents:    ents,
	})
}

func (e *Engine) applyQtsUpdates(qts int, updatesIface []interface{}) error {
	var (
		updates []tg.UpdateClass
		ents    = newEntities()
	)

	for _, u := range updatesIface {
		update := u.(update) //nolint:govet
		updates = append(updates, update.Value.(tg.UpdateClass))
		ents.merge(update.Ents)
	}

	if err := e.storage.SetQts(qts); err != nil {
		return err
	}
	return e.handler.HandleUpdates(Updates{
		Updates: updates,
		Ents:    ents,
	})
}

func (e *Engine) channelUpdateApplyFunc(channelID int) func(int, []interface{}) error {
	return func(pts int, updatesIface []interface{}) error {
		var (
			updates []tg.UpdateClass
			ents    = newEntities()
		)

		for _, u := range updatesIface {
			update := u.(update) //nolint:govet
			updates = append(updates, update.Value.(tg.UpdateClass))
			ents.merge(update.Ents)
		}

		if err := e.storage.SetChannelPts(channelID, pts); err != nil {
			return err
		}
		return e.handler.HandleUpdates(Updates{
			Updates: updates,
			Ents:    ents,
		})
	}
}
