# 023. Premium logic

**Статус:** done  
**Фаза:** monetization  
**Зависимости:** 006, 022

## Описание

Бизнес-логика premium-статуса: проверка активной подписки, срок действия, влияние на blur фото и отображение в профиле.

## Scope

- `GET /users/{id}/premium` → `{ active: bool, expires_at: datetime | null }`
- Функция `is_premium(user_id)` — используется в 018, 023, 028
- При покупке: `expires_at = now + 30 days` (или extend если уже active)
- Expired check: lazy при запросе + daily job emit `premium.expired`
- Отображение в профиле: «отсутствует» / «действует до DD.MM.YYYY HH:mm UTC+0»

## Acceptance criteria

- [ ] Premium user получает adult фото без blur
- [ ] Non-premium после покупки сразу видит эффект (без рестарта)
- [ ] Истёкший premium → снова blur на новых adult фото
- [ ] Профиль корректно показывает статус и дату
- [ ] is_premium false для deleted users

## Технические заметки

- Одна запись per user в `premium_subscriptions`, update expires_at
- UTC+0 в отображении — явно указать timezone в формате
- Premium не влияет на обычный P2P (M+M, F+M без дефицита)
- Premium **приоритет на live F** при M→F — задача [037](037-live-f-priority.md)
- Не откатывать premium при soft-delete (правила в 026)

## Out of scope

- Tiered premium (gold/platinum)
- Trial period
- Referral bonuses
