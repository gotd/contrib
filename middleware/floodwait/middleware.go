package floodwait

import (
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/middleware"
)

// MiddlewareOption configures new SimpleWaiter in middleware constructor.
type MiddlewareOption func(w *SimpleWaiter) *SimpleWaiter

// Middleware returns a new SimpleWaiter middleware constructor.
func Middleware(opts ...MiddlewareOption) middleware.Middleware {
	return func(invoker tg.Invoker) tg.Invoker {
		waiter := NewSimpleWaiter(invoker)
		for _, f := range opts {
			waiter = f(waiter)
		}
		return waiter
	}
}
