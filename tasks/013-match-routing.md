# 013. Match routing

**Статус:** done  
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

- [x] M+F всегда попадает в AI-ветку
- [x] M+M попадает в P2P-очередь
- [x] F+F всегда попадает в AI-ветку
- [x] F+M попадает в P2P-очередь
- [x] Повторный `start` при активном диалоге → отказ
- [x] `match_route` пишется в `dialog.started` event

## Технические заметки

- `internal/match/route.go` — `Resolve(gender, seeking)`
- `internal/match/service.go` — `Start(telegram_id)`
- AI route: создаёт `dialogs` type=ai, emit `queue.entered` + `dialog.started`
- P2P route: `matchqueue.Enqueue`, emit `queue.entered`
- Bot: «Начать разговор» → `POST /match/start`
- Persona assignment — задача 017
- **Live F override для M→F** — [037](037-live-f-priority.md)

## Out of scope

- Live F priority / hybrid M→F → [037](037-live-f-priority.md)
- Гео-фильтры, возрастные фильтры при матче
