# 006. Database schema

**Статус:** todo  
**Фаза:** infra  
**Зависимости:** 001, 005

## Описание

Спроектировать и применить схему PostgreSQL для пользователей, профилей, диалогов, персон, фото, премиума и событий. Основа для всей бизнес-логики.

## Scope

- Таблицы:
  - `users` — id, telegram_id (unique), public_uuid, created_at, deleted_at
  - `profiles` — user_id, gender, seeking, age, language
  - `premium_subscriptions` — user_id, expires_at, purchased_at
  - `personas` — id, name, gender, prompt_version, system_prompt, active
  - `photos` — id, persona_id, tags (text[]), nsfw_level, telegram_file_id, unlock_price_stars, **embedding vector** (pgvector)
  - `dialogs` — id, user_id, type (ai|p2p), persona_id, partner_user_id, started_at, ended_at, end_reason
  - `dialog_messages` — dialog_id, role (user|assistant|system), content, created_at
  - `dialog_photos_sent` — dialog_id, photo_id, was_blurred, was_unlocked, sent_at
  - `events` — id, user_id, dialog_id, event_type, metadata (jsonb), created_at
  - `deletion_benefits` — telegram_id, free_unlock_used_at (антиабуз при удалении)
- Alembic migrations; extension **pgvector** для embedding search (020, 036)
- Индексы: telegram_id, dialog user_id + ended_at, events by type + created_at

## Acceptance criteria

- [ ] Миграции применяются с нуля: `alembic upgrade head`
- [ ] Все FK и constraints на месте
- [ ] Enum-поля: gender (male|female), seeking, language (ru|en), nsfw_level (safe|adult), dialog type
- [ ] Soft-delete users через `deleted_at`
- [ ] Seed-скрипт для 1 тестовой персоны и 2–3 фото (опционально)

## Технические заметки

- `public_uuid` — UUID v4, показывается пользователю в профиле
- `telegram_file_id` — ссылка на файл в Telegram, не S3 в v1
- `events.metadata` — гибкое хранение: persona_id, photo_id, reason, etc.
- Не хранить полный контекст LLM в Postgres — только последние N сообщений или summary (Redis для hot context)

## Out of scope

- ClickHouse / аналитическое хранилище
- Шифрование at-rest (managed DB в проде)
