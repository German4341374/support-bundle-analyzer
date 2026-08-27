SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

.PHONY: help setup format lint typecheck test test-go test-node test-python test-dotnet test-java test-php build demo up down clean verify docker-build helm-lint terraform-validate

help: ## Show available commands
	@awk 'BEGIN {FS = ":.*## "; printf "Usage: make <target>\n\n"} /^[a-zA-Z_-]+:.*## / {printf "  %-22s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Install Go, Node, and Python development dependencies
	go mod download
	npm ci
	python3 -m pip install -e "./analyzers/log-intelligence-python[dev]"

format: ## Format source files
	gofmt -w apps internal
	npm run format
	python3 -m ruff format analyzers/log-intelligence-python

lint: ## Run available linters and static checks
	test -z "$$(gofmt -l apps internal)"
	go vet ./apps/... ./internal/...
	staticcheck ./apps/... ./internal/...
	npm run lint
	python3 -m ruff check analyzers/log-intelligence-python

typecheck: ## Run language type checkers
	npm run typecheck
	python3 -m mypy analyzers/log-intelligence-python/src

test: test-go test-node test-python ## Run the portable core test suite

test-go:
	go test -race ./apps/... ./internal/...

test-node:
	npm test

test-python:
	python3 -m pytest analyzers/log-intelligence-python

test-dotnet: ## Run Windows analyzer tests when .NET 8 is installed
	dotnet test analyzers/windows-diagnostics-analyzer/tests/WindowsDiagnosticsAnalyzer.Tests.csproj --configuration Release

test-java: ## Run JVM analyzer tests when Maven and Java 17+ are installed
	mvn -B -f analyzers/jvm-diagnostics-analyzer/pom.xml verify

test-php: ## Run PHP analyzer checks when PHP and Composer dependencies are installed
	cd analyzers/php-web-diagnostics-analyzer && composer check

build: ## Build Go CLI and Node API
	mkdir -p bin
	go build -trimpath -o bin/support-bundle-analyzer ./apps/cli
	npm run build

demo: build ## Generate, compare, and sanitize deterministic healthy/outage demos
	scripts/demo.sh

up: ## Start the local Compose stack
	docker compose up --build --wait

down: ## Stop the local Compose stack
	docker compose down

docker-build: ## Build the production container image
	docker build --pull -t support-bundle-analyzer:local .

helm-lint: ## Validate the Helm chart
	helm lint deploy/helm/support-bundle-analyzer

terraform-validate: ## Format and statically validate Terraform
	terraform -chdir=deploy/terraform fmt -check -recursive
	terraform -chdir=deploy/terraform init -backend=false
	terraform -chdir=deploy/terraform validate

verify: ## Run the portable release-gate checks
	scripts/verify.sh

clean: ## Remove generated local artifacts and containers
	docker compose down --volumes --remove-orphans 2>/dev/null || true
	rm -rf .tmp bin coverage apps/api/dist analysis-workspace sanitized-workspace
