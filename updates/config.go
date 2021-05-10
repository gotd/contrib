package updates

import (
	"context"

	"go.uber.org/zap"

	"github.com/gotd/td/tg"
)

// RawClient interface contains Telegram RPC calls used by the engine
// for state synchronization.
type RawClient interface {
	UpdatesGetState(ctx context.Context) (*tg.UpdatesState, error)
	UpdatesGetDifference(ctx context.Context, request *tg.UpdatesGetDifferenceRequest) (tg.UpdatesDifferenceClass, error)
	UpdatesGetChannelDifference(ctx context.Context, request *tg.UpdatesGetChannelDifferenceRequest) (tg.UpdatesChannelDifferenceClass, error)
}

// Config is an engine config.
type Config struct {
	RawClient RawClient
	Handler   Handler
	SelfID    int
	IsBot     bool

	// Storage is the place whether the state is stored.
	// In-mem by default.
	Storage Storage
	// Number of workers which handles updates.
	// Min is 5.
	Concurrency int
	Forget      bool
	Logger      *zap.Logger
}

func (cfg *Config) setDefaults() {
	if cfg.RawClient == nil {
		panic("raw client is nil")
	}
	if cfg.Handler == nil {
		panic("handler is nil")
	}
	if cfg.Storage == nil {
		cfg.Storage = newMemStorage()
	}
	if cfg.Concurrency < 5 {
		cfg.Concurrency = 5
	}
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}
}
