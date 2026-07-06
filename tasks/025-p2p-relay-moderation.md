# 025. P2P relay & moderation

**Статус:** done  
**Фаза:** p2p  
**Зависимости:** 024, 015

## Описание

Пересылка сообщений между двумя живыми собеседниками без раскрытия Telegram-контактов. Базовая модерация и report.

## Scope

- Relay: user A message → API → bot → user B (и наоборот)
- Поддержка: text, photo (max 3/dialog), sticker
- Не пересылать: contact, location, forward metadata
- Inline-кнопки «Пожаловаться» / «Заблокировать» при P2P-матче
- Report → event `dialog.reported`, optional notify `REPORT_CHAT_ID`
- Block → end dialog + `blocked_pair` Redis TTL 24h (matcher skip)
- Rate limit: 30 msg/min per user in P2P

## Acceptance criteria

- [x] Текст A → доставлен B без username/chat_id A
- [x] Фото relay работает, file_id re-send
- [x] Report создаёт event для модерации
- [x] Block завершает dialog для обоих
- [x] Контакты и геолокация не проходят через relay
- [x] End dialog (015) корректно закрывает P2P с обеих сторон

## Технические заметки

- API: `POST /dialogs/{id}/relay`, `/report`, `/block`
- Bot: `send_message` / `send_photo` / `send_sticker` (не forward)
- Ban list: `anonimus:blocked_pair:{min}:{max}` TTL 24h
- Admin channel: `REPORT_CHAT_ID` (optional)

## Out of scope

- AI-модерация контента в real-time
- End-to-end encryption
- Voice messages
