module github.com/gotd/contrib

go 1.16

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/cockroachdb/pebble v0.0.0-20210414141038-bee0c60e96bc
	github.com/etcd-io/bbolt v1.3.3
	github.com/gen2brain/dlgs v0.0.0-20210406143744-f512297a108e
	github.com/go-redis/redis/v8 v8.8.2
	github.com/gotd/neo v0.1.3
	github.com/gotd/td v0.37.0
	github.com/hashicorp/vault/api v1.1.0
	github.com/m3db/prometheus_client_golang v0.8.1 // indirect
	github.com/m3db/prometheus_client_model v0.1.0 // indirect
	github.com/m3db/prometheus_common v0.1.0 // indirect
	github.com/m3db/prometheus_procfs v0.8.1 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/traefik/yaegi v0.9.17
	github.com/uber-go/tally v3.3.17+incompatible
	go.etcd.io/bbolt v1.3.5 // indirect
	go.etcd.io/etcd/client/v3 v3.5.0-alpha.0
	go.uber.org/multierr v1.6.0
	golang.org/x/term v0.0.0-20210422114643-f5beecf764ed
	golang.org/x/text v0.3.6
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible => github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
)
