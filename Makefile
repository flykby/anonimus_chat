.PHONY: dev dev-bot dev-api dev-ai lint lint-go lint-py test build build-bot build-docker \
	push push-bot push-api push-ai migrate fmt help ci install-py tidy

PYTHON ?= python3
PIP ?= $(PYTHON) -m pip
BOT_DIR := bot
BIN_DIR := bin

GIT_SHA ?= $(shell git rev-parse HEAD)
GIT_SHA_SHORT ?= $(shell git rev-parse --short HEAD)
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

REGISTRY_URL ?=
REGISTRY_HOST ?= $(firstword $(subst /, ,$(REGISTRY_URL)))
BOT_IMAGE ?= anonimus/bot:local

help:
	@echo "Targets: dev-bot dev-api dev-ai lint test build build-docker push ci migrate fmt"

# --- Dev ---

dev: dev-bot

dev-bot:
	cd $(BOT_DIR) && go run ./cmd/bot

dev-api:
	$(PYTHON) -m uvicorn api.main:app --reload --host 0.0.0.0 --port 8000

dev-ai:
	$(PYTHON) -m uvicorn ai.main:app --reload --host 0.0.0.0 --port 8001

# --- CI ---

ci:
	bash scripts/ci.sh all

# --- Quality ---

lint: lint-go lint-py

lint-go:
	cd $(BOT_DIR) && go vet ./...
	cd $(BOT_DIR) && test -z "$$(gofmt -l .)"
ifeq ($(CI),true)
	cd $(BOT_DIR) && golangci-lint run ./...
else
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd $(BOT_DIR) && golangci-lint run ./...; \
	fi
endif

lint-py:
	$(PYTHON) -m ruff check api ai shared tests
	$(PYTHON) -m ruff format --check api ai shared tests

fmt:
	cd $(BOT_DIR) && gofmt -w .
	$(PYTHON) -m ruff format api ai shared tests
	$(PYTHON) -m ruff check --fix api ai shared tests

test: test-go test-py

test-go:
	cd $(BOT_DIR) && go test ./...

test-py:
	$(PYTHON) -m pytest tests/ -q

# --- Build ---

build: build-bot

build-bot:
	mkdir -p $(BIN_DIR)
	cd $(BOT_DIR) && CGO_ENABLED=0 go build -o ../$(BIN_DIR)/bot ./cmd/bot

build-docker: build-docker-bot build-docker-api build-docker-ai

build-docker-bot:
	docker build -f docker/bot.Dockerfile \
		-t anonimus/bot:local \
		-t anonimus/bot:$(GIT_SHA_SHORT) \
		.

build-docker-api:
	docker build -f docker/api.Dockerfile \
		-t anonimus/api:local \
		-t anonimus/api:$(GIT_SHA_SHORT) \
		.

build-docker-ai:
	docker build -f docker/ai.Dockerfile \
		-t anonimus/ai:local \
		-t anonimus/ai:$(GIT_SHA_SHORT) \
		.

# --- Registry push (003+) ---

push: push-bot push-api push-ai

push-bot: build-docker-bot
	@test -n "$(REGISTRY_URL)" || (echo "REGISTRY_URL is required for push" && exit 1)
	@test -n "$(REGISTRY_HOST)" || (echo "could not parse registry host from REGISTRY_URL" && exit 1)
	@if [ -n "$(REGISTRY_PASSWORD)" ]; then \
		echo "$$REGISTRY_PASSWORD" | docker login "$(REGISTRY_HOST)" -u "$(REGISTRY_USER)" --password-stdin; \
	fi
	docker tag anonimus/bot:local $(REGISTRY_URL)/bot:$(GIT_SHA)
	docker tag anonimus/bot:local $(REGISTRY_URL)/bot:$(GIT_SHA_SHORT)
	@if [ "$(GIT_BRANCH)" = "main" ]; then \
		docker tag anonimus/bot:local $(REGISTRY_URL)/bot:latest; \
	fi
	docker push $(REGISTRY_URL)/bot:$(GIT_SHA)
	docker push $(REGISTRY_URL)/bot:$(GIT_SHA_SHORT)
	@if [ "$(GIT_BRANCH)" = "main" ]; then docker push $(REGISTRY_URL)/bot:latest; fi

push-api: build-docker-api
	@test -n "$(REGISTRY_URL)" || (echo "REGISTRY_URL is required for push" && exit 1)
	docker tag anonimus/api:local $(REGISTRY_URL)/api:$(GIT_SHA_SHORT)
	docker push $(REGISTRY_URL)/api:$(GIT_SHA_SHORT)

push-ai: build-docker-ai
	@test -n "$(REGISTRY_URL)" || (echo "REGISTRY_URL is required for push" && exit 1)
	docker tag anonimus/ai:local $(REGISTRY_URL)/ai:$(GIT_SHA_SHORT)
	docker push $(REGISTRY_URL)/ai:$(GIT_SHA_SHORT)

# --- Migrations (006+) ---

migrate:
	@echo "Alembic migrations not configured yet — see tasks/006-database-schema.md"

# --- Setup ---

install-py:
	$(PIP) install -r requirements.txt

tidy:
	cd $(BOT_DIR) && go mod tidy
