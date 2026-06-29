# 001. Project scaffold

**Статус:** todo  
**Фаза:** infra  
**Зависимости:** —

## Описание

Создать базовую структуру монорепозитория для Telegram-бота анонимного общения. Заложить разделение на сервисы, общие конфиги и точки входа для дальнейшей разработки.

## Scope

- Директории: `bot/`, `api/`, `ai/`, `docker/`, `migrations/`, `shared/`
- `.env.example` с переменными: `BOT_TOKEN`, `DATABASE_URL`, `REDIS_URL`, `LLM_API_KEY`, `WEBHOOK_URL`
- `Makefile` с командами: `dev`, `migrate`, `lint`, `test`
- `pyproject.toml` или отдельные `requirements.txt` per service (Python: aiogram + FastAPI)
- `.gitignore` для Python, Docker, `.env`
- Базовый `README.md` в корне со ссылкой на `tasks/`

## Acceptance criteria

- [ ] Структура репозитория создана и задокументирована
- [ ] `.env.example` содержит все обязательные переменные
- [ ] `make dev` (или аналог) описан и готов к подключению docker-compose
- [ ] Каждый сервис имеет минимальный entrypoint (пустой health-check endpoint / bot stub)

## Технические заметки

- **Стек:** Python 3.12+, aiogram 3 (bot), FastAPI (api, ai)
- `bot/` — только Telegram I/O и FSM, бизнес-логика через HTTP к `api/`
- `api/` — Core API: users, dialogs, matchmaking, payments
- `ai/` — LLM-диалоги, классификация intent, подбор фото
- `shared/` — общие модели Pydantic, enum'ы (Gender, Language, NsfwLevel)

## Out of scope

- Реализация бизнес-логики сервисов
- CI/CD pipeline
- Production deployment (Kubernetes, etc.)
