# 013. Match routing

**Статус:** todo  
**Фаза:** dialog  
**Зависимости:** 010, 007, 008

## Описание

Определить маршрут матчинга по комбинации (пол пользователя, кого ищет). Четыре поддерживаемых сценария v1.

## Scope

- Таблица маршрутизации:

| Пол | Ищу | Маршрут |
|-----|-----|---------|
| M | F | AI (нейросеть) |
| M | M | P2P (живой собеседник) |
| F | F | AI (нейросеть) |
| F | M | P2P (живой собеседник) |

- API: `POST /match/start` → `{ route: "ai" | "p2p", dialog_id? }`
- Валидация профиля перед матчем (зарегистрирован, нет активного диалога)
- Emit `queue.entered` с metadata `{ route, gender, seeking }`

## Acceptance criteria

- [ ] M+F всегда попадает в AI-ветку
- [ ] M+M попадает в P2P-очередь
- [ ] F+F всегда попадает в AI-ветку
- [ ] F+M попадает в P2P-очередь
- [ ] Повторный `start` при активном диалоге → отказ
- [ ] `match_route` пишется в `dialog.started` event

## Технические заметки

- Route resolver — чистая функция: `resolve_route(gender, seeking) -> MatchRoute`
- AI route: сразу создать dialog + назначить persona (задача 017, 032)
- P2P route: добавить в Redis queue (задача 024)
  - M+M → `anonimus:queue:p2p:male` (оба M, seeking M)
  - F+M → очередь для hetero P2P (F seeking M; матч с совместимым M — задача 024)
- `match_route` keys: `m_seeks_f` (ai), `m_seeks_m` (p2p), `f_seeks_f` (ai), `f_seeks_m` (p2p)
- **Live F override для M→F** (живая F вместо AI при наличии в очереди, premium priority) — отдельная задача [037](037-live-f-priority.md)

## Out of scope

- Live F priority / hybrid M→F → [037](037-live-f-priority.md)
- Гео-фильтры, возрастные фильтры при матче
