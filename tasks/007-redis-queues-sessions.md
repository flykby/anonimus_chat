# 007. Redis queues & sessions

**Статус:** done  
**Фаза:** infra  
**Зависимости:** 001, 005

## Описание

Настроить Redis для очереди матчинга P2P, хранения активных сессий диалогов, FSM-состояний бота и rate limiting.

## Scope

- **Match queue (P2P):** sorted set `anonimus:queue:p2p:{gender}` — user_id + timestamp
- **Active sessions:** hash `anonimus:session:{user_id}` → dialog_id, type, partner_id, persona_id, started_at
- **FSM state:** `anonimus:fsm:{telegram_id}` — состояние бота (TTL 24h)
- **Rate limits:** `anonimus:ratelimit:{user_id}:{action}` — счётчик с TTL
- **Dialog context (AI):** list `anonimus:dialog_ctx:{dialog_id}` — последние N сообщений для LLM
- **Go-клиент:** `github.com/redis/go-redis/v9` — packages в `internal/redis/`

## Acceptance criteria

- [x] Пользователь может встать в P2P-очередь и быть извлечён парой
- [x] Активная сессия читается/записывается за < 5ms (local)
- [x] При завершении диалога session key удаляется
- [x] Rate limit блокирует спам (> N сообщений/мин)
- [x] FSM state переживает рестарт бота (персист в Redis, не in-memory)

## Технические заметки

- Packages: `matchqueue`, `session`, `fsm`, `ratelimit`, `dialogctx`, `keys`
- Atomic pop двух пользователей: Lua script в `matchqueue.TryMatchPair`
- TTL на `dialog_ctx` — 4h после последней активности
- При матче P2P — `session.SetP2PPair` для обоих user_id
- API `/health` проверяет `redis_ok`

## Out of scope

- Redis Cluster / Sentinel (single instance для dev и early prod)
- Pub/Sub для real-time (polling/webhook достаточно для Telegram)
