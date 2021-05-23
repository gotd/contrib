package invoker

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// UpdateHookFn is called on each tg.UpdatesClass method result.
//
// Function is called before invoker return. Returned error will be wrapped
// and returned as InvokeRaw result.
type UpdateHookFn func(ctx context.Context, u tg.UpdatesClass) error

// UpdateHook calls hook for each tg.UpdatesClass result.
type UpdateHook struct {
	hook func(ctx context.Context, u tg.UpdatesClass) error
	next tg.Invoker
}

// Invoke implements tg.Invoker.
func (h UpdateHook) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	if err := h.next.Invoke(ctx, input, output); err != nil {
		return err
	}
	if u, ok := output.(*tg.UpdatesBox); ok {
		if err := h.hook(ctx, u.Updates); err != nil {
			return xerrors.Errorf("hook: %w", err)
		}
	}

	return nil
}

// NewUpdateHook creates new update hook middleware.
//
// The fn callback is called on each successful invocation of method
// with the tg.UpdatesClass result.
//
// See UpdateHookFn.
func NewUpdateHook(next tg.Invoker, fn UpdateHookFn) *UpdateHook {
	return &UpdateHook{
		hook: fn,
		next: next,
	}
}
