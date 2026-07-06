# 024. P2P matchmaking

**Статус:** done  
**Фаза:** dialog  
**Зависимости:** 013, 007, 014

## Описание

Очередь и матчинг для P2P-сценариев **M+M** и **F+M** (живой собеседник).

## Scope

- **M+M:** enqueue `anonimus:queue:p2p:male` — два пользователя (male, seeking male) → P2P-диалог
- **F+M:** enqueue `anonimus:queue:hetero:female`; male-партнёр из `anonimus:queue:hetero:male` (037 наполняет male-очередь)
- Matcher на `POST /match/poll` и при `POST /match/start`: atomic pop compatible pair → create P2P dialog
- Создание `dialogs` type=p2p, partner_user_id для обоих (отдельный dialog row на каждого)
- Redis sessions для обоих с partner_id и своим dialog_id
- Bot polling: «Собеседник найден» + dialog keyboard
- Cancel снимает с правильной очереди (p2p vs hetero)
- Timeout 120 сек (связь с 014)

## Acceptance criteria

- [x] Два M seeking M одновременно в очереди → match < 5 сек
- [x] F seeking M находит совместимого male-партнёра (hetero P2P)
- [x] Один в очереди → ждёт до timeout или cancel
- [x] Активный dialog блокирует повторный enqueue
- [x] Emit `queue.matched` + `dialog.started` type=p2p для обоих

## Технические заметки

- Same-gender: Lua `TryMatchPair` на `anonimus:queue:p2p:male`
- Hetero F+M: Lua `TryMatchHeteroPair` — FIFO female + FIFO male
- API: `POST /match/poll` — run matchers, return `matched` / `queued`
- `session.SetP2PPair(userA, userB, dialogAID, dialogBID, startedAt)`
- **M+F live override:** male в hetero pool — задача [037](037-live-f-priority.md)
- Не матчить user с самим собой; при active dialog — re-enqueue оставшихся

## Out of scope

- M+F P2P (M seeking F остаётся AI-only до 037)
- Video/voice calls
- Priority queue
