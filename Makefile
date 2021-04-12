test:
	@./go.test.sh
.PHONY: test

e2e_redis_test:
	@./go.e2e.sh ./auth/redis/...
.PHONY: e2e_redis_test

e2e_vault_test:
	@./go.e2e.sh ./auth/vault/...
.PHONY: e2e_vault_test

coverage:
	@./go.coverage.sh
.PHONY: coverage

generate:
	go generate ./...
.PHONY: generate
