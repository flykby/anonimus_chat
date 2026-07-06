# 024. P2P matchmaking

**Статус:** todo  
**Фаза:** dialog  
**Зависимости:** 013, 007, 014

## Описание

Очередь и матчинг для P2P-сценариев **M+M** и **F+M** (живой собеседник).

## Scope

- **M+M:** enqueue `anonimus:queue:p2p:male` — два пользователя (male, seeking male) → P2P-диалог
- **F+M:** enqueue в hetero P2P-очередь — пользователь (female, seeking male) матчится с совместимым male-партнёром (логика пары — в matcher)
- Matcher worker (loop или trigger): atomic pop compatible pair → create P2P dialog
- Создание `dialogs` type=p2p, partner_user_id для обоих
- Redis sessions для обоих с partner_id
- Уведомление обоим: «Собеседник найден, напиши первым»
- Обработка disconnect: если один ушёл из очереди — второй остаётся
- Timeout 120 сек (связь с 014)

## Acceptance criteria

- [ ] Два M seeking M одновременно в очереди → match < 5 сек
- [ ] F seeking M находит совместимого male-партнёра (hetero P2P)
- [ ] Один в очереди → ждёт до timeout или cancel
- [ ] Активный dialog блокирует повторный enqueue
- [ ] Emit `queue.matched` + `dialog.started` type=p2p для обоих

## Технические заметки

- Lua script для atomic pair pop (уже есть `TryMatchPair` для same-gender queue)
- Hetero P2P (F+M): отдельная очередь или cross-match — уточнить при реализации
- Проверка: оба user всё ещё online (опционально: last_seen < 60 sec)
- Не матчить user с самим собой (sanity)
- Fairness: FIFO по timestamp

## Out of scope

- M+F P2P (M seeking F остаётся AI-only)
- Video/voice calls
- Priority queue
