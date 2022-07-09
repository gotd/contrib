// Package oteltg provides OpenTelemetry instrumentation for gotd.
package oteltg

import (
	"context"
	"strconv"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
)

// Middleware is prometheus metrics middleware for Telegram.
type Middleware struct {
	count    asyncint64.Counter
	failures asyncint64.Counter
	duration syncfloat64.Histogram
}

// Handle implements telegram.Middleware.
func (m Middleware) Handle(next tg.Invoker) telegram.InvokeFunc {
	return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		// Prepare.
		attrs := m.attributes(input)
		m.count.Observe(ctx, 1, attrs...)
		start := time.Now()

		// Call actual method.
		err := next.Invoke(ctx, input, output)

		// Observe.
		m.duration.Record(ctx, time.Since(start).Seconds(), attrs...)
		if err != nil {
			if rpcErr, ok := tgerr.As(err); ok {
				attrs = append(attrs,
					attribute.String("tg.rpc.err", rpcErr.Type),
					attribute.String("tg.rpc.code", strconv.Itoa(rpcErr.Code)),
				)
			} else {
				attrs = append(attrs,
					attribute.String("tg.rpc.err", "CLIENT"),
				)
			}
			m.failures.Observe(ctx, 1, attrs...)
		}

		return err
	}
}

// object is a abstraction for Telegram API object with TypeName.
type object interface {
	TypeName() string
}

func (m Middleware) attributes(input bin.Encoder) []attribute.KeyValue {
	obj, ok := input.(object)
	if !ok {
		return []attribute.KeyValue{}
	}
	return []attribute.KeyValue{
		attribute.String("tg.method", obj.TypeName()),
	}
}

// New initializes and returns new prometheus middleware.
func New(provider metric.MeterProvider) (*Middleware, error) {
	meter := provider.Meter("github.com/gotd/contrib/oteltg")
	m := &Middleware{}
	var err error
	if m.count, err = meter.AsyncInt64().Counter("tg.rpc.count"); err != nil {
		return nil, err
	}
	if m.failures, err = meter.AsyncInt64().Counter("tg.rpc.failures"); err != nil {
		return nil, err
	}
	if m.duration, err = meter.SyncFloat64().Histogram("tg.rpc.duration"); err != nil {
		return nil, err
	}
	return m, nil
}
