# 009. i18n RU/EN

**Статус:** todo  
**Фаза:** bot  
**Зависимости:** 007

## Описание

Локализация всех пользовательских строк на русский и английский. Язык выбирается при регистрации и может быть изменён в профиле.

## Scope

- Файлы локализации: `bot/locales/ru.yaml`, `bot/locales/en.yaml`
- Функция `t(key, lang, **kwargs)` — gettext-style
- Покрыть строки:
  - Регистрация (все шаги, ошибки)
  - Главное меню
  - Очередь поиска
  - Диалог (завершение, подтверждение)
  - Профиль
  - Правила
  - Платежи и premium
- Fallback: RU если ключ не найден
- API endpoint: `GET /users/{id}/language`

## Acceptance criteria

- [ ] Все UI-строки бота вынесены в locale-файлы (нет хардкода в handlers)
- [ ] Переключение RU ↔ EN меняет все видимые тексты
- [ ] Параметризация работает: `t("queue.count", lang, n=5)` → «...подходящих: 5»
- [ ] Язык AI-диалога передаётся в AI service (persona отвечает на языке пользователя)

## Технические заметки

- Структура ключей: `registration.age.prompt`, `menu.start_chat`, `queue.searching`
- Не переводить user-generated content
- Правила (задача 027) — отдельные длинные тексты per lang
- При смене языка (задача 025) — invalidate cached strings в bot session

## Out of scope

- Другие языки кроме RU/EN
- Автоопределение языка по Telegram locale
