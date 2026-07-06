# 022. Telegram Stars payments

**Статус:** done  
**Фаза:** monetization  
**Зависимости:** 009, 021

## Описание

Интеграция платежей через Telegram Stars (XTR): покупка premium-подписки и разблокировка отдельного adult-фото.

## Scope

- `sendInvoice` для premium (subscription 30 дней)
- `sendInvoice` для unlock single photo (из callback 018)
- Handlers: `pre_checkout_query` → approve, `successful_payment` → fulfill
- Payload format: `premium:{user_id}` / `unlock:{dialog_photo_id}`
- Idempotency: `telegram_payment_charge_id` unique в БД
- Emit `premium.purchased`, `photo.unlocked`
- Цены в Stars: конфигурируемые в env/config

## Acceptance criteria

- [ ] Premium invoice открывается в Telegram
- [ ] Успешная оплата → premium активен (задача 023)
- [ ] Unlock invoice → оригинал фото отправляется без blur
- [ ] Повторный webhook payment → idempotent, не двойной unlock
- [ ] Failed/cancelled payment → состояние не меняется

## Технические заметки

- Currency: `XTR` (Telegram Stars)
- `provider_token` пустой для Stars
- Тестирование в Telegram test environment
- Таблица `payments`: id, user_id, type, amount_stars, charge_id, created_at
- Refund policy — описать в правилах (030), техреализация refund optional

## Out of scope

- Fiat payments (Stripe, ЮKassa)
- Crypto
- Подписка с auto-renew (Telegram Stars limitations)
