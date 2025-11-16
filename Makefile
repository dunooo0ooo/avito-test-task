ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif

APP_NAME       := avito-test-task
APP_BIN        := bin/$(APP_NAME)
MIGRATOR_BIN   := bin/migrator

POSTGRES_HOST      ?= 127.0.0.1
POSTGRES_PORT      ?= 5432
POSTGRES_USER      ?= postgres
POSTGRES_PASSWORD  ?= postgres
POSTGRES_DB        ?= app_db
POSTGRES_SSLMODE   ?= disable
POSTGRES_DB_TEST ?= app_db_test
MIGRATIONS_DIR     ?= ./migrations

HTTP_ADDR          ?= :8080

DB_DSN = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)
DB_DSN_TEST = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB_TEST)?sslmode=$(POSTGRES_SSLMODE)

.PHONY: all
all: build

.PHONY: build
build:
	@mkdir -p bin
	go build -o $(APP_BIN) ./cmd/app
	go build -o $(MIGRATOR_BIN) ./cmd/migrator

.PHONY: run
run: build
	POSTGRES_HOST=$(POSTGRES_HOST) \
	POSTGRES_PORT=$(POSTGRES_PORT) \
	POSTGRES_USER=$(POSTGRES_USER) \
	POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	POSTGRES_DB=$(POSTGRES_DB) \
	POSTGRES_SSLMODE=$(POSTGRES_SSLMODE) \
	MIGRATIONS_DIR=$(MIGRATIONS_DIR) \
	HTTP_ADDR=$(HTTP_ADDR) \
	LOG_LEVEL=$${LOG_LEVEL:-debug} \
	$(APP_BIN)

.PHONY: run-go
run-go:
	POSTGRES_HOST=$(POSTGRES_HOST) \
	POSTGRES_PORT=$(POSTGRES_PORT) \
	POSTGRES_USER=$(POSTGRES_USER) \
	POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	POSTGRES_DB=$(POSTGRES_DB) \
	POSTGRES_SSLMODE=$(POSTGRES_SSLMODE) \
	MIGRATIONS_DIR=$(MIGRATIONS_DIR) \
	HTTP_ADDR=$(HTTP_ADDR) \
	LOG_LEVEL=$${LOG_LEVEL:-debug} \
	go run ./cmd/main.go

.PHONY: migrate-up-local
migrate-up-local: build
	POSTGRES_HOST=$(POSTGRES_HOST) \
	POSTGRES_PORT=$(POSTGRES_PORT) \
	POSTGRES_USER=$(POSTGRES_USER) \
	POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	POSTGRES_DB=$(POSTGRES_DB) \
	POSTGRES_SSLMODE=$(POSTGRES_SSLMODE) \
	MIGRATIONS_DIR=$(MIGRATIONS_DIR) \
	LOG_LEVEL=$${LOG_LEVEL:-info} \
	$(MIGRATOR_BIN)

.PHONY: migrate-up-local-go
migrate-up-local-go:
	POSTGRES_HOST=$(POSTGRES_HOST) \
	POSTGRES_PORT=$(POSTGRES_PORT) \
	POSTGRES_USER=$(POSTGRES_USER) \
	POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	POSTGRES_DB=$(POSTGRES_DB) \
	POSTGRES_SSLMODE=$(POSTGRES_SSLMODE) \
	MIGRATIONS_DIR=$(MIGRATIONS_DIR) \
	LOG_LEVEL=$${LOG_LEVEL:-info} \
	$(GO) run ./cmd/migrator

.PHONY: test
test:
	go test ./... -cover

.PHONY: test-race
test-race:
	go test -race ./...

.PHONY: gen-mocks
gen-mocks:
	mockgen -source=internal/user/domain/repository.go \
		-destination=internal/user/mocks/user_repository_mock.go \
		-package=mocks

	mockgen -source=internal/pullrequest/domain/repository.go \
		-destination=internal/pullrequest/mocks/pullrequest_repository_mock.go \
		-package=mocks

	mockgen -source=internal/team/domain/repository.go \
		-destination=internal/team/mocks/team_repository_mock.go \
		-package=mocks

	mockgen -source=internal/user/delivery/http/handler.go \
		-destination=internal/user/mocks/user_service_mock.go \
		-package=mocks

	mockgen -source=internal/pullrequest/delivery/http/handler.go \
		-destination=internal/pullrequest/mocks/pullrequest_service_mock.go \
		-package=mocks

	mockgen -source=internal/team/delivery/http/handler.go \
		-destination=internal/team/mocks/team_service_mock.go \
		-package=mocks

test-integration:
	go test ./tests/integration/... -tags=integration -v
