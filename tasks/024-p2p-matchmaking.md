# 024. P2P matchmaking

**Статус:** todo  
**Фаза:** dialog  
**Зависимости:** 013, 007, 014

## Описание

Очередь и матчинг для сценария M+M: два пользователя с профилем (male, seeking male) объединяются в P2P-диалог.

## Scope

- Enqueue: `LPUSH queue:p2p:male {user_id, timestamp}`
- Matcher worker (loop или trigger): atomic pop 2 users → create P2P dialog
- Создание `dialogs` type=p2p, partner_user_id для обоих
- Redis sessions для обоих с partner_id
- Уведомление обоим: «Собеседник найден, напиши первым»
- Обработка disconnect: если один ушёл из очереди — второй остаётся
- Timeout 120 сек (связь с 011)

## Acceptance criteria

- [ ] Два M seeking M одновременно в очереди → match < 5 сек
- [ ] Один в очереди → ждёт до timeout или cancel
- [ ] Третий M → ждёт следующего свободного
- [ ] Активный dialog блокирует повторный enqueue
- [ ] Emit `queue.matched` + `dialog.started` type=p2p для обоих

## Технические заметки

- Lua script для atomic pair pop
- Проверка: оба user всё ещё online (опционально: last_seen < 60 sec)
- Не матчить user с самим собой (sanity)
- Fairness: FIFO по timestamp

## Out of scope

- M+F P2P
- Video/voice calls
- Priority queue
