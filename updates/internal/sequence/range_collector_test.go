package sequence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRangeCollector(t *testing.T) {
	type event struct {
		Update
		Consume bool
	}

	tests := []struct {
		Name     string
		From, To int
		Events   []event
	}{
		{
			Name: "test1",
			From: 5, // 5 6 7 8 9 10 11 12 13 14 = 10 updates
			To:   14,
			Events: []event{
				{
					Update: Update{
						Value: 1,
						State: 4, // ignore
						Count: 1, // cursor 5
					},
					Consume: false,
				},
				{
					Update: Update{
						Value: 2,
						State: 5, // apply
						Count: 1, // cursor 6
					},
					Consume: true,
				},
				{
					Update: Update{
						Value: 3,
						State: 8, // apply
						Count: 3, // cursor 8
					},
					Consume: true,
				},
				{
					Update: Update{
						Value: 4,
						State: 16, // ignore, out of range
						Count: 8,  // cursor 8
					},
					Consume: false,
				},
				{
					Update: Update{
						Value: 5,
						State: 9, // apply
						Count: 1, // cursor 9
					},
					Consume: true,
				},
				{
					Update: Update{
						Value: 6,
						State: 14, // apply
						Count: 5,  // cursor 14
					},
					Consume: true,
				},
			},
		},
	}

	for _, test := range tests {
		c := newRangeCollector(test.From, test.To, zap.NewNop())

		var acceptedEvents []interface{}
		for _, ev := range test.Events {
			consumed := c.Consume(ev.Update)
			require.Equal(t, ev.Consume, consumed, ev)
			if consumed {
				acceptedEvents = append(acceptedEvents, ev.Value)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		updates, err := c.Wait(ctx)
		require.NoError(t, err)
		require.Equal(t, len(acceptedEvents), len(updates))
		for i, u := range updates {
			require.True(t, acceptedEvents[i] == u)
		}
	}
}
