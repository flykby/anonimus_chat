# 006. Telegram bot webhook

**Статус:** todo  
**Фаза:** bot  
**Зависимости:** 001, 002, 004

## Описание

Поднять Telegram-бота на webhook (не long polling в проде). Бот принимает updates, маршрутизирует в handlers и проксирует бизнес-логику в Core API по HTTP.

## Scope

- aiogram 3 Bot + Dispatcher
- `POST /telegram/webhook` — endpoint для Telegram updates (в bot или api)
- Регистрация webhook: `setWebhook(url, secret_token)`
- Secret token validation на входящих запросах
- Middleware: логирование update_id, user_id, latency
- HTTP-клиент к Core API (`httpx.AsyncClient`)
- Обработка: message, callback_query, pre_checkout_query, successful_payment

## Acceptance criteria

- [ ] Бот отвечает на `/start` в Telegram
- [ ] Webhook зарегистрирован и принимает updates
- [ ] Invalid secret token → 403
- [ ] Ошибки API не роняют бота (graceful error message пользователю)
- [ ] Health endpoint бота независим от webhook

## Технические заметки

- Dev: ngrok → local webhook URL
- Prod: HTTPS обязателен (Telegram requirement)
- `allowed_updates`: message, callback_query, pre_checkout_query, successful_payment
- Bot token только в env, никогда в коде
- Разделение: bot = thin client, api = state + business rules

## Out of scope

- FSM регистрации (задача 007)
- Inline payments logic (задача 019)
