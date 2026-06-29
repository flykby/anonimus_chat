# 026. Delete profile anti-abuse

**Статус:** todo  
**Фаза:** profile  
**Зависимости:** 023, 018, 005

## Описание

Удаление профиля с двойным подтверждением и антиабуз-механикой: одноразовое предложение бесплатно разблокировать adult-фото из диалога.

## Scope

- Flow:
  1. «Удалить профиль» → «Точно удалить?» Confirm/Cancel
  2. Если Confirm и есть locked adult photos в последних N диалогах И `free_unlock_used` false для telegram_id:
     → «Разблокируй одну фото бесплатно» + gallery picker / random one
  3. Финальное подтверждение → soft-delete user (`deleted_at`), очистка sessions
- Таблица `deletion_benefits`: telegram_id, free_unlock_used_at
- Флаг привязан к **telegram_id**, не к user row — пересоздание профиля не сбрасывает
- Активный dialog → force end перед удалением
- Emit `user.deleted`

## Acceptance criteria

- [ ] Cancel на любом шаге → профиль сохранён
- [ ] Бесплатный unlock предлагается максимум 1 раз per telegram_id ever
- [ ] После удаления /start → новая регистрация
- [ ] Premium не восстанавливается на новом профиле
- [ ] P2P partner уведомлён если был активный dialog
- [ ] Данные: soft-delete + purge schedule 30 дней (document in rules)

## Технические заметки

- N диалогов для unlock offer: последние 5
- Gallery: inline buttons с preview blurred или список «Фото 1, 2, 3»
- После unlock → отправить original → продолжить delete flow
- Optional: 7-day cooldown на повторную регистрацию (config flag)

## Out of scope

- GDPR export before delete
- Hard delete immediate
- Refund premium on delete
