# 001. Project scaffold

**Статус:** done  
**Фаза:** milestone-1  
**Зависимости:** —

## Описание

Минимальная структура **Go monorepo**: bot, api, ai как отдельные `cmd/*`, общие пакеты в `internal/`.

## Scope

- Директории: `cmd/bot`, `cmd/api`, `cmd/ai`, `internal/`, `migrations/`, `docker/`
- Корневой `go.mod`, `Makefile`, `.env.example`
- **Go 1.22+** для всех сервисов
- Telegram SDK: [`go-telegram/bot`](https://github.com/go-telegram/bot) (bot)
- Миграции: **goose** (`migrations/`, `make migrate-up`)
- `.gitignore`, `README.md`

## Acceptance criteria

- [x] Структура репозитория создана
- [x] `make lint` и `make test` запускаются
- [x] `cmd/bot`, `cmd/api`, `cmd/ai` — entrypoints с `/health`
- [x] `.env.example` документирует переменные (RunPod, registry, DATABASE_URL)
- [x] goose настроен (`migrations/`, `make migrate-up`)

## Технические заметки

- Единый module path: `github.com/flykby/anonimus_chat`
- `internal/shared/` — enum'ы и модели
- `internal/platform/` — env, httputil
- Контракты между сервисами — HTTP/JSON

## Out of scope

- CI pipeline (003)
- Echo-логика (002)
- Реальная схема БД (006)
