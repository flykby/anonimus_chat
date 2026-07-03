# 007. Redis queues & sessions

**Статус:** todo  
**Фаза:** infra  
**Зависимости:** 001, 005

## Описание

Настроить Redis для очереди матчинга P2P, хранения активных сессий диалогов, FSM-состояний бота и rate limiting.

## Scope

- **Match queue (P2P):** sorted set или list `queue:p2p:male` — telegram_id / user_id + timestamp
- **Active sessions:** hash `session:{user_id}` → dialog_id, type, partner_id, persona_id, started_at
- **FSM state:** `fsm:{telegram_id}` — текущее состояние бота (TTL 24h)
- **Rate limits:** `ratelimit:{user_id}:{action}` — счётчик с TTL (сообщения, поиск, фото)
- **Dialog context (AI):** list `dialog_ctx:{dialog_id}` — последние N сообщений для LLM
- **Go-клиент:** `github.com/redis/go-redis/v9` (bot FSM, api matchmaking)

## Acceptance criteria

- [ ] Пользователь может встать в P2P-очередь и быть извлечён парой
- [ ] Активная сессия читается/записывается за < 5ms (local)
- [ ] При завершении диалога session key удаляется
- [ ] Rate limit блокирует спам (> N сообщений/мин)
- [ ] FSM state переживает рестарт бота (персист в Redis, не in-memory)

## Технические заметки

- Ключи с namespace prefix: `anonimus:`
- TTL на `dialog_ctx` — 4h после последней активности
- Atomic pop двух пользователей: Lua script или `MULTI/EXEC`
- При матче P2P — оба user_id получают session с partner_id друг друга

## Out of scope

- Redis Cluster / Sentinel (single instance для dev и early prod)
- Pub/Sub для real-time (polling/webhook достаточно для Telegram)
