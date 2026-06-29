# 001. Project scaffold

**Статус:** todo  
**Фаза:** milestone-1  
**Зависимости:** —

## Описание

Создать минимальную структуру монорепозитория, достаточную для echo-бота и CI. **Telegram-бот — на Go.** Сервисы api/ai — Python stubs, наполняются позже.

## Scope

- Директории:
  - `bot/` — Go module (`go.mod`), Telegram bot
  - `api/`, `ai/` — Python FastAPI stubs
  - `docker/`, `migrations/`, `shared/`, `tests/`
- `.env.example`:
  - `BOT_TOKEN`
  - `DATABASE_URL`, `REDIS_URL` (для 005+)
  - `RUNPOD_LLM_URL`, `RUNPOD_LLM_API_KEY` (для 036+)
  - `RUNPOD_EMBEDDING_URL`, `RUNPOD_EMBEDDING_API_KEY`
  - `REGISTRY_URL`, `WEBHOOK_URL`
- `Makefile`: `dev`, `lint`, `test`, `build`, `migrate`
- **Go 1.22+** для `bot/` (Telegram SDK: [`go-telegram/bot`](https://github.com/go-telegram/bot))
- **Python 3.12+**, FastAPI stubs для `api/`, `ai/`
- `.gitignore` для Go, Python, Docker, `.env`
- `README.md` в корне со ссылкой на `tasks/`

## Acceptance criteria

- [ ] Структура репозитория создана
- [ ] `make lint` и `make test` запускаются для bot (Go) и stubs api/ai
- [ ] `bot/` имеет `cmd/bot/main.go`, готовый к задаче 002
- [ ] `api/` и `ai/` отвечают `GET /health` → 200 (Python stub)
- [ ] `.env.example` документирует все переменные, включая RunPod и registry

## Технические заметки

- **Деплой:** prod на VM в Docker-контейнерах; БД тоже в контейнере на VM (005, 004)
- **Inference:** LLM и embeddings на RunPod, не на prod VM (036)
- `bot/` — thin client на **Go**; бизнес-логика через HTTP к `api/`
- `shared/` — Pydantic-модели для api/ai; контракты с bot через OpenAPI/JSON (не общий код Go↔Python)

## Out of scope

- CI pipeline (003)
- Echo-логика (002)
- Реализация бизнес-логики api/ai
