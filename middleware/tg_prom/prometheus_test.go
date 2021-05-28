package tg_prom

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestPrometheus(t *testing.T) {
	r := prometheus.NewPedanticRegistry()
	p := New()

	for _, m := range p.Metrics() {
		require.NoError(t, r.Register(m))
	}
}
