// Code generated by 'yaegi extract github.com/gotd/td/telegram/message/peer'. DO NOT EDIT.

package yaegi

import (
	"context"
	"reflect"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

func init() {
	Symbols["github.com/gotd/td/telegram/message/peer"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Channel":              reflect.ValueOf(peer.Channel),
		"Chat":                 reflect.ValueOf(peer.Chat),
		"DefaultResolver":      reflect.ValueOf(peer.DefaultResolver),
		"EntitiesFromResult":   reflect.ValueOf(peer.EntitiesFromResult),
		"EntitiesFromUpdate":   reflect.ValueOf(peer.EntitiesFromUpdate),
		"NewEntities":          reflect.ValueOf(peer.NewEntities),
		"NewLRUResolver":       reflect.ValueOf(peer.NewLRUResolver),
		"OnlyChannel":          reflect.ValueOf(peer.OnlyChannel),
		"OnlyChat":             reflect.ValueOf(peer.OnlyChat),
		"OnlyUser":             reflect.ValueOf(peer.OnlyUser),
		"OnlyUserID":           reflect.ValueOf(peer.OnlyUserID),
		"Plain":                reflect.ValueOf(peer.Plain),
		"Resolve":              reflect.ValueOf(peer.Resolve),
		"ResolveDeeplink":      reflect.ValueOf(peer.ResolveDeeplink),
		"ResolveDomain":        reflect.ValueOf(peer.ResolveDomain),
		"ResolvePhone":         reflect.ValueOf(peer.ResolvePhone),
		"SingleflightResolver": reflect.ValueOf(peer.SingleflightResolver),
		"ToInputChannel":       reflect.ValueOf(peer.ToInputChannel),
		"ToInputUser":          reflect.ValueOf(peer.ToInputUser),
		"User":                 reflect.ValueOf(peer.User),

		// type definitions
		"ConstraintError":    reflect.ValueOf((*peer.ConstraintError)(nil)),
		"DialogKey":          reflect.ValueOf((*peer.DialogKey)(nil)),
		"Entities":           reflect.ValueOf((*peer.Entities)(nil)),
		"EntitySearchResult": reflect.ValueOf((*peer.EntitySearchResult)(nil)),
		"Kind":               reflect.ValueOf((*peer.Kind)(nil)),
		"LRUResolver":        reflect.ValueOf((*peer.LRUResolver)(nil)),
		"Promise":            reflect.ValueOf((*peer.Promise)(nil)),
		"PromiseDecorator":   reflect.ValueOf((*peer.PromiseDecorator)(nil)),
		"Resolver":           reflect.ValueOf((*peer.Resolver)(nil)),

		// interface wrapper definitions
		"_EntitySearchResult": reflect.ValueOf((*_github_com_gotd_td_telegram_message_peer_EntitySearchResult)(nil)),
		"_Resolver":           reflect.ValueOf((*_github_com_gotd_td_telegram_message_peer_Resolver)(nil)),
	}
}

// _github_com_gotd_td_telegram_message_peer_EntitySearchResult is an interface wrapper for EntitySearchResult type
type _github_com_gotd_td_telegram_message_peer_EntitySearchResult struct {
	WMapChats func() tg.ChatClassArray
	WMapUsers func() tg.UserClassArray
}

func (W _github_com_gotd_td_telegram_message_peer_EntitySearchResult) MapChats() tg.ChatClassArray {
	return W.WMapChats()
}
func (W _github_com_gotd_td_telegram_message_peer_EntitySearchResult) MapUsers() tg.UserClassArray {
	return W.WMapUsers()
}

// _github_com_gotd_td_telegram_message_peer_Resolver is an interface wrapper for Resolver type
type _github_com_gotd_td_telegram_message_peer_Resolver struct {
	WResolveDomain func(ctx context.Context, domain string) (tg.InputPeerClass, error)
	WResolvePhone  func(ctx context.Context, phone string) (tg.InputPeerClass, error)
}

func (W _github_com_gotd_td_telegram_message_peer_Resolver) ResolveDomain(ctx context.Context, domain string) (tg.InputPeerClass, error) {
	return W.WResolveDomain(ctx, domain)
}
func (W _github_com_gotd_td_telegram_message_peer_Resolver) ResolvePhone(ctx context.Context, phone string) (tg.InputPeerClass, error) {
	return W.WResolvePhone(ctx, phone)
}
