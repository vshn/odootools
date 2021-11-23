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

# Note: We only use package.json for Renovate support

generate: ## Generate code and assets
	curl -sSLo templates/bootstrap.min.css https://cdn.jsdelivr.net/npm/bootstrap@$(shell jq --raw-output '.packages."node_modules/bootstrap".version' package-lock.json)/dist/css/bootstrap.min.css
	curl -sSLo templates/bootstrap.min.css.map https://cdn.jsdelivr.net/npm/bootstrap@$(shell jq --raw-output '.packages."node_modules/bootstrap".version' package-lock.json)/dist/css/bootstrap.min.css.map

.PHONY: build
build: build.bin build.docker ## All-in-one build

.PHONY: build.bin
build.bin: export CGO_ENABLED = 0
build.bin: fmt vet templates/bootstrap.min.css ## Build binary
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
lint: fmt vet generate ## All-in-one linting
	@echo 'Check for uncommitted changes ...'
	git diff --exit-code

run: export LISTEN_ADDRESS=localhost:4200
run: export SECRET_KEY=$(LOCAL_SECRET_KEY)
run: ## Run a local instance on localhost:4200
	go run main.go web

run.docker: build.docker ## Run in docker on port 8080
	docker run --rm -it --env "SECRET_KEY=$(LOCAL_SECRET_KEY)" --env ODOO_DB --env ODOO_URL --env "LISTEN_ADDRESS=:8080" --publish "8080:8080" $(CONTAINER_IMG) web

templates/bootstrap.min.css: generate

.helmfile:
	helmfile -e preview -f envs/helmfile.yaml $(helm_cmd)

preview.template: helm_cmd = template
preview.template: export IMG_TAG = latest
preview.template: .helmfile

preview.push: export CONTAINER_IMG = registry.cloudscale-lpg-1.appuio.cloud/odootools-preview/odootools
preview.push: build.docker
	docker push $(CONTAINER_IMG)

preview.deploy: export IMG_TAG = latest
preview.deploy: helm_cmd = apply
preview.deploy: preview.push .helmfile
