package yaegi

import "reflect"

//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/transport

//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/tg
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/tg/e2e

//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/uploader
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/telegram/downloader

//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract github.com/gotd/td/clock

// Symbols variable stores the map of gotd symbols per package.
var Symbols = map[string]map[string]reflect.Value{}
