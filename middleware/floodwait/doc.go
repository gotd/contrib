// Package floodwait implements a tg.Invoker that handles flood wait errors.
//
// It provides two implementations: the scheduler-based Waiter, suitable for
// long-running, highly concurrent programs, and the simpler timer-based
// SimpleWaiter, suitable for one-off tasks.
//
// # Recognized wait errors
//
// Following TDLib's NetQueryDelayer and Telethon's flood handling, both waiters
// retry on a broader set of "wait" errors than just FLOOD_WAIT:
//
//   - FLOOD_WAIT (420) and FLOOD_PREMIUM_WAIT (420): per-method rate limits.
//   - SLOWMODE_WAIT, 2FA_CONFIRM_WAIT, TAKEOUT_INIT_DELAY,
//     FLOOD_TEST_PHONE_WAIT (420): chat- or operation-specific waits.
//   - FLOOD_SKIP_FAILED_WAIT (420): argument-less, retried after 1s.
//   - WORKER_BUSY_TOO_LONG_RETRY (500): transient, retried after 1s.
//
// The reported wait duration is clamped into the [1s, 14d] range, matching
// TDLib's clamp(timeout, 1, 14*24*60*60).
//
// # Per-method versus request-specific waits
//
// FLOOD_WAIT and FLOOD_PREMIUM_WAIT throttle every request of the same type, so
// the scheduler-based Waiter proactively delays future requests of that type
// after one is seen. The remaining wait errors are tied to a specific chat or
// operation rather than the method (Telethon notes "SLOW_MODE_WAIT is
// chat-specific, not request-specific"), so the Waiter retries only the
// offending request and leaves unrelated requests of the same type untouched.
//
// # Limits
//
// Both waiters bound retries via WithMaxRetries and the per-attempt wait via
// WithMaxWait: a wait longer than the limit is returned as an error instead of
// being slept on. Use a context with a deadline to bound the total wait time.
package floodwait
