# contrib

[![Go](https://github.com/gotd/contrib/workflows/CI/badge.svg)](https://github.com/gotd/contrib/actions)
[![Documentation](https://godoc.org/github.com/gotd/contrib?status.svg)](https://pkg.go.dev/github.com/gotd/contrib)
[![license](https://img.shields.io/github/license/gotd/contrib.svg?maxAge=2592000)](https://github.com/gotd/contrib/blob/master/LICENSE)

Companion packages for [gotd](https://github.com/gotd/td) — the Telegram MTProto
client in Go. These are optional, batteries-included helpers that bring in
heavier third-party dependencies (databases, object stores, OpenTelemetry, NTP,
…) and therefore live outside the core `gotd/td` module so the core stays
dependency-light.

## Install

```bash
go get github.com/gotd/contrib
```

Each package is independent — importing one only pulls in the dependencies that
package actually needs.

## Packages

### Running & lifecycle

| Package | Description |
| --- | --- |
| [`bg`](https://pkg.go.dev/github.com/gotd/contrib/bg) | Run a client in the background. `Connect` blocks until the client is connected and ready, then returns a `StopFunc` you call to shut it down — handy when `client.Run`'s callback style does not fit your control flow. Supports `WithContext` and `WithStartupTimeout`. |

### Middleware & RPC

| Package | Description |
| --- | --- |
| [`middleware/floodwait`](https://pkg.go.dev/github.com/gotd/contrib/middleware/floodwait) | Catches Telegram `FLOOD_WAIT` errors and retries transparently. `Waiter` is a scheduler-based implementation for long-running, concurrent programs (wrap your run loop with `Waiter.Run`); `SimpleWaiter` is a timer-based variant for one-off scripts. Both support `WithMaxRetries`/`WithMaxWait`. |
| [`middleware/ratelimit`](https://pkg.go.dev/github.com/gotd/contrib/middleware/ratelimit) | Token-bucket rate limiter (`golang.org/x/time/rate`) that paces outgoing requests to stay under Telegram's limits. Pairs naturally with `floodwait`. |
| [`invoker`](https://pkg.go.dev/github.com/gotd/contrib/invoker) | RPC invoker helpers and middlewares, including a debug invoker and an update-aware invoker. |
| [`oteltg`](https://pkg.go.dev/github.com/gotd/contrib/oteltg) | OpenTelemetry instrumentation for gotd: traces and metrics for outgoing RPCs. |

### Authentication

| Package | Description |
| --- | --- |
| [`auth`](https://pkg.go.dev/github.com/gotd/contrib/auth) | Interfaces, implementations and utilities for `telegram.UserAuthenticator` — read credentials from constructors/env, ask interactively, and compose sign-up flows. |
| [`auth/terminal`](https://pkg.go.dev/github.com/gotd/contrib/auth/terminal) | Terminal-based `UserAuthenticator` that prompts for phone, code, password and sign-up info. Uses an interactive terminal when stdin is a tty and falls back to a buffered reader for pipes, files and CI. |
| [`auth/dialog`](https://pkg.go.dev/github.com/gotd/contrib/auth/dialog) | Compose an authenticator from individual dialog functions. |
| [`auth/kv`](https://pkg.go.dev/github.com/gotd/contrib/auth/kv) | Credential/session helpers built over a generic key-value store. |
| [`auth/localization`](https://pkg.go.dev/github.com/gotd/contrib/auth/localization) | Localizable prompt strings for the terminal authenticator. |

### Storage — sessions, peers & state

These implement the `telegram.SessionStorage`, peer storage and update-state
storage interfaces against various backends. `storage` defines the common peer
abstractions; the rest are backend implementations.

| Package | Description |
| --- | --- |
| [`storage`](https://pkg.go.dev/github.com/gotd/contrib/storage) | Common peer-storage structures: a `PeerStorage` interface, peer collector, resolver cache and iteration helpers shared by the backends below. |
| [`bbolt`](https://pkg.go.dev/github.com/gotd/contrib/bbolt) | Session, peer and update-state storage backed by [etcd bbolt](https://github.com/etcd-io/bbolt) (embedded). |
| [`pebble`](https://pkg.go.dev/github.com/gotd/contrib/pebble) | Storage backed by [CockroachDB Pebble](https://github.com/cockroachdb/pebble) (embedded LSM). |
| [`redis`](https://pkg.go.dev/github.com/gotd/contrib/redis) | Storage backed by [Redis](https://redis.io). |
| [`s3`](https://pkg.go.dev/github.com/gotd/contrib/s3) | Session storage backed by any S3-compatible object store (MinIO client). |
| [`vault`](https://pkg.go.dev/github.com/gotd/contrib/vault) | Secret/session storage backed by [HashiCorp Vault](https://www.vaultproject.io). |

### I/O & streaming

| Package | Description |
| --- | --- |
| [`tg_io`](https://pkg.go.dev/github.com/gotd/contrib/tg_io) | Partial (ranged) I/O over Telegram — download arbitrary byte ranges of a file. |
| [`partio`](https://pkg.go.dev/github.com/gotd/contrib/partio) | Chunk-based reader/writer primitives that align arbitrary reads/writes to fixed-size chunks. |
| [`http_io`](https://pkg.go.dev/github.com/gotd/contrib/http_io) | HTTP handlers built on the partial I/O primitives, e.g. serving Telegram media over HTTP. |
| [`http_range`](https://pkg.go.dev/github.com/gotd/contrib/http_range) | Parser for HTTP `Range` request headers. |

### Utilities

| Package | Description |
| --- | --- |
| [`clock`](https://pkg.go.dev/github.com/gotd/contrib/clock) | Clock sources, including an NTP-backed clock so MTProto time sync works on hosts with a skewed system clock. |

## Documentation

Per-package API docs are on [pkg.go.dev](https://pkg.go.dev/github.com/gotd/contrib),
and guides live in the [gotd documentation site](https://gotd.dev) (see the
*contrib packages* page under Helpers).

## License

[MIT](LICENSE)
