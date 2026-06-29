# 001. Project scaffold

**Статус:** todo  
**Фаза:** milestone-1  
**Зависимости:** —

## Описание

Создать минимальную структуру монорепозитория, достаточную для echo-бота и CI. Полные сервисы api/ai — заготовки (stub), наполняются позже.

## Scope

- Директории: `bot/`, `api/`, `ai/`, `docker/`, `migrations/`, `shared/`, `tests/`
- `.env.example`:
  - `BOT_TOKEN`
  - `DATABASE_URL`, `REDIS_URL` (для 005+)
  - `RUNPOD_LLM_URL`, `RUNPOD_LLM_API_KEY` (для 036+)
  - `RUNPOD_EMBEDDING_URL`, `RUNPOD_EMBEDDING_API_KEY`
  - `REGISTRY_URL`, `WEBHOOK_URL`
- `Makefile`: `dev`, `lint`, `test`, `build`, `migrate`
- Python 3.12+, aiogram 3 (bot), FastAPI stubs (api, ai)
- `.gitignore` для Python, Docker, `.env`
- `README.md` в корне со ссылкой на `tasks/`

## Acceptance criteria

- [ ] Структура репозитория создана
- [ ] `make lint` и `make test` запускаются (даже если тестов пока мало)
- [ ] `bot/` имеет entrypoint, готовый к задаче 002
- [ ] `api/` и `ai/` отвечают `GET /health` → 200 (stub)
- [ ] `.env.example` документирует все переменные, включая RunPod и registry

## Технические заметки

- **Деплой:** prod на VM в Docker-контейнерах; БД тоже в контейнере на VM (005, 004)
- **Inference:** LLM и embeddings на RunPod, не на prod VM (036)
- `bot/` — Telegram I/O; бизнес-логика позже через HTTP к `api/`
- `shared/` — Pydantic-модели, enum'ы (Gender, Language, NsfwLevel)

## Out of scope

- CI pipeline (003)
- Echo-логика (002)
- Реализация бизнес-логики api/ai
