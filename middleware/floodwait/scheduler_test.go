package floodwait

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gotd/neo"
)

func TestScheduler(t *testing.T) {
	a := require.New(t)
	n := neo.NewTime(time.Now())
	sch := newScheduler(n, time.Second)

	r := request{
		key: 1,
	}
	// Schedule request and send it.
	sch.schedule(r)
	a.Len(sch.gather(nil), 1)
	// Got flood wait.
	sch.flood(r, 5*time.Second)
	// Ensure that request re-scheduled.
	a.Empty(sch.gather(nil))

	// Travel and ensure that request re-scheduled correctly.
	n.Travel(5*time.Second + time.Millisecond)
	a.Len(sch.gather(nil), 1)

	// Decrease wait timeout.
	sch.nice(1)
	// Schedule yet one
	sch.schedule(request{
		key: 1,
	})
	// Ensure that timer decreased correctly.
	n.Travel(4*time.Second + time.Millisecond)
	a.Len(sch.gather(nil), 1)
}

func TestSchedulerFloodNotDelayedTwice(t *testing.T) {
	// The flooded request itself is sent after the wait, not after twice the
	// wait: moving queued requests of the type must not move it again.
	n := neo.NewTime(time.Unix(0, 0))
	s := newScheduler(n, time.Second)
	s.flood(request{key: 1}, 5*time.Second)

	n.Travel(5*time.Second - time.Millisecond)
	require.Empty(t, s.gather(nil))
	n.Travel(2 * time.Millisecond)
	require.Len(t, s.gather(nil), 1)
}
