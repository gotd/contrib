package invoker

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"time"

	"go.uber.org/multierr"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tdp"
	"github.com/gotd/td/tg"
)

// Debug is pretty-print debugging invoker middleware.
type Debug struct {
	next tg.Invoker
	out  io.Writer
}

// NewDebug creates new Debug middleware.
func NewDebug(next tg.Invoker) *Debug {
	return &Debug{next: next}
}

// WithOutput sets output writer.
func (d *Debug) WithOutput(out io.Writer) *Debug {
	d.out = out
	return d
}

func formatObject(input interface{}) string {
	o, ok := input.(tdp.Object)
	if !ok {
		// Handle tg.*Box values.
		rv := reflect.Indirect(reflect.ValueOf(input))
		for i := 0; i < rv.NumField(); i++ {
			if v, ok := rv.Field(i).Interface().(tdp.Object); ok {
				return formatObject(v)
			}
		}

		return fmt.Sprintf("%T (not object)", input)
	}
	return tdp.Format(o)
}

// Invoke implements tg.Invoker.
func (d *Debug) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	_, rerr := fmt.Fprintln(d.out, "→", formatObject(input))

	start := time.Now()
	if err := d.next.Invoke(ctx, input, output); err != nil {
		rerr = multierr.Append(rerr, err)
		_, err := fmt.Fprintln(d.out, "←", err)
		return multierr.Append(rerr, err)
	}

	_, err := fmt.Fprintf(d.out,
		"← (%s) %s\n",
		time.Since(start).Round(time.Millisecond),
		formatObject(output),
	)
	return multierr.Append(rerr, err)
}
