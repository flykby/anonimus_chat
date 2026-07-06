# 009. Telegram bot webhook

**Статус:** done  
**Фаза:** bot  
**Зависимости:** 002, 004, 007

## Описание

Перевести бота с long polling (002) на webhook для prod на VM. HTTPS через reverse proxy. Бот принимает updates, маршрутизирует в handlers; бизнес-логика позже через HTTP к Core API.

## Scope

- Эволюция echo-бота (002): aiogram 3 Bot + Dispatcher
- `POST /telegram/webhook` — endpoint для Telegram updates
- Регистрация webhook: `setWebhook(url, secret_token)`
- Secret token validation → 403 при mismatch
- Middleware: логирование update_id, user_id, latency
- HTTP-клиент к Core API (`httpx.AsyncClient`) — заготовка
- Обработка: message, callback_query, pre_checkout_query, successful_payment (stub handlers)
- nginx/caddy на VM: TLS termination → bot container

## Acceptance criteria

- [ ] Бот отвечает на `/start` и echo через webhook (не polling)
- [ ] Webhook зарегистрирован на prod HTTPS URL
- [ ] Invalid secret token → 403
- [ ] Ошибки не роняют процесс (graceful error message)
- [ ] Health endpoint независим от webhook

## Технические заметки

- Prod: HTTPS обязателен (Telegram requirement), proxy из 004
- Dev: ngrok → local webhook URL
- `allowed_updates`: message, callback_query, pre_checkout_query, successful_payment
- Bot token только в env
- Разделение: bot = thin client, api = state + business rules

## Out of scope

- FSM регистрации (010)
- Inline payments logic (022)
