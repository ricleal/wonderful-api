# Makefile settings
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

# DB settings
DB_HOSTNAME ?= localhost
DB_PORT ?= 5432
DB_NAME ?= wonderful
DB_USERNAME ?= $(USERNAME)

DB_URL ?= "postgres://$(DB_USERNAME)@$(DB_HOSTNAME):$(DB_PORT)/$(DB_NAME)?sslmode=disable"

MIGRATIONS_PATH ?= $(shell pwd)/migrations

LOG_LEVEL ?= debug
API_PORT ?= 8888

ENV_VARS = \
	DB_HOSTNAME=$(DB_HOSTNAME) \
	DB_PORT=$(DB_PORT) \
	DB_NAME=$(DB_NAME) \
	DB_USERNAME=$(DB_USERNAME) \
	LOG_LEVEL=$(LOG_LEVEL) \
	API_PORT=$(API_PORT) \
	DB_URL=$(DB_URL) \
	$(NULL)


## Development targets

.PHONY: dev
dev: ## Run development server
	DB_URL=$(DB_URL) LOG_LEVEL=$(LOG_LEVEL) go run ./cmd/wonderful -port $(API_PORT)

.PHONY: test
test: ## Run unit and integration tests
	@$(ENV_VARS) MIGRATIONS_PATH=$(MIGRATIONS_PATH) go test -race ./...

# Instalation: brew install golangci-lint
.PHONY: lint
lint: ## Lint and format source code based on golangci configuration
	@command -v golangci-lint || (echo "Please install `golangci-lint`" && exit 1)
	golangci-lint run --fix -v ./...

## DB targets

.PHONY: db-start
db-start: ## Postgres start
	@$(ENV_VARS) docker-compose -f docker-compose-db.yaml -p db up --detach postgres-dev

.PHONY: db-stop
db-stop: ## Postgres stop
	@$(ENV_VARS) docker-compose -f docker-compose-db.yaml -p db stop postgres-dev

.PHONY: db-cli
db-cli: ## Start the Postgres CLI
	@command -v pgcli || (echo "Please install `pgcli`." && exit 1)
	pgcli -h $(DB_HOSTNAME) -u $(DB_USERNAME) -p $(DB_PORT) -d $(DB_NAME)

### DB migration targets

# https://github.com/golang-migrate/migrate
# brew install golang-migrate
db-migrate-up: ## Run database upgrade migrations
	migrate -database $(DB_URL) -path migrations up

db-migrate-down:  ## Run database downgrade the last migration
	migrate -database $(DB_URL) -path migrations down 1

db-migrate-version:  ## Print the current migration version
	migrate -database $(DB_URL) -path migrations version

db-migrate-create:  ## Create a new migration file
	@if [ -z "$(name)" ]; then echo "name is required"; exit 1; fi
	migrate create -ext sql -dir migrations -seq $(name)

db-migrate-force:  ## Force mark the migration version
	@if [ -z "$(version)" ]; then echo "version is required"; exit 1; fi
	migrate -database $(DB_URL) -path migrations force $(version)

#### Code generation ####

## OpenAPI targets
# Install: go install "github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest"
.PHONY: openapi-generate
openapi-generate: ## Generate OpenAPI client
	go version
	mkdir -p internal/api/v1/openapi
	rm -rf internal/api/v1/openapi/*
	oapi-codegen \
		-generate types \
		-package openapi \
		-o internal/api/v1/openapi/types.go \
		open-api/v1.yaml
	oapi-codegen \
		-generate chi-server \
		-package openapi \
		-o internal/api/v1/openapi/router.go \
		open-api/v1.yaml
	oapi-codegen \
		-generate spec \
		-package openapi \
		-o internal/api/v1/openapi/spec.go \
		open-api/v1.yaml

## DB MODEL targets
# go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
.PHONY: db-models
db-models: ## Generate Go database models
	sqlc generate

.PHONY: help
help:
	@grep -hE '^[a-zA-Z_-][0-9a-zA-Z_-]*:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

#### Docker targets ####

.PHONY: docker-up
docker-up: ## Run docker container
	@$(ENV_VARS) docker-compose -f docker-compose.yaml up --build

.PHONY: docker-down
docker-down: ## Stop docker container
	@$(ENV_VARS) docker-compose -f docker-compose.yaml down