package yaegi

import "reflect"

//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/transport

//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/tg
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/tg/e2e

//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/dcs
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/downloader
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/message
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/message/entity
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/message/html
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/message/inline
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/message/markup
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/message/peer
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/message/styling
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/query
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/query/channels/participants
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/query/contacts/blocked
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/query/dialogs
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/query/hasher
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/query/messages
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/query/messages/stickers/featured
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/query/photos
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/uploader

//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/clock

// Symbols variable stores the map of gotd symbols per package.
var Symbols = map[string]map[string]reflect.Value{} // nolint:gochecknoglobals
