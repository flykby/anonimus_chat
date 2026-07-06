# 026. Profile view

**Статус:** done  
**Фаза:** profile  
**Зависимости:** 011, 009

## Описание

Экран профиля пользователя: public UUID, статус premium, данные анкеты. Точка входа для редактирования, покупки premium и удаления.

## Scope

- Текст профиля (RU/EN) с UUID, premium, анкетой
- Inline-кнопки: premium / edit / language / delete / back
- API: `GET /users/by-telegram/{id}/profile`, `GET /users/me/profile?telegram_id=`
- Premium статус из `premium_subscriptions` (активная подписка)

## Acceptance criteria

- [x] Профиль отображает актуальные данные из БД
- [x] UUID — public_uuid, не internal id
- [x] Premium статус корректен (active + expires_at UTC+0)
- [x] Все 5 кнопок ведут в правильные flows (stubs → 027–029, 031)
- [x] «Назад» возвращает в главное меню

## Технические заметки

- Формат даты premium: `DD.MM.YYYY HH:MM UTC+0`
- Gender EN: Guy / Girl
- telegram_id не показывается пользователю
- Premium active → кнопка «Продлить премиум»

## Out of scope

- Аватар / фото профиля
- Статистика диалогов в профиле
- Реальная покупка premium (031)
