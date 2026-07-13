.PHONY: help build run test test-short lint fmt vet tidy clean sqlc-generate check \
    migrate migrate-up migrate-down migrate-create check-migrate-tool check-db-url check-sqlc-tool

BINARY_DIR := bin
BINARY_NAME := api
BINARY_PATH := $(BINARY_DIR)/$(BINARY_NAME)
MIGRATIONS_DIR := ./migrations

.DEFAULT_GOAL := help

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'

## build: compile cmd/api into bin/api
build:
	go build -o $(BINARY_PATH) ./cmd/api

## run: build and run the API server
run: build
	./$(BINARY_PATH)

## test: run the full test suite with the race detector
test:
	go test ./... -race

## test-short: run the full test suite without the race detector (faster local iteration)
test-short:
	go test ./...

## lint: run golangci-lint using the repo's .golangci.yml (same config CI uses)
lint:
	golangci-lint run

## fmt: format the tree with gofmt and goimports
fmt:
	gofmt -w .
	goimports -w .

## vet: run go vet across the tree
vet:
	go vet ./...

## tidy: run go mod tidy
tidy:
	go mod tidy

## clean: remove build artifacts
clean:
	rm -rf $(BINARY_DIR)

## sqlc-generate: regenerate Go code from SQL queries via sqlc
sqlc-generate:
	sqlc generate

check-sqlc-tool:
    @command -v sqlc >/dev/null 2>&1 || { \
        echo "error: 'sqlc' CLI not found in PATH."; \
        echo "install it with: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"; \
        exit 1; \
    }

## check: run the same gate CI runs (vet, lint, race-tested tests)
check: vet lint test

# --- migrations ---------------------------------------------------------

check-migrate-tool:
	@command -v migrate >/dev/null 2>&1 || { \
		echo "error: 'migrate' CLI not found in PATH."; \
		echo "install it with: go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	}

check-db-url:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "error: DATABASE_URL is not set."; \
		echo "export DATABASE_URL=postgres://user:pass@host:port/dbname?sslmode=disable"; \
		exit 1; \
	fi

## migrate: alias for migrate-up
migrate: migrate-up

## migrate-up: apply all up migrations
migrate-up: check-migrate-tool check-db-url
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) up

## migrate-down: roll back the most recent migration
migrate-down: check-migrate-tool check-db-url
	migrate -database "$(DATABASE_URL)" -path $(MIGRATIONS_DIR) down 1

## migrate-create: scaffold a new up/down migration pair
##   usage: make migrate-create NAME=create_users_table
migrate-create: check-migrate-tool
	@if [ -z "$(NAME)" ]; then \
		echo "error: NAME is required."; \
		echo "usage: make migrate-create NAME=create_users_table"; \
		exit 1; \
	fi
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)