# 018. Photo delivery & blur

**Статус:** todo  
**Фаза:** monetization  
**Зависимости:** 016, 017, 020

## Описание

Пайплайн отправки фото пользователю: подбор по tags, применение blur для adult без premium, inline-кнопка разблокировки.

## Scope

- Подбор фото: `SELECT` по persona_id + tag overlap + NOT IN dialog_photos_sent
- **safe:** отправить оригинал без blur
- **adult + no premium:** blur на сервере (Pillow: GaussianBlur radius 30) → sendPhoto + кнопка «Разблокировать за N ⭐»
- **adult + premium:** оригинал без blur
- Запись в `dialog_photos_sent`
- Emit `photo.sent` с was_blurred, nsfw_level
- Callback `unlock_photo:{dialog_photo_id}` → payment flow (019)

## Acceptance criteria

- [ ] Photo request → фото приходит в течение 15 сек после текстового ответа
- [ ] Safe фото без blur для всех пользователей
- [ ] Adult без premium — видимый blur, кнопка unlock
- [ ] Adult с premium — без blur
- [ ] Одно фото не повторяется в рамках dialog
- [ ] Нет подходящего фото → персона отвечает текстом «позже скину»

## Технические заметки

- Blur on-the-fly: download by file_id → blur → send → discard temp file
- Кэш blurred version в Redis optional (file_id → blurred_file_id)
- Spoiler effect Telegram (`has_spoiler=True`) — дополнительно к blur или вместо
- Tag matching: максимальный overlap, random tie-break

## Out of scope

- Video / GIF
- Watermark с user_id
- Reverse image search protection
