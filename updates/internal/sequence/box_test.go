package sequence

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/xerrors"
)

func nopApply(int, []interface{}) error { return nil }

func TestFastgapResolve(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		m   = NewMockedHooks(ctrl)
		box = New(Config{
			InitialState: 1,
			Logger:       zaptest.NewLogger(t),
			Apply:        nopApply,
			hooks:        m,
		})
	)

	gomock.InOrder(
		m.EXPECT().Apply(2, []interface{}{"a"}),
		m.EXPECT().Transition(modeNormal, modeFastgap),
		m.EXPECT().FastgapBegin().Do(func() {
			// Fill the gap.
			err := box.Handle(Update{
				Value: "b",
				State: 4,
				Count: 2,
			})
			require.NoError(t, err)
		}),

		m.EXPECT().FastgapCollectorFinished(),
		m.EXPECT().Transition(modeFastgap, modeNormal),
		m.EXPECT().Apply(5, []interface{}{"b", "c"}),
		m.EXPECT().FastgapEnd(),
	)

	// Apply update.
	err := box.Handle(Update{
		Value: "a",
		State: 2,
		Count: 1,
	})
	require.NoError(t, err)

	// Trigger fastgap mode.
	err = box.Handle(Update{
		Value: "c",
		State: 5,
		Count: 1,
	})
	require.NoError(t, err)
}

func TestFastgapTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		m   = NewMockedHooks(ctrl)
		box = New(Config{
			InitialState:   1,
			FastgapTimeout: 1,
			Logger:         zaptest.NewLogger(t),
			Apply:          nopApply,
			hooks:          m,
		})
	)

	gomock.InOrder(
		m.EXPECT().Apply(2, []interface{}{"a"}),
		m.EXPECT().Transition(modeNormal, modeFastgap),
		m.EXPECT().FastgapBegin(),
		m.EXPECT().FastgapCollectorFinished(),
		m.EXPECT().Transition(modeFastgap, modeBuffer),
		m.EXPECT().FastgapEnd(),
	)

	// Apply update.
	err := box.Handle(Update{
		Value: "a",
		State: 2,
		Count: 1,
	})
	require.NoError(t, err)

	// Trigger
	err = box.Handle(Update{
		Value: "c",
		State: 5,
		Count: 1,
	})
	require.ErrorIs(t, err, ErrGap)
}

func TestFastgapResolveForcedBuffer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		m   = NewMockedHooks(ctrl)
		box = New(Config{
			InitialState: 1,
			Logger:       zaptest.NewLogger(t),
			Apply:        nopApply,
			hooks:        m,
		})
		onFGEnd = make(chan struct{})
	)

	gomock.InOrder(
		m.EXPECT().Apply(2, []interface{}{"a"}),
		m.EXPECT().Transition(modeNormal, modeFastgap),
		m.EXPECT().FastgapBegin().Do(func() {
			// Manually do what EnableBuffering does
			// but in a non-blocking way.
			box.mux.Lock()
			require.Equal(t, modeFastgap, box.mode)
			box.fgForceBuf = true
			box.fgWaiters = append(box.fgWaiters, onFGEnd)
			box.mux.Unlock()

			// Resolve fastgap.
			err := box.Handle(Update{
				Value: "b",
				State: 4,
				Count: 2,
			})
			require.NoError(t, err)
		}),
		m.EXPECT().FastgapCollectorFinished(),
		m.EXPECT().Transition(modeFastgap, modeBuffer),
		m.EXPECT().Apply(5, []interface{}{"b", "c"}),
		m.EXPECT().FastgapEnd().Do(func() {
			// Make sure that fastgap goroutine
			// closed all channels inside 'fgWaiters' slice.
			<-onFGEnd
			require.Equal(t, modeBuffer, box.mode)
		}),
	)

	// Apply update.
	err := box.Handle(Update{
		Value: "a",
		State: 2,
		Count: 1,
	})
	require.NoError(t, err)

	// Trigger fastgap mode.
	err = box.Handle(Update{
		Value: "c",
		State: 5,
		Count: 1,
	})
	require.NoError(t, err)
}

func TestResultError(t *testing.T) {
	myError := xerrors.Errorf("foobar")

	box := New(Config{
		InitialState: 1,
		Logger:       zaptest.NewLogger(t),
		Apply: func(int, []interface{}) error {
			return &ResultError{myError}
		},
	})

	err := box.Handle(Update{
		State: 2,
		Count: 1,
	})
	require.ErrorIs(t, err, myError)
	require.Equal(t, 2, box.state)
}
