# 006. Database schema

**Статус:** done  
**Фаза:** infra  
**Зависимости:** 001, 005

## Описание

Спроектировать и применить схему PostgreSQL для пользователей, профилей, диалогов, персон, фото, премиума и событий. Миграции — **goose** (SQL-файлы в `migrations/`).

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
- **goose** SQL migrations в `migrations/`; extension **pgvector**
- Индексы: telegram_id, dialog user_id + ended_at, events by type + created_at
- Go-модели / repository layer в `internal/api/` (pgx или database/sql)

## Acceptance criteria

- [x] Миграции применяются с нуля: `make migrate-up` (goose)
- [x] Все FK и constraints на месте
- [x] Enum-поля: gender (male|female), seeking, language (ru|en), nsfw_level (safe|adult), dialog type
- [x] Soft-delete users через `deleted_at`
- [x] Seed-скрипт для 1 тестовой персоны и 2–3 фото (опционально)

## Технические заметки

- **goose:** `make migrate-create NAME=...`, файлы `migrations/00002_*.sql`
- `DATABASE_URL` с `sslmode=disable` для dev compose
- Enum'ы в Go: `internal/shared/enums.go` (Gender, Language, NsfwLevel)
- Не хранить полный контекст LLM в Postgres — hot context в Redis

## Out of scope

- ClickHouse / аналитическое хранилище
- Alembic / Python ORM
