# 015. End dialog flow

**Статус:** todo  
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

- [ ] Кнопка видна только во время активного диалога
- [ ] Confirm/Cancel работают корректно
- [ ] После завершения пользователь в главном меню
- [ ] P2P-партнёр получает уведомление и тоже возвращается в меню
- [ ] Повторное завершение уже закрытого диалога — no-op
- [ ] `ended_at` и `end_reason` записаны в `dialogs`

## Технические заметки

- Inline keyboard для confirm: `end:confirm`, `end:cancel`
- duration_sec = `ended_at - started_at`
- message_count из `dialog_messages` или счётчик в Redis
- При end P2P — end dialog для обоих user_id атомарно

## Out of scope

- AI-initiated end (задача 018)
- Оценка собеседника / feedback после диалога
