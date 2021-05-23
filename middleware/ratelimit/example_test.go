package ratelimit_test

import (
	"time"

	"golang.org/x/time/rate"

	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/middleware/ratelimit"
)

func ExampleRateLimiter() {
	var invoker tg.Invoker // e.g. *telegram.Client

	limiter := ratelimit.NewRateLimiter(invoker,
		rate.NewLimiter(rate.Every(100*time.Millisecond), 1),
	)

	tg.NewClient(limiter)
}
