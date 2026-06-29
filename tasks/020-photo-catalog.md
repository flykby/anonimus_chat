# 020. Photo catalog

**Статус:** todo  
**Фаза:** monetization  
**Зависимости:** 006, 017, 036

## Описание

Каталог предгенерированных фотографий для каждой AI-персоны. Поиск по тегам + **семантический поиск через embedding-модель на RunPod** (036).

## Scope

- CRUD API (admin): `POST /admin/photos`, `GET /personas/{id}/photos`
- Поля фото: persona_id, tags[], nsfw_level, telegram_file_id, unlock_price_stars, **embedding vector**
- При upload: вычислить embedding через RunPod → сохранить в pgvector или JSONB
- Поиск: `tags @> ...` + cosine similarity по embedding запроса пользователя
- CLI/скрипт загрузки: local file → admin chat file_id → embedding → БД
- Constraint: не отправлять одно фото дважды в одном dialog
- Минимум 5 safe + 5 adult фото на персону для prod

## Acceptance criteria

- [ ] Фото привязаны к persona_id, tags и nsfw_level заполнены
- [ ] Embedding сохранён для каждого фото
- [ ] Семантический запрос («селфи у окна») находит релевантное фото
- [ ] telegram_file_id валиден
- [ ] Bulk upload script работает для `assets/personas/{name}/`

## Технические заметки

- **Embeddings на RunPod** — отдельный pod от chat LLM (036)
- pgvector extension в Postgres (006) или fallback: in-memory для MVP
- Хранить только file_id, не бинарники в БД
- `unlock_price_stars` default: 50 adult, 0 safe
- Offline генерация контента — 032

## Out of scope

- On-the-fly image generation
- S3 / CDN
- User-uploaded photos
