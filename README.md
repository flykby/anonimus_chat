# anonimus_chat

Telegram-бот для анонимного общения и обмена фотографиями.

**Стек:** все сервисы на **Go 1.22+**. Миграции БД — **goose** (SQL). Inference — **RunPod**.

## Структура репозитория

```
anonimus_chat/
├── cmd/
│   ├── bot/       # Telegram I/O
│   ├── api/       # Core API
│   └── ai/        # LLM/embeddings proxy → RunPod
├── internal/      # shared packages (bot handlers, platform utils)
├── migrations/    # goose SQL migrations
├── docker/        # Dockerfiles
└── tasks/         # Backlog
```

## Быстрый старт

```bash
cp .env.example .env   # BOT_TOKEN, DATABASE_URL, …

make tidy
make lint
make test
make build

# Локально (без Docker)
make dev-bot    # long polling; нужен BOT_TOKEN
make dev-api    # http://localhost:8000/health
make dev-ai     # http://localhost:8001/health

# Docker Compose (postgres + redis + все сервисы)
make compose-up
make compose-down

# Миграции (goose)
make migrate-up
make migrate-status

# CI / deploy
make ci
cp .env.prod.example .env && bash scripts/deploy.sh --tag latest
```

Подробнее: [docs/deploy.md](docs/deploy.md)

## Backlog

[tasks/README.md](tasks/README.md)
