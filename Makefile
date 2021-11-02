# Set Shell to bash, otherwise some targets fail with dash/zsh etc.
SHELL := /bin/bash

# Disable built-in rules
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-builtin-variables
.SUFFIXES:
.SECONDARY:
.DEFAULT_GOAL := help

include Makefile.vars.mk

.PHONY: help
help: ## Show this help
	@grep -E -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: build.bin build.docker ## All-in-one build

.PHONY: build.bin
build.bin: export CGO_ENABLED = 0
build.bin: fmt vet ## Build binary
	@go build -o odootools main.go

.PHONY: build.docker
build.docker: build.bin ## Build docker image
	docker build -t $(CONTAINER_IMG) .

.PHONY: test
test:
	@go test -race -coverprofile cover.out -covermode atomic -count 1 ./...

.PHONY: fmt
fmt: ## Run 'go fmt' against code
	go fmt ./...

.PHONY: vet
vet: ## Run 'go vet' against code
	go vet ./...

.PHONY: lint
lint: fmt vet ## All-in-one linting
	@echo 'Check for uncommitted changes ...'
	git diff --exit-code

run: export LISTEN_ADDRESS=localhost:4200
run: export SECRET_KEY=$(LOCAL_SECRET_KEY)
run: ## Run a local instance on localhost:4200
	go run main.go web

run.docker: build.docker ## Run in docker on port 8080
	docker run --rm -it --env "SECRET_KEY=$(LOCAL_SECRET_KEY)" --env ODOO_DB --env ODOO_URL --env "LISTEN_ADDRESS=:8080" --publish "8080:8080" $(CONTAINER_IMG) web
