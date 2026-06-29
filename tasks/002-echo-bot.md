# 002. Echo bot

**Статус:** todo  
**Фаза:** milestone-1  
**Зависимости:** 001

## Описание

Минимально работающий Telegram-бот на **Go**: поднимается, принимает сообщения и отвечает тем же текстом (echo). Первый end-to-end сценарий для проверки деплоя и CI до подключения БД, Redis и бизнес-логики.

## Scope

- Go + [`github.com/go-telegram/bot`](https://github.com/go-telegram/bot): handler на текстовые сообщения
- Long polling для локальной разработки (`make dev-bot`)
- Webhook-режим опционально (полная реализация — задача 009)
- `/start` → приветствие; любой текст → тот же текст обратно
- `GET /health` на `:8080` (отдельный goroutine / `net/http`) для Docker healthcheck
- Structured logging: `slog` — update_id, user_id, message length
- Graceful shutdown по SIGTERM/SIGINT (важно для Docker)

## Acceptance criteria

- [ ] Бот отвечает echo на текстовое сообщение в Telegram
- [ ] `/start` возвращает короткое приветствие
- [ ] Бот стартует из Docker-образа с `BOT_TOKEN` из env
- [ ] При невалидном токене — понятная ошибка при старте, не silent hang
- [ ] `GET /health` → 200; контейнер проходит health check на VM

## Технические заметки

- Только thin bot: без HTTP к api/, без Postgres/Redis
- `BOT_TOKEN`, `LOG_LEVEL` — обязательные env на этом этапе
- Структура:
  - `bot/cmd/bot/main.go`
  - `bot/internal/handlers/echo.go`
  - `bot/internal/config/config.go`
- Образ: multi-stage `docker/bot.Dockerfile` — `golang:1.22` builder → `distroless`/`alpine` runtime со static binary
- Сборка образа — в задаче 003

## Out of scope

- Регистрация, FSM, меню (010–011)
- Webhook + secret token (009)
- Интеграция с Core API
