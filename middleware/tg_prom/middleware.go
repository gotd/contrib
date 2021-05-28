// Package tg_prom implements middleware for prometheus metrics.
package tg_prom

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// Middleware is prometheus metrics middleware for Telegram.
type Middleware struct {
	count    *prometheus.CounterVec
	failures *prometheus.CounterVec
	duration prometheus.ObserverVec
}

// Metrics returns slice of provided prometheus metrics.
func (m Middleware) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		m.count,
		m.failures,
		m.duration,
	}
}

const (
	labelErrType = "tg_err_type"
	labelErrCode = "tg_err_code"
	labelMethod  = "tg_method"
)

// Handle implements telegram.Middleware.
func (m Middleware) Handle(next tg.Invoker) telegram.InvokeFunc {
	return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
		// Prepare.
		labels := m.labels(input)
		m.count.With(labels).Inc()
		start := time.Now()

		// Call actual method.
		err := next.Invoke(ctx, input, output)

		// Observe.
		m.duration.With(labels).Observe(time.Since(start).Seconds())
		if err != nil {
			failureLabels := prometheus.Labels{}
			for k, v := range labels {
				failureLabels[k] = v
			}
			if rpcErr, ok := tgerr.As(err); ok {
				failureLabels[labelErrType] = rpcErr.Type
				failureLabels[labelErrCode] = strconv.Itoa(rpcErr.Code)
			} else {
				failureLabels[labelErrType] = "CLIENT"
			}
			m.failures.With(failureLabels)
		}

		return err
	}
}

// object is a abstraction for Telegram API object with TypeName.
type object interface {
	TypeName() string
}

func (m Middleware) labels(input bin.Encoder) prometheus.Labels {
	obj, ok := input.(object)
	if !ok {
		return prometheus.Labels{}
	}
	return prometheus.Labels{
		labelMethod: obj.TypeName(),
	}
}

// New initializes and returns new prometheus middleware.
func New() *Middleware {
	return &Middleware{
		count: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tg_rpc_count_total",
			Help: "Telegram RPC calls total count.",
		}, []string{labelMethod}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "tg_rpc_duration_seconds",
			Help: "Telegram RPC calls duration histogram.",
		}, []string{labelMethod}),
		failures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tg_rpc_failures_total",
			Help: "Telegram failed RPC calls total count.",
		}, []string{labelMethod, labelErrCode, labelErrType}),
	}
}
