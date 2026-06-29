# 017. Persona prompts

**Статус:** todo  
**Фаза:** ai  
**Зависимости:** 006, 016

## Описание

System prompts для AI-персон: характер, стиль общения, границы поведения. Версионирование промптов для A/B и отката.

## Scope

- Таблица `personas`: name, gender, system_prompt, prompt_version, active
- Минимум 1 персона для dev (например «Алиса»)
- Структура промпта:
  - Имя, возраст, характер
  - Стиль сообщений (короткие, с эмодзи, flirty level)
  - Правила: не раскрывать что AI, не выходить за рамки персоны
  - Язык: отвечать на языке пользователя (RU/EN)
  - Когда соглашаться прислать фото (для задачи 019)
- Assign persona при AI match: round-robin или random среди active
- Admin: обновление prompt_version без деплоя (из БД)

## Acceptance criteria

- [ ] При AI match пользователю назначается persona
- [ ] Ответы соответствуют характеру из system prompt (manual QA 10 диалогов)
- [ ] `prompt_version` записывается в `dialog.started` metadata
- [ ] Неактивная persona (`active=false`) не назначается
- [ ] M+F и F+F получают female personas

## Технические заметки

- Промпты хранить в БД, не в коде (кроме default seed)
- Шаблон: `personas/seed/alisa.yaml` для миграции
- Persona assignment: `SELECT * FROM personas WHERE gender='female' AND active ORDER BY random() LIMIT 1`
- Для 5 персон и A/B — задача 032

## Out of scope

- Генерация фото персон
- Динамическая смена персоны mid-dialog
