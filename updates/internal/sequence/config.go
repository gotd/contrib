package sequence

import (
	"time"

	"go.uber.org/zap"
)

// Config is the sequence box config.
type Config struct {
	InitialState   int
	FastgapTimeout time.Duration
	Apply          func(state int, updates []interface{}) error
	Logger         *zap.Logger

	hooks hooks
}

func (cfg *Config) setDefaults() {
	if cfg.FastgapTimeout == 0 {
		cfg.FastgapTimeout = time.Millisecond * 500
	}
	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}
	if cfg.hooks == nil {
		cfg.hooks = nopHooks{}
	}
	if cfg.Apply == nil {
		panic("apply function is nil")
	}
}
