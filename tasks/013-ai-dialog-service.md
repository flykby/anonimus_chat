# 013. AI dialog service

**Статус:** todo  
**Фаза:** ai  
**Зависимости:** 010, 003, 004, 005

## Описание

Отдельный сервис для ведения AI-диалогов: хранение контекста, вызов LLM, генерация ответа персоны. Используется для M+F и F+F сценариев.

## Scope

- Endpoints:
  - `POST /dialog/{dialog_id}/message` — user message → assistant response
  - `GET /dialog/{dialog_id}/context` — debug/admin
- Контекст: последние 20 сообщений в Redis `dialog_ctx:{dialog_id}`
- LLM provider: OpenAI-compatible API (GPT-4o-mini / local)
- System prompt из persona (задача 014)
- Latency budget: p95 < 5 сек
- Сохранение сообщений в `dialog_messages`
- Emit `message.sent`, `message.received`

## Acceptance criteria

- [ ] Пользователь отправляет текст → получает ответ персоны в Telegram < 10 сек
- [ ] Контекст сохраняется между сообщениями в рамках одного dialog
- [ ] Новый dialog — чистый контекст
- [ ] Ошибка LLM → fallback сообщение «попробуй позже», не crash
- [ ] Язык ответа соответствует language профиля пользователя

## Технические заметки

- Структура messages для LLM: `[system, ...history, user]`
- Temperature ~0.8 для живости, max_tokens ~300
- Retry 1 раз при timeout
- Отдельный HTTP client с connection pool
- Не блокировать bot webhook — async processing с typing indicator

## Out of scope

- Streaming ответов в Telegram (можно v2)
- Fine-tuning моделей
- Photo intent (задача 016)
