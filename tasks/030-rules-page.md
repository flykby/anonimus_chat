# 030. Rules page

**Статус:** done  
**Фаза:** profile  
**Зависимости:** 011, 009

## Описание

Страница с правилами использования бота на RU и EN. Доступна из главного меню.

## Scope

- Тексты правил в `internal/bot/rules/rules_ru.md` / `rules_en.md` (embed)
- Разделы: возраст 18+, анонимность, P2P, AI, adult/premium, Stars, удаление, disclaimer
- HTML parse_mode для заголовков
- Кнопка «Назад» на последнем сообщении → главное меню
- Авто-разбиение на части если &gt; 4096 символов

## Acceptance criteria

- [x] Правила открываются из меню на языке пользователя
- [x] RU и EN версии полные и согласованные
- [x] «Назад» работает
- [x] Текст покрывает premium, adult, P2P moderation

## Технические заметки

- Package `internal/bot/rules`, `RulesVersion = v1`
- Юридически нейтральный шаблонный текст, review перед launch (035)

## Out of scope

- Юридическая экспертиза
- Multi-page rules с pagination UI
