.PHONY: dev dev-bot dev-api dev-ai lint test build migrate fmt help

PYTHON ?= python3
PIP ?= $(PYTHON) -m pip
BOT_DIR := bot
BIN_DIR := bin

help:
	@echo "Targets: dev-bot dev-api dev-ai lint test build migrate fmt"

# --- Dev ---

dev: dev-bot

dev-bot:
	cd $(BOT_DIR) && go run ./cmd/bot

dev-api:
	$(PYTHON) -m uvicorn api.main:app --reload --host 0.0.0.0 --port 8000

dev-ai:
	$(PYTHON) -m uvicorn ai.main:app --reload --host 0.0.0.0 --port 8001

# --- Quality ---

lint: lint-go lint-py

lint-go:
	cd $(BOT_DIR) && go vet ./...
	cd $(BOT_DIR) && test -z "$$(gofmt -l .)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd $(BOT_DIR) && golangci-lint run ./...; \
	fi

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

build-docker:
	docker build -f docker/bot.Dockerfile -t anonimus/bot:local .
	docker build -f docker/api.Dockerfile -t anonimus/api:local .
	docker build -f docker/ai.Dockerfile -t anonimus/ai:local .

# --- Migrations (006+) ---

migrate:
	@echo "Alembic migrations not configured yet — see tasks/006-database-schema.md"

# --- Setup ---

install-py:
	$(PIP) install -r requirements.txt

tidy:
	cd $(BOT_DIR) && go mod tidy
