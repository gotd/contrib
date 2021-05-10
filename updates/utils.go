package updates

import (
	"golang.org/x/xerrors"

	"github.com/gotd/td/tg"
)

func validatePts(pts, ptsCount int) error {
	if pts < 0 {
		return xerrors.Errorf("invalid pts value: %d", pts)
	}

	if ptsCount < 0 {
		return xerrors.Errorf("invalid pts_count value: %d", ptsCount)
	}

	return nil
}

func validateQts(qts int) error {
	if qts < 0 {
		return xerrors.Errorf("negative qts value: %d", qts)
	}

	return nil
}

func validateSeq(seq, seqStart int) error {
	if seq < 0 {
		return xerrors.Errorf("invalid seq value: %d", seq)
	}

	if seqStart < 0 {
		return xerrors.Errorf("invalid seq_start value: %d", seq)
	}

	return nil
}

func asCombined(values []interface{}) []*tg.UpdatesCombined {
	var (
		updates []*tg.UpdatesCombined
	)
	for _, u := range values {
		update, ok := u.(*tg.UpdatesCombined)
		if !ok {
			panic("unreachable")
		}

		updates = append(updates, update)
	}

	return updates
}

func asUpdateClasses(values []interface{}) ([]tg.UpdateClass, *Entities) {
	var (
		updates []tg.UpdateClass
		ents    = newEntities()
	)
	for _, u := range values {
		update, ok := u.(update) //nolint:govet
		if !ok {
			panic("unreachable")
		}

		updates = append(updates, update.Value.(tg.UpdateClass))
		if update.Ents != nil {
			ents.merge(update.Ents)
		}
	}

	return updates, ents
}

func extractUpdatesFromCombs(combs []*tg.UpdatesCombined) ([]tg.UpdateClass, *Entities) {
	var (
		updates []tg.UpdateClass
		ents    = newEntities()
	)

	for _, comb := range combs {
		updates = append(updates, comb.Updates...)
		comb.MapUsers().FillUserMap(ents.Users)
		comb.MapChats().FillChatMap(ents.Chats)
		comb.MapChats().FillChannelMap(ents.Channels)
	}

	return updates, ents
}
