# anonimus_chat

Telegram-бот для анонимного общения и обмена фотографиями.

**Стек:** Telegram-бот — **Go**; Core API и AI service — **Python (FastAPI)**.  
**Деплой:** Docker на VM. LLM и embeddings — RunPod.

## Структура репозитория

```
anonimus_chat/
├── bot/           # Go — Telegram I/O
│   └── cmd/bot/   # entrypoint
├── api/           # Python FastAPI — Core API
├── ai/            # Python FastAPI — LLM/embeddings proxy
├── shared/        # Python enums/models (api + ai)
├── docker/        # Dockerfiles
├── migrations/    # Alembic (006+)
├── tests/         # Python tests
└── tasks/         # Backlog
```

## Быстрый старт

```bash
cp .env.example .env   # заполнить BOT_TOKEN позже (002)

# Python deps
make install-py

# Go deps
make tidy

# Lint & test
make lint
make test

# Запуск stubs
make dev-bot    # long polling; нужен BOT_TOKEN в .env
make dev-api    # API  — GET http://localhost:8000/health
make dev-ai     # AI   — GET http://localhost:8001/health

# CI (локально или на VM)
make ci                   # lint → test → build → smoke → push
bash scripts/ci-docker.sh # CI в контейнере через docker.sock
cp .env.ci.example .env.ci  # registry credentials для push
```

## Backlog

Задачи проекта: [tasks/README.md](tasks/README.md)

**Milestone 1:** 001 → 002 → 003 → 004 — scaffold, echo-бот, CI, деплой на VM.
