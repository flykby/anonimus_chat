# 002. Echo bot

**Статус:** done  
**Фаза:** milestone-1  
**Зависимости:** 001

## Описание

Минимально работающий Telegram-бот: поднимается, принимает сообщения и отвечает тем же текстом (echo). Первый end-to-end сценарий для проверки деплоя и CI до подключения БД, Redis и бизнес-логики.

## Scope

- Go + [`github.com/go-telegram/bot`](https://github.com/go-telegram/bot): `/start` и default handler для echo
- Long polling для локальной разработки (`make dev`)
- Webhook-режим опционально (полная реализация — задача 009)
- `/start` → приветствие; любой текст → тот же текст обратно
- Health endpoint или простой `GET /health` (если bot поднимается через FastAPI sidecar)
- Логирование: update_id, user_id, message length
- Graceful shutdown по SIGTERM (важно для Docker)

## Acceptance criteria

- [x] Бот отвечает echo на текстовое сообщение в Telegram
- [x] `/start` возвращает короткое приветствие
- [x] Бот стартует из Docker-образа с `BOT_TOKEN` из env
- [x] При невалидном токене — понятная ошибка при старте, не silent hang
- [x] Контейнер проходит health check / restart policy на VM

## Технические заметки

- Только thin bot: без HTTP к api/, без Postgres/Redis
- `BOT_TOKEN` и опционально `LOG_LEVEL` — единственные обязательные env на этом этапе
- Структура: `bot/cmd/bot/main.go`, `bot/internal/handlers/echo.go`, `bot/internal/config/`
- Образ: multi-stage Dockerfile в `docker/bot.Dockerfile` (собирается в задаче 003)

## Out of scope

- Регистрация, FSM, меню (010–011)
- Webhook + secret token (009)
- Интеграция с Core API
