package floodwait

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	a := require.New(t)

	q := newQueue(16)
	now := time.Date(2077, 10, 23, 0, 3, 0, 0, time.UTC)
	for i := range [10]struct{}{} {
		q.add(request{
			key: key(i),
		}, now.Add(time.Duration(i)*time.Second))
	}

	a.Equal(10, q.len())

	a.Len(q.gather(now.Add(1*time.Millisecond), nil), 1)
	a.Equal(9, q.len())

	now = now.Add(10 * time.Second)
	q.move(5, now, 10*time.Second)

	a.Len(q.gather(now, nil), 8)
	a.Equal(1, q.len())

	a.Len(q.gather(now.Add(10*time.Second), nil), 1)
	a.Equal(0, q.len())
}

func TestQueueMoveUnsortedHeap(t *testing.T) {
	// The heap array is not sorted: a later request of the key may come
	// before an earlier one, which must still be moved.
	now := time.Unix(100, 0)
	q := newQueue(4)
	q.add(request{key: 2}, now)                     // root
	q.add(request{key: 1}, now.Add(10*time.Second)) // index 1, not moved
	q.add(request{key: 1}, now.Add(time.Second))    // index 2, moved

	q.move(1, now, 5*time.Second)

	got := map[time.Duration]int{}
	for _, s := range q.requests {
		if s.request.key == 1 {
			got[s.sendTime.Sub(now)]++
		}
	}
	require.Equal(t, map[time.Duration]int{
		10 * time.Second: 1,
		6 * time.Second:  1,
	}, got)
}
