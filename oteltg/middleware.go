// Package oteltg provides OpenTelemetry instrumentation for gotd.
package oteltg

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// Middleware is prometheus metrics middleware for Telegram.
type Middleware struct {
	count    metric.Int64Counter
	failures metric.Int64Counter
	duration metric.Float64Histogram
	tracer   trace.Tracer
}

// Handle implements telegram.Middleware.
func (m Middleware) Handle(next tg.Invoker) telegram.InvokeFunc {
	return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		// Prepare.
		attrs := m.attributes(input)

		spanName := "tg.rpc"
		for _, attr := range attrs {
			if attr.Key == "tg.method" {
				spanName = fmt.Sprintf("%s: %s", spanName, attr.Value.AsString())
			}
		}

		ctx, span := m.tracer.Start(ctx, spanName, trace.WithAttributes(attrs...))
		defer span.End()
		m.count.Add(ctx, 1, metric.WithAttributes(attrs...))
		start := time.Now()

		// Call actual method.
		err := next.Invoke(ctx, input, output)

		// Observe.
		m.duration.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(attrs...))
		if err != nil {
			var errAttrs []attribute.KeyValue
			if rpcErr, ok := tgerr.As(err); ok {
				span.SetStatus(codes.Error, "RPC error")
				errAttrs = append(errAttrs,
					attribute.String("tg.rpc.err", rpcErr.Type),
					attribute.String("tg.rpc.code", strconv.Itoa(rpcErr.Code)),
				)
			} else {
				span.SetStatus(codes.Error, "Internal error")
				errAttrs = append(errAttrs,
					attribute.String("tg.rpc.err", "CLIENT"),
				)
			}
			span.RecordError(err, trace.WithAttributes(errAttrs...))
			attrs = append(attrs, errAttrs...)
			m.failures.Add(ctx, 1, metric.WithAttributes(attrs...))
		} else {
			span.SetStatus(codes.Ok, "")
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
func New(meterProvider metric.MeterProvider, tracerProvider trace.TracerProvider) (*Middleware, error) {
	const name = "github.com/gotd/contrib/oteltg"
	meter := meterProvider.Meter(name)
	m := &Middleware{
		tracer: tracerProvider.Tracer(name),
	}
	var err error
	if m.count, err = meter.Int64Counter("tg.rpc.count"); err != nil {
		return nil, err
	}
	if m.failures, err = meter.Int64Counter("tg.rpc.failures"); err != nil {
		return nil, err
	}
	if m.duration, err = meter.Float64Histogram("tg.rpc.duration"); err != nil {
		return nil, err
	}
	return m, nil
}
