# 015. End dialog flow

**Статус:** done  
**Фаза:** dialog  
**Зависимости:** 011, 008, 007

## Описание

Позволить пользователю завершить диалог в любой момент: кнопка «Завершить диалог» → подтверждение Confirm/Cancel → закрытие сессии.

## Scope

- Persistent кнопка «Завершить диалог» во время активного чата (reply keyboard)
- Callback/dialog:
  - «Завершить» → закрыть dialog, очистить Redis session, вернуть главное меню
  - «Отменить» → остаться в диалоге
- API: `POST /dialogs/{id}/end` body `{ reason: "user_confirmed" }`
- Emit `dialog.ended` с duration_sec, message_count
- Для P2P: уведомить партнёра «Собеседник завершил диалог»
- Для AI: остановить контекст в Redis

## Acceptance criteria

- [x] Кнопка видна только во время активного диалога
- [x] Confirm/Cancel работают корректно
- [x] После завершения пользователь в главном меню
- [x] P2P-партнёр получает уведомление и тоже возвращается в меню
- [x] Повторное завершение уже закрытого диалога — no-op
- [x] `ended_at` и `end_reason` записаны в `dialogs`

## Технические заметки

- Inline keyboard: `end:confirm`, `end:cancel`
- `internal/dialog/service.go` — End(), P2P partner end in one tx
- Profile API: `active_dialog_id` для bot
- Redis session set on AI match complete; cleared on end
- AI-initiated end — задача 018

## Out of scope

- AI-initiated end (задача 018)
- Оценка собеседника / feedback после диалога
