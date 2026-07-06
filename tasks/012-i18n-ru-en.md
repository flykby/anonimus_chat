# 012. i18n RU/EN

**Статус:** done  
**Фаза:** bot  
**Зависимости:** 010

## Описание

Локализация всех пользовательских строк на русский и английский. Язык выбирается при регистрации и хранится в профиле.

## Scope

- Файлы локализации: `internal/bot/locales/ru.yaml`, `en.yaml`
- Функция `locales.T(key, lang, params)` — gettext-style с `{param}` подстановкой
- Покрыты строки: регистрация, меню, очередь, диалог, профиль, P2P, common errors
- Fallback: RU если ключ не найден в EN
- API: `GET /users/by-telegram/{id}/language`, `GET /users/me/language?telegram_id=`
- Правила (030) — отдельные markdown-файлы per lang

## Acceptance criteria

- [x] Все UI-строки бота вынесены в locale-файлы (нет хардкода в handlers)
- [x] Переключение RU ↔ EN меняет все видимые тексты (через profile.language)
- [x] Параметризация работает: `queue.waiting` с `{count}`, `{seeking}`
- [x] Язык пользователя доступен через API (AI service использует profile.language из БД)

## Технические заметки

- Структура ключей: `registration.age.prompt`, `menu.start_chat`, `queue.waiting`
- `menu.LabelsFor(lang)` строится из locales
- Смена языка в рантайме — задача 028

## Out of scope

- Другие языки кроме RU/EN
- Автоопределение языка по Telegram locale
