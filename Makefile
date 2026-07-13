.PHONY: help build run version test vet clean

BINARY_DIR := bin
BINARY_NAME := goforge
BINARY_PATH := $(BINARY_DIR)/$(BINARY_NAME)

LDFLAGS := -X github.com/devaraja-anu/goforge/internal/cli.Version=$(VERSION)

.DEFAULT_GOAL := help

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'

## build: compile cmd/goforge into bin/goforge, embedding VERSION (default: dev)
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) ./cmd/goforge

## run: build and run goforge (pass args via ARGS="new foo --module ...")
run: build
	./$(BINARY_PATH) $(ARGS)

## version: print the version that would be embedded in a build right now
version:
	@echo $(VERSION)

## test: run the test suite
test:
	go test ./...

## vet: run go vet across the tree
vet:
	go vet ./...

## clean: remove build artifacts
clean:
	rm -rf $(BINARY_DIR)

VERSION ?= dev


## sync-blueprint: regenerate blueprintsrc/ (embeddable, go.mod-less mirror of blueprint/)
sync-blueprint:
	@bash scripts/sync-blueprint.sh
	
## check-blueprint-sync: fail if blueprintsrc/ has drifted from blueprint/
check-blueprint-sync: sync-blueprint
	@git diff --exit-code blueprintsrc/ || { \
		echo "error: blueprintsrc/ is out of sync with blueprint/."; \
		echo "run 'make sync-blueprint' and commit the result."; \
		exit 1; \
	}

install-hooks:
	cp scripts/pre-commit.sh .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@echo "Hooks installed! Remember to run 'make install-hooks' again if you update the scripts."