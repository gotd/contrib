package main

import (
	"context"
	"reflect"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/tg"
)

func helpers(ctx context.Context, client *telegram.Client) map[string]reflect.Value {
	raw := tg.NewClient(client)
	sender := message.NewSender(raw)
	q := query.NewQuery(raw)
	return map[string]reflect.Value{
		"Client": reflect.ValueOf(client),
		"RPC":    reflect.ValueOf(raw),
		"Ctx":    reflect.ValueOf(ctx),

		"Sender": reflect.ValueOf(sender),
		"Send": reflect.ValueOf(func(msg, to string) error {
			_, err := sender.Resolve(to).Text(ctx, msg)
			return err
		}),
		"HTML": reflect.ValueOf(func(msg, to string) error {
			_, err := sender.Resolve(to).StyledText(ctx, html.String(msg))
			return err
		}),

		"Query": reflect.ValueOf(q),
	}
}
