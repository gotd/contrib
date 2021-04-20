module github.com/gotd/contrib

go 1.16

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/cockroachdb/pebble v0.0.0-20210406003833-3d4c32f510a8
	github.com/go-redis/redis/v8 v8.8.2
	github.com/gotd/neo v0.1.3
	github.com/gotd/td v0.37.0
	github.com/hashicorp/vault/api v1.1.0
	github.com/stretchr/testify v1.7.0
	github.com/traefik/yaegi v0.9.17
	go.uber.org/multierr v1.6.0
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
)
