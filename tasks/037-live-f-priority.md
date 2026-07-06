# 037. Live F priority for M→F (hybrid match)

**Статус:** todo  
**Фаза:** dialog  
**Зависимости:** 013, 024, 023

## Описание

Гибридный матч для **M seeking F**: по умолчанию AI-диалог (013), но если в hetero P2P-очереди есть **живая девушка** (F seeking M), приоритетно отдать её мужчине в **live P2P** вместо нейросети.

При дефиците live F (много M, мало F) — **premium-пользователи** получают приоритет на живой матч.

## Мотивация

- Улучшить качество опыта M→F без отказа от AI fallback
- Монетизировать дефицит live F через premium priority
- Пример: N мужчин в ожидании, 1 живая F → она уходит **случайному premium** M, остальные → AI или очередь

## Scope

- При `POST /match/start` для **M seeking F**:
  1. Проверить hetero P2P pool: есть ли F (seeking M) в очереди / доступна для матча
  2. **Если есть live F** → создать P2P dialog (не AI), emit `dialog.started` type=p2p, `match_route: m_seeks_f_live`
  3. **Если live F нет** → fallback на AI (текущее поведение 013), `match_route: m_seeks_f`
- **Premium priority** при `waiting_males > live_females`:
  - Выбор male-кандидата для пары с live F — weighted random среди premium, затем среди free (или strict premium-first)
  - Free users ждут или получают AI fallback — зафиксировать в UX (014)
- Счётчики supply/demand для matcher (Redis или in-memory snapshot)
- Events:
  - `queue.matched` с `{ route: "p2p", match_type: "live_f", premium_priority: true }`
  - `dialog.started` metadata: `{ type: "p2p", match_route: "m_seeks_f_live", premium_male: bool }`

## Acceptance criteria

- [ ] M seeking F + live F в очереди → P2P dialog, не AI
- [ ] M seeking F + нет live F → AI dialog (fallback)
- [ ] N males, 1 live F, K premium среди males → live F матчится с одним из premium (random среди premium)
- [ ] После live match оба пользователя уходят из очереди, активный dialog блокирует повторный start
- [ ] Free user не «перебивает» premium при дефиците live F
- [ ] Метрики: `live_f_matches`, `ai_fallback_m_seeks_f`, `premium_priority_wins` в events

## Технические заметки

- Расширяет 013, не заменяет: базовый resolver M→F = AI; 037 — **override layer** после enqueue/check
- Hetero pool: F seeking M (024) + males seeking F, ждущие live match (отдельная очередь или флаг на session)
- Premium check: `is_premium(user_id)` из 023
- Random among premium: `ORDER BY random()` или reservoir на Redis ZSET с score = premium ? 0 : 1 + timestamp
- Защита live F: rate limit matches/hour, report flow (025)
- AI persona для того же user не стартует, если live P2P успешен

### Пример алгоритма (дефицит)

```
live_f = count_available_f_seeking_m()
waiting_m = males_waiting_live_f()  // premium + free

if live_f == 0:
    return ai_fallback()

if waiting_m > live_f:
    candidates = filter_premium(waiting_m) or waiting_m  // premium first
    pick random male from candidates
    pair with oldest available F (FIFO for F)
else:
    pair FIFO male with FIFO F
```

## UX (связь с 014)

- M seeking F видит: «Ищем собеседницу…» (как AI queue)
- Если live match: «Собеседница найдена» (не «нейросеть»)
- Если fallback AI: без изменений vs чистый AI path
- Premium hint optional: «Premium — приоритет на живой чат»

## Out of scope

- M seeking F **только** P2P (AI fallback обязателен)
- Гео/возраст фильтры при live match
- Оплата за один live match (только subscription premium)
- F seeking M меняет маршрут (остаётся P2P в 013)

## Связанные задачи

- **013** — базовый route M→F = AI; 037 добавляет live override
- **024** — hetero P2P queue, pop/pair logic
- **023** — `is_premium()` для priority
- **014** — тексты очереди / «найдена» для hybrid path
