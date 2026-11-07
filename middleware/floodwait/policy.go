package floodwait

import (
	"context"
	"time"
)

// Policy overrides Waiter limits for requests made with a context, see
// WithPolicy.
type Policy struct {
	// MaxWait limits wait time per attempt, overriding Waiter.WithMaxWait.
	// Zero keeps the Waiter limit, negative means no limit ("background: wait
	// as long as it takes").
	MaxWait time.Duration
	// MaxRetries limits number of retries, overriding Waiter.WithMaxRetries.
	// Zero keeps the Waiter limit, negative means no limit.
	MaxRetries int
	// NoRetry makes wait errors returned immediately ("interactive: fail
	// fast"). Scheduling delays of the request type still apply.
	NoRetry bool
}

type policyKey struct{}

// WithPolicy returns context with Policy for requests made with it.
func WithPolicy(ctx context.Context, p Policy) context.Context {
	return context.WithValue(ctx, policyKey{}, p)
}

func policyFrom(ctx context.Context) Policy {
	p, _ := ctx.Value(policyKey{}).(Policy)
	return p
}

// limit returns effective limit: override if set (negative means none),
// default otherwise.
func limit[T time.Duration | int](override, def T) T {
	switch {
	case override < 0:
		return 0
	case override > 0:
		return override
	default:
		return def
	}
}
