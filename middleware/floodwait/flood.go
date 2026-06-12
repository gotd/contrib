package floodwait

import (
	"time"

	"github.com/gotd/td/tgerr"
)

// The following error types are treated as retriable wait errors, mirroring
// TDLib's NetQueryDelayer.delay and Telethon's _call logic.
//
// See td/telegram/net/NetQueryDelayer.cpp and
// telethon/client/users.py.
const (
	// errFloodWait is the standard flood wait error (code 420).
	errFloodWait = tgerr.ErrFloodWait
	// errFloodPremiumWait is reported for non-premium accounts hitting
	// premium-only rate limits (code 420).
	errFloodPremiumWait = tgerr.ErrPremiumFloodWait
	// errSlowmodeWait is reported when sending messages too fast in a
	// slow-mode chat (code 420).
	errSlowmodeWait = "SLOWMODE_WAIT"
	// err2FAConfirmWait is reported during the 2FA reset waiting period
	// (code 420).
	err2FAConfirmWait = "2FA_CONFIRM_WAIT"
	// errTakeoutInitDelay is reported when a takeout session must wait
	// before it can start (code 420).
	errTakeoutInitDelay = "TAKEOUT_INIT_DELAY"
	// errFloodTestPhoneWait is reported on test servers/phone numbers (code
	// 420). Recognized by Telethon as FloodTestPhoneWaitError.
	errFloodTestPhoneWait = "FLOOD_TEST_PHONE_WAIT"
	// errFloodSkipFailedWait carries no argument and asks to retry shortly
	// (code 420).
	errFloodSkipFailedWait = "FLOOD_SKIP_FAILED_WAIT"
	// errWorkerBusy is a transient server error (code 500) that should be
	// retried after a small delay.
	errWorkerBusy = "WORKER_BUSY_TOO_LONG_RETRY"
)

// perTypeWaitErrors is the set of code 420 errors that represent a per-method
// rate limit: every request of the same type is throttled, so the waiter may
// proactively delay future requests of that type.
var perTypeWaitErrors = []string{
	errFloodWait,
	errFloodPremiumWait,
}

// localWaitErrors is the set of code 420 errors that carry a wait argument but
// are specific to the target chat or operation rather than the request type.
// Telethon notes "SLOW_MODE_WAIT is chat-specific, not request-specific" and
// does not cache these per CONSTRUCTOR_ID; we likewise retry only the offending
// request without throttling unrelated calls of the same type.
var localWaitErrors = []string{
	errSlowmodeWait,
	err2FAConfirmWait,
	errTakeoutInitDelay,
	errFloodTestPhoneWait,
}

const (
	// minWaitSeconds is the lower bound (in seconds) TDLib clamps wait
	// timeouts to.
	minWaitSeconds = 1
	// maxWaitSeconds is the upper bound (in seconds) TDLib clamps wait
	// timeouts to (14 days).
	maxWaitSeconds = 14 * 24 * 60 * 60

	// minFloodWait is the lower bound TDLib clamps wait timeouts to.
	minFloodWait = minWaitSeconds * time.Second
	// maxFloodWait is the upper bound TDLib clamps wait timeouts to (14 days).
	maxFloodWait = maxWaitSeconds * time.Second
)

// asWait reports the duration to wait before retrying err, following TDLib's
// NetQueryDelayer.delay and Telethon's flood handling. It returns ok=false for
// errors that the waiter should not retry.
//
// perType reports whether the wait applies to every request of the same type (a
// per-method rate limit such as FLOOD_WAIT) rather than just this single
// request. Chat- or operation-specific waits (SLOWMODE_WAIT, 2FA_CONFIRM_WAIT,
// TAKEOUT_INIT_DELAY, FLOOD_TEST_PHONE_WAIT) and transient retries report
// perType=false so unrelated requests of the same type are not throttled.
//
// Compared to tgerr.AsFloodWait it additionally recognizes SLOWMODE_WAIT,
// 2FA_CONFIRM_WAIT, TAKEOUT_INIT_DELAY and FLOOD_TEST_PHONE_WAIT (code 420),
// the argument-less FLOOD_SKIP_FAILED_WAIT (code 420) and
// WORKER_BUSY_TOO_LONG_RETRY (code 500), and clamps the wait into the [1s, 14d]
// range.
func asWait(err error) (d time.Duration, perType, ok bool) {
	rpcErr, ok := tgerr.As(err)
	if !ok {
		return 0, false, false
	}

	switch rpcErr.Code {
	case 420:
		if rpcErr.IsOneOf(perTypeWaitErrors...) {
			return clampWait(rpcErr.Argument), true, true
		}
		if rpcErr.IsOneOf(localWaitErrors...) {
			return clampWait(rpcErr.Argument), false, true
		}
		// FLOOD_SKIP_FAILED_WAIT carries no argument: it is dangerous to
		// resend the query without a timeout, so retry after the minimum.
		if rpcErr.IsType(errFloodSkipFailedWait) {
			return minFloodWait, false, true
		}
		return 0, false, false
	case 500:
		// It is dangerous to resend the query without a timeout, so use 1s.
		if rpcErr.IsType(errWorkerBusy) {
			return minFloodWait, false, true
		}
		return 0, false, false
	default:
		return 0, false, false
	}
}

// clampWait clamps a wait argument expressed in seconds into the
// [minWaitSeconds, maxWaitSeconds] range and converts it to a duration, like
// TDLib's clamp(timeout, 1, 14 * 24 * 60 * 60). Clamping the seconds before
// converting avoids overflowing the nanosecond-based time.Duration.
func clampWait(seconds int) time.Duration {
	if seconds < minWaitSeconds {
		seconds = minWaitSeconds
	}
	if seconds > maxWaitSeconds {
		seconds = maxWaitSeconds
	}
	return time.Duration(seconds) * time.Second
}
