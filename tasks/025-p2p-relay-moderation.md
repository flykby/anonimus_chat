# 025. P2P relay & moderation

**Статус:** todo  
**Фаза:** p2p  
**Зависимости:** 024, 015

## Описание

Пересылка сообщений между двумя живыми собеседниками без раскрытия Telegram-контактов. Базовая модерация и report.

## Scope

- Relay: user A message → bot → user B (и наоборот)
- Поддержка: text, photo (с лимитом), sticker optional
- Не пересылать: contact, location, forward metadata
- Prefix опционально: «Собеседник:» (или без prefix для immersion)
- Кнопка «Пожаловаться» → emit event, flag dialog, notify admin
- Кнопка «Заблокировать» → end dialog, ban pair repeat match 24h
- Rate limit: 30 msg/min per user in P2P
- Фото в P2P: max 3 за dialog (anti-spam)

## Acceptance criteria

- [ ] Текст A → доставлен B без username/chat_id A
- [ ] Фото relay работает, file_id re-send
- [ ] Report создаёт запись для модерации
- [ ] Block завершает dialog для обоих
- [ ] Контакты и геолокация не проходят через relay
- [ ] End dialog (015) корректно закрывает P2P с обеих сторон

## Технические заметки

- `copy_message` vs `send_message` — не использовать forward
- Admin channel для reports: `REPORT_CHAT_ID`
- Ban list: `blocked_pairs:{min_id}:{max_id}` TTL 24h
- Логировать hash контента, не plain text (privacy) — optional

## Out of scope

- AI-модерация контента в real-time
- End-to-end encryption
- Voice messages
