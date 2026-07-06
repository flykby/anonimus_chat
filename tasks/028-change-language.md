# 028. Change language

**Статус:** done  
**Фаза:** profile  
**Зависимости:** 026, 009

## Описание

Смена языка интерфейса бота и языка общения с AI-собеседником. Отдельный flow из профиля.

## Scope

- Кнопка «Сменить язык» → выбор RU / EN
- API: `PATCH /users/me/profile` `{ language: "ru" | "en" }`
- Немедленное обновление всех UI-строк
- Передача нового language в AI service для следующих сообщений
- Emit `user.profile_updated` field=language
- Подтверждение: «Язык изменён на RU/EN»

## Acceptance criteria

- [x] Смена RU→EN обновляет меню и профиль
- [x] AI-диалог продолжается на новом языке с следующего сообщения
- [x] P2P не затрагивается (язык только UI)
- [x] Выбор сохраняется после рестарта бота

## Технические заметки

- Обновить Redis FSM/cache lang key
- System prompt persona: «respond in user's language»
- Правила (030) показывать на новом языке

## Out of scope

- Автоперевод истории диалога
- Другие языки
