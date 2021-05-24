// Code generated by 'yaegi extract github.com/gotd/td/telegram/query/channels/participants'. DO NOT EDIT.

package yaegi

import (
	"context"
	"reflect"

	"github.com/gotd/td/telegram/query/channels/participants"
	"github.com/gotd/td/tg"
)

func init() {
	Symbols["github.com/gotd/td/telegram/query/channels/participants"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"NewIterator":     reflect.ValueOf(participants.NewIterator),
		"NewQueryBuilder": reflect.ValueOf(participants.NewQueryBuilder),

		// type definitions
		"Elem":                        reflect.ValueOf((*participants.Elem)(nil)),
		"GetParticipantsQueryBuilder": reflect.ValueOf((*participants.GetParticipantsQueryBuilder)(nil)),
		"Iterator":                    reflect.ValueOf((*participants.Iterator)(nil)),
		"Query":                       reflect.ValueOf((*participants.Query)(nil)),
		"QueryBuilder":                reflect.ValueOf((*participants.QueryBuilder)(nil)),
		"QueryFunc":                   reflect.ValueOf((*participants.QueryFunc)(nil)),
		"Request":                     reflect.ValueOf((*participants.Request)(nil)),

		// interface wrapper definitions
		"_Query": reflect.ValueOf((*_github_com_gotd_td_telegram_query_channels_participants_Query)(nil)),
	}
}

// _github_com_gotd_td_telegram_query_channels_participants_Query is an interface wrapper for Query type
type _github_com_gotd_td_telegram_query_channels_participants_Query struct {
	WQuery func(ctx context.Context, req participants.Request) (tg.ChannelsChannelParticipantsClass, error)
}

func (W _github_com_gotd_td_telegram_query_channels_participants_Query) Query(ctx context.Context, req participants.Request) (tg.ChannelsChannelParticipantsClass, error) {
	return W.WQuery(ctx, req)
}
