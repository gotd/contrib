package invoker

import (
	"context"
	"strconv"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"github.com/uber-go/tally"
)

// Metrics is a metrics exporting middleware for tg.Invoker.
type Metrics struct {
	next  tg.Invoker
	stats tally.Scope
}

// NewMetrics creates new Metrics.
func NewMetrics(next tg.Invoker, stats tally.Scope) Metrics {
	return Metrics{next: next, stats: stats}
}

type tgObject interface {
	TypeID() uint32
}

// Invoke implements tg.Invoker.
func (m Metrics) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	stats := m.stats
	tagging := stats.Capabilities().Tagging()

	if tagging {
		if obj, ok := input.(tgObject); ok {
			typeID := obj.TypeID()
			stats = stats.Tagged(map[string]string{
				"type_id": strconv.Itoa(int(typeID)),
			})
		}
	}

	sw := stats.Timer("tg_request_duration").Start()
	err := m.next.Invoke(ctx, input, output)
	sw.Stop()

	stats.Counter("tg_total_requests").Inc(1)
	if rpcErr, ok := tgerr.As(err); ok {
		if tagging {
			stats = stats.Tagged(map[string]string{
				"type":     rpcErr.Type,
				"code":     strconv.Itoa(rpcErr.Code),
				"argument": strconv.Itoa(rpcErr.Argument),
			})
		}

		stats.Counter("tg_requests_errors").Inc(1)
	}

	return err
}
