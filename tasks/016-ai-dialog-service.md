# 016. AI dialog service

**Статус:** todo  
**Фаза:** ai  
**Зависимости:** 013, 006, 007, 008, 036

## Описание

Отдельный сервис для ведения AI-диалогов: хранение контекста, вызов LLM на **RunPod**, генерация ответа персоны. Используется для M+F и F+F сценариев.

## Scope

- Endpoints:
  - `POST /dialog/{dialog_id}/message` — user message → assistant response
  - `GET /dialog/{dialog_id}/context` — debug/admin
- Контекст: последние 20 сообщений в Redis `dialog_ctx:{dialog_id}`
- LLM через RunPod client (036): OpenAI-compatible chat completions API
- System prompt из persona (017)
- Latency budget: p95 < 5 сек (с учётом сети VM → RunPod)
- Сохранение сообщений в `dialog_messages`
- Emit `message.sent`, `message.received`

## Acceptance criteria

- [ ] Пользователь отправляет текст → получает ответ персоны в Telegram < 10 сек
- [ ] Контекст сохраняется между сообщениями в рамках одного dialog
- [ ] Новый dialog — чистый контекст
- [ ] RunPod недоступен → fallback «попробуй позже», не crash
- [ ] Язык ответа соответствует language профиля пользователя

## Технические заметки

- **LLM на RunPod**, не локально на VM и не OpenAI cloud (если не fallback)
- Структура messages: `[system, ...history, user]`
- Temperature ~0.8, max_tokens ~300
- Retry 1 раз при timeout (036 client)
- Typing indicator в bot, не блокировать webhook
- Cold start RunPod: учесть в UX (019 может объединять intent + reply)

## Out of scope

- Streaming в Telegram (v2)
- Fine-tuning
- Photo intent (019)
- Embedding inference (036, используется в 020)
