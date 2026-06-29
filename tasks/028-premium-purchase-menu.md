# 028. Premium purchase menu

**Статус:** todo  
**Фаза:** profile  
**Зависимости:** 023, 019, 020

## Описание

Меню покупки premium из профиля: описание преимуществ, кнопка оплаты через Stars.

## Scope

- Текст преимуществ premium:
  - Adult-фото без blur
  - (опционально) приоритет в очереди — если добавите позже
- Кнопка «Купить Premium — N ⭐ / 30 дней»
- Если premium active: «Продлить» + дата окончания
- Redirect в invoice flow (019)
- После оплаты → обновлённый профиль
- i18n RU/EN

## Acceptance criteria

- [ ] Кнопка из профиля открывает меню покупки
- [ ] Invoice создаётся с корректной ценой
- [ ] После оплаты профиль показывает новый expires_at
- [ ] Продление добавляет 30 дней к текущему expires_at (не с now)
- [ ] Пользователь без premium видит список benefits

## Технические заметки

- Цена в config: `PREMIUM_PRICE_STARS=200`, `PREMIUM_DURATION_DAYS=30`
- Extend logic: `expires_at = max(now, current_expires) + 30d`
- Не показывать меню во время активного dialog optional

## Out of scope

- Промокоды
- Gift premium
- A/B pricing
