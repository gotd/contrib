module github.com/gotd/contrib

go 1.21

require (
	github.com/beevik/ntp v1.3.1
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/cockroachdb/pebble v1.1.0
	github.com/gen2brain/dlgs v0.0.0-20211108104213-bade24837f0b
	github.com/go-faster/errors v0.7.1
	github.com/go-faster/jx v1.1.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gotd/neo v0.1.5
	github.com/gotd/td v0.97.0
	github.com/hashicorp/vault/api v1.12.0
	github.com/minio/minio-go/v7 v7.0.67
	github.com/prometheus/client_golang v1.18.0
	github.com/stretchr/testify v1.8.4
	go.etcd.io/bbolt v1.3.8
	go.opentelemetry.io/otel v1.23.1
	go.opentelemetry.io/otel/metric v1.23.1
	go.opentelemetry.io/otel/trace v1.23.1
	go.uber.org/atomic v1.11.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.26.0
	golang.org/x/sync v0.6.0
	golang.org/x/term v0.17.0
	golang.org/x/text v0.14.0
	golang.org/x/time v0.5.0
)

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v3 v3.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cockroachdb/errors v1.11.1 // indirect
	github.com/cockroachdb/logtags v0.0.0-20230118201751-21c54148d20b // indirect
	github.com/cockroachdb/redact v1.1.5 // indirect
	github.com/cockroachdb/tokenbucket v0.0.0-20230807174530-cc333fc44b06 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/getsentry/sentry-go v0.18.0 // indirect
	github.com/go-faster/xor v1.0.0 // indirect
	github.com/go-jose/go-jose/v3 v3.0.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/gotd/ige v0.2.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.6 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.5 // indirect
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/rs/xid v1.5.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/exp v0.0.0-20230626212559-97b1e661b5df // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	nhooyr.io/websocket v1.8.10 // indirect
	rsc.io/qr v0.2.0 // indirect
)

replace (
	github.com/coreos/etcd v3.3.10+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible => github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/gorilla/websocket v1.4.0 => github.com/gorilla/websocket v1.4.2
)
