# 017. Photo catalog

**Статус:** todo  
**Фаза:** monetization  
**Зависимости:** 003, 014

## Описание

Каталог предгенерированных фотографий для каждой AI-персоны. Фото загружаются в Telegram один раз, в БД хранятся metadata и `telegram_file_id`.

## Scope

- CRUD API (admin): `POST /admin/photos`, `GET /personas/{id}/photos`
- Поля фото: persona_id, tags[], nsfw_level (safe|adult), telegram_file_id, unlock_price_stars
- CLI/скрипт загрузки: local file → `sendPhoto` to admin chat → save file_id
- Tag vocabulary: selfie, full_body, smile, outdoor, lingerie, nude, etc.
- Constraint: не отправлять одно фото дважды в одном dialog
- Минимум 5 safe + 5 adult фото на персону для prod

## Acceptance criteria

- [ ] Фото привязаны к persona_id
- [ ] tags и nsfw_level заполнены для всех фото
- [ ] telegram_file_id валиден (фото отправляется через bot)
- [ ] Скрипт bulk upload работает для папки `assets/personas/{name}/`
- [ ] Дубликаты file_id невозможны (unique constraint)

## Технические заметки

- Хранить только file_id, не бинарники в БД
- `unlock_price_stars` default: 50 для adult, 0 для safe
- Индекс GIN на tags для поиска `tags @> ARRAY['selfie']`
- Offline процесс генерации контента — часть задачи 029, не hot path

## Out of scope

- On-the-fly image generation
- S3 / CDN storage
- User-uploaded photos
