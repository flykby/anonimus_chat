# 008. Event logging

**Статус:** todo  
**Фаза:** infra  
**Зависимости:** 006

## Описание

Единый слой логирования продуктовых событий в таблицу `events`. Основа для метрик удержания, атрибуции ухода и A/B персон.

## Scope

- Event types:
  - `user.registered`, `user.profile_updated`, `user.deleted`
  - `dialog.started`, `dialog.ended`
  - `message.sent` (user), `message.received` (ai/partner)
  - `photo.requested`, `photo.sent`, `photo.unlocked`
  - `premium.purchased`, `premium.expired`
  - `queue.entered`, `queue.matched`
- Функция `emit_event(user_id, event_type, dialog_id=None, metadata={})`
- Async запись в Postgres (не блокировать ответ пользователю)
- Structured logging (JSON) в stdout для dev

## Acceptance criteria

- [ ] Все перечисленные event types определены как enum
- [ ] `emit_event` вызывается из API layer, не из bot напрямую
- [ ] metadata валидируется по схеме per event_type (Pydantic)
- [ ] События пишутся с `created_at` UTC
- [ ] Тест: цепочка register → dialog.started → message.sent → dialog.ended создаёт 4 записи

## Технические заметки

Примеры metadata:
- `dialog.started`: `{ "type": "ai", "persona_id": "...", "match_route": "m_seeks_f" }`
- `dialog.ended`: `{ "reason": "user_confirmed", "duration_sec": 342, "message_count": 28 }`
- `photo.sent`: `{ "photo_id": "...", "nsfw_level": "adult", "was_blurred": true }`

- Batch insert опционально для high load
- Индекс `(event_type, created_at)` для аналитики

## Out of scope

- Grafana дашборды (задача 033)
- ClickHouse export
- Real-time streaming
