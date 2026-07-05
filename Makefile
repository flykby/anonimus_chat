.PHONY: dev dev-bot dev-api dev-ai lint test build build-bot build-docker \
	push push-bot push-api push-ai deploy deploy-rollback deploy-check \
	compose-up compose-down compose-ps compose-logs compose-config \
	migrate-up migrate-down migrate-status migrate-create seed tidy fmt help ci

GO ?= go
BIN_DIR := bin
GOOSE ?= goose

GIT_SHA ?= $(shell git rev-parse HEAD)
GIT_SHA_SHORT ?= $(shell git rev-parse --short HEAD)
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

REGISTRY_URL ?=
REGISTRY_HOST ?= $(firstword $(subst /, ,$(REGISTRY_URL)))
COMPOSE ?= docker compose
COMPOSE_FILES := -f docker-compose.yml
COMPOSE_PROD_FILES := -f docker-compose.yml -f docker-compose.prod.yml

DATABASE_URL ?= postgresql://anonimus:anonimus@localhost:5432/anonimus?sslmode=disable

help:
	@echo "Targets: compose-up dev-bot dev-api dev-ai lint test build-docker push ci deploy migrate-up"

# --- Docker Compose ---

compose-up:
	$(COMPOSE) $(COMPOSE_FILES) up -d --build

compose-up-infra:
	$(COMPOSE) $(COMPOSE_FILES) up -d postgres redis

compose-down:
	$(COMPOSE) $(COMPOSE_FILES) down

compose-ps:
	$(COMPOSE) $(COMPOSE_FILES) ps

compose-logs:
	$(COMPOSE) $(COMPOSE_FILES) logs -f

compose-config:
	$(COMPOSE) $(COMPOSE_PROD_FILES) config

# --- Dev ---

dev: dev-bot

dev-bot:
	$(GO) run ./cmd/bot

dev-api:
	$(GO) run ./cmd/api

dev-ai:
	$(GO) run ./cmd/ai

# --- Deploy ---

deploy:
	bash scripts/deploy.sh

deploy-rollback:
	bash scripts/deploy.sh --rollback

deploy-check:
	bash -n scripts/deploy.sh
	bash -n scripts/remote-deploy.sh
	bash -n scripts/setup-vm-ghcr.sh
	bash -n scripts/migrate-prod.sh
	bash scripts/deploy.sh --help >/dev/null

# --- CI ---

ci:
	bash scripts/ci.sh all

# --- Quality ---

lint:
	$(GO) vet ./...
	test -z "$$(gofmt -l .)"
ifeq ($(CI),true)
	golangci-lint run ./...
endif

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

# --- Build ---

build: build-bot build-api build-ai

build-bot:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -o $(BIN_DIR)/bot ./cmd/bot

build-api:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -o $(BIN_DIR)/api ./cmd/api

build-ai:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -o $(BIN_DIR)/ai ./cmd/ai

build-docker: build-docker-bot build-docker-api build-docker-ai

build-docker-bot:
	docker build -f docker/bot.Dockerfile -t anonimus/bot:local -t anonimus/bot:$(GIT_SHA_SHORT) .

build-docker-api:
	docker build -f docker/api.Dockerfile -t anonimus/api:local -t anonimus/api:$(GIT_SHA_SHORT) .

build-docker-ai:
	docker build -f docker/ai.Dockerfile -t anonimus/ai:local -t anonimus/ai:$(GIT_SHA_SHORT) .

# --- Registry push ---

push: push-bot push-api push-ai

push-bot: build-docker-bot
	@test -n "$(REGISTRY_URL)" || (echo "REGISTRY_URL is required for push" && exit 1)
	@if [ -n "$(REGISTRY_PASSWORD)" ]; then \
		echo "$$REGISTRY_PASSWORD" | docker login "$(REGISTRY_HOST)" -u "$(REGISTRY_USER)" --password-stdin; \
	fi
	docker tag anonimus/bot:local $(REGISTRY_URL)/bot:$(GIT_SHA)
	docker tag anonimus/bot:local $(REGISTRY_URL)/bot:$(GIT_SHA_SHORT)
	@if [ "$(GIT_BRANCH)" = "main" ]; then docker tag anonimus/bot:local $(REGISTRY_URL)/bot:latest; fi
	docker push $(REGISTRY_URL)/bot:$(GIT_SHA)
	docker push $(REGISTRY_URL)/bot:$(GIT_SHA_SHORT)
	@if [ "$(GIT_BRANCH)" = "main" ]; then docker push $(REGISTRY_URL)/bot:latest; fi

push-api: build-docker-api
	@test -n "$(REGISTRY_URL)" || (echo "REGISTRY_URL is required for push" && exit 1)
	docker tag anonimus/api:local $(REGISTRY_URL)/api:$(GIT_SHA)
	docker tag anonimus/api:local $(REGISTRY_URL)/api:$(GIT_SHA_SHORT)
	@if [ "$(GIT_BRANCH)" = "main" ]; then docker tag anonimus/api:local $(REGISTRY_URL)/api:latest; fi
	docker push $(REGISTRY_URL)/api:$(GIT_SHA)
	docker push $(REGISTRY_URL)/api:$(GIT_SHA_SHORT)
	@if [ "$(GIT_BRANCH)" = "main" ]; then docker push $(REGISTRY_URL)/api:latest; fi

push-ai: build-docker-ai
	@test -n "$(REGISTRY_URL)" || (echo "REGISTRY_URL is required for push" && exit 1)
	docker tag anonimus/ai:local $(REGISTRY_URL)/ai:$(GIT_SHA)
	docker tag anonimus/ai:local $(REGISTRY_URL)/ai:$(GIT_SHA_SHORT)
	@if [ "$(GIT_BRANCH)" = "main" ]; then docker tag anonimus/ai:local $(REGISTRY_URL)/ai:latest; fi
	docker push $(REGISTRY_URL)/ai:$(GIT_SHA)
	docker push $(REGISTRY_URL)/ai:$(GIT_SHA_SHORT)
	@if [ "$(GIT_BRANCH)" = "main" ]; then docker push $(REGISTRY_URL)/ai:latest; fi

# --- Migrations (goose) ---

migrate-up:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required" && exit 1)
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required" && exit 1)
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required" && exit 1)
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" status

migrate-create:
	@test -n "$(NAME)" || (echo "NAME is required, e.g. make migrate-create NAME=add_users" && exit 1)
	$(GOOSE) -dir migrations create $(NAME) sql

migrate: migrate-up

migrate-prod:
	bash scripts/migrate-prod.sh

migrate-prod-status:
	bash scripts/migrate-prod.sh status

seed:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required" && exit 1)
	psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f scripts/seed.sql

tidy:
	$(GO) mod tidy
