package sequence

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

type rangeCollector struct {
	from, to  int // immutable
	left      int
	collected map[int]Update
	done      chan struct{}
	closed    bool
	mux       sync.Mutex
	log       *zap.Logger
}

func newRangeCollector(from, to int, log *zap.Logger) *rangeCollector {
	return &rangeCollector{
		from:      from,
		to:        to,
		left:      to - from + 1,
		collected: map[int]Update{},
		done:      make(chan struct{}),
		log: log.With(
			zap.Int("gap_from", from),
			zap.Int("gap_to", to),
		),
	}
}

func (c *rangeCollector) Consume(u Update) bool {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.closed {
		return false
	}

	log := c.log.With(
		zap.Int("upd_from", u.start()),
		zap.Int("upd_to", u.end()),
		zap.Int("total_left", c.left),
	)

	if c.from > u.start() || c.to < u.end() {
		log.Debug("Update out of range, ignoring")
		return false
	}

	for i := u.start(); i <= u.end(); i++ {
		if conflict, intersects := c.collected[i]; intersects {
			log.Warn("Update range conflict, ignoring",
				zap.Int("intersection_state", i),
				zap.Any("current_update", u.Value),
				zap.Any("conflicts_with", conflict.Value),
			)
			return false
		}
	}

	log.Debug("Update accepted")
	for i := u.start(); i <= u.end(); i++ {
		c.collected[i] = u
	}

	c.left -= u.Count
	if c.left == 0 {
		c.log.Debug("All required updates have been received")
		c.closed = true
		close(c.done)
	}

	if c.left < 0 {
		panic("unreachable")
	}

	return true
}

func (c *rangeCollector) Wait(ctx context.Context) ([]interface{}, error) {
	defer func() {
		c.mux.Lock()
		c.closed = true
		c.mux.Unlock()
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		var updates []interface{}
		for i := c.from; i <= c.to; {
			u, ok := c.collected[i]
			if !ok {
				c.log.Warn("Update missed", zap.Int("sequence_state", i))
				return nil, xerrors.Errorf("update %d missed", i)
			}

			updates = append(updates, u.Value)
			i += u.Count
		}

		return updates, nil
	}
}
