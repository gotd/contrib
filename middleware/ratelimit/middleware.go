package ratelimit

import (
	"golang.org/x/time/rate"

	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/middleware"
)

// MiddlewareOption configures new RateLimiter in middleware constructor.
type MiddlewareOption func(r *RateLimiter) *RateLimiter

// Middleware returns a new RateLimiter middleware constructor.
func Middleware(lim *rate.Limiter, opts ...MiddlewareOption) middleware.Middleware {
	return func(invoker tg.Invoker) tg.Invoker {
		limiter := NewRateLimiter(invoker, lim)
		for _, f := range opts {
			limiter = f(limiter)
		}
		return limiter
	}
}
