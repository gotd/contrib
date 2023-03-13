// Package clock wraps clock sources.
package clock

import (
	"fmt"
	"time"

	"github.com/gotd/td/clock"

	"github.com/beevik/ntp"
)

var _ clock.Clock = (*ntpClock)(nil)

const defaultNTP = "pool.ntp.org"

type ntpClock struct {
	offset time.Duration
}

func (n *ntpClock) Now() time.Time {
	return time.Now().Add(n.offset)
}

func (n *ntpClock) Timer(d time.Duration) clock.Timer {
	return clock.System.Timer(d)
}

func (n *ntpClock) Ticker(d time.Duration) clock.Ticker {
	return clock.System.Ticker(d)
}

// NewNTP creates new NTP clock.
func NewNTP(ntpHost ...string) (clock.Clock, error) {
	var host string
	switch len(ntpHost) {
	case 0:
		host = defaultNTP
	case 1:
		host = ntpHost[0]
	default:
		return nil, fmt.Errorf("too many ntp hosts")
	}

	resp, err := ntp.Query(host)
	if err != nil {
		return nil, err
	}

	return &ntpClock{
		offset: resp.ClockOffset,
	}, nil
}
