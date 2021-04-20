test:
	@./go.test.sh
.PHONY: test

e2e_redis_test:
	@./go.e2e.sh ./redis/...
.PHONY: e2e_redis_test

e2e_vault_test:
	@./go.e2e.sh ./vault/...
.PHONY: e2e_vault_test

e2e_etcd_test:
	@./go.e2e.sh ./etcd/...
.PHONY: e2e_etcd_test

coverage:
	@./go.coverage.sh
.PHONY: coverage

generate:
	go generate ./...
.PHONY: generate
