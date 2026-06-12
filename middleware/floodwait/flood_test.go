package floodwait

import (
	"testing"
	"time"

	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"

	"github.com/gotd/td/tgerr"
)

func TestAsWait(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		want        time.Duration
		wantPerType bool
		wantOK      bool
	}{
		{
			name:        "FloodWait",
			err:         tgerr.New(420, "FLOOD_WAIT_30"),
			want:        30 * time.Second,
			wantPerType: true,
			wantOK:      true,
		},
		{
			name:        "FloodPremiumWait",
			err:         tgerr.New(420, "FLOOD_PREMIUM_WAIT_10"),
			want:        10 * time.Second,
			wantPerType: true,
			wantOK:      true,
		},
		{
			name:   "SlowmodeWait",
			err:    tgerr.New(420, "SLOWMODE_WAIT_5"),
			want:   5 * time.Second,
			wantOK: true,
		},
		{
			name:   "TwoFAConfirmWait",
			err:    tgerr.New(420, "2FA_CONFIRM_WAIT_3600"),
			want:   3600 * time.Second,
			wantOK: true,
		},
		{
			name:   "TakeoutInitDelay",
			err:    tgerr.New(420, "TAKEOUT_INIT_DELAY_60"),
			want:   60 * time.Second,
			wantOK: true,
		},
		{
			name:   "FloodTestPhoneWait",
			err:    tgerr.New(420, "FLOOD_TEST_PHONE_WAIT_5"),
			want:   5 * time.Second,
			wantOK: true,
		},
		{
			name:        "ClampLow",
			err:         tgerr.New(420, "FLOOD_WAIT_0"),
			want:        minFloodWait,
			wantPerType: true,
			wantOK:      true,
		},
		{
			name: "ClampHigh",
			// Above maxWaitSeconds but within a 32-bit int so the argument
			// parses on 386 as well as 64-bit platforms.
			err:         tgerr.New(420, "FLOOD_WAIT_99999999"),
			want:        maxFloodWait,
			wantPerType: true,
			wantOK:      true,
		},
		{
			name:   "FloodSkipFailedWait",
			err:    tgerr.New(420, "FLOOD_SKIP_FAILED_WAIT"),
			want:   minFloodWait,
			wantOK: true,
		},
		{
			name:   "WorkerBusy",
			err:    tgerr.New(500, "WORKER_BUSY_TOO_LONG_RETRY"),
			want:   minFloodWait,
			wantOK: true,
		},
		{
			name:   "OtherInternal",
			err:    tgerr.New(500, "AUTH_KEY_UNREGISTERED"),
			wantOK: false,
		},
		{
			name:   "OtherFlood",
			err:    tgerr.New(420, "SOMETHING_ELSE"),
			wantOK: false,
		},
		{
			name:   "NotRPC",
			err:    errors.New("boom"),
			wantOK: false,
		},
		{
			name:   "Nil",
			err:    nil,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, perType, ok := asWait(tt.err)
			require.Equal(t, tt.wantOK, ok)
			require.Equal(t, tt.wantPerType, perType)
			if tt.wantOK {
				require.Equal(t, tt.want, d)
			}
		})
	}
}
