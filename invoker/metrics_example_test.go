package invoker_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/go-faster/errors"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"
	"go.uber.org/multierr"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"

	"github.com/gotd/contrib/invoker"
)

func metricsInvoker(ctx context.Context) (rerr error) {
	prom := prometheus.NewReporter(prometheus.Options{})
	client, err := telegram.ClientFromEnvironment(telegram.Options{})
	if err != nil {
		return errors.Errorf("create client: %w", err)
	}

	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Prefix:         "my_gotd_service",
		Tags:           map[string]string{},
		CachedReporter: prom,
		Separator:      prometheus.DefaultSeparator,
	}, 1*time.Second)
	defer func() {
		multierr.AppendInto(&rerr, closer.Close())
	}()

	r := invoker.NewMetrics(client, scope)
	raw := tg.NewClient(r)
	s := message.NewSender(raw)

	return client.Run(ctx, func(ctx context.Context) error {
		_, err := s.Self().Text(ctx, "hello")
		return err
	})
}

func ExampleMetrics() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := metricsInvoker(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
