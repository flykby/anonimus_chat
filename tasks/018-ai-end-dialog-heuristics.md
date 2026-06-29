# 018. AI end dialog heuristics

**Статус:** todo  
**Фаза:** ai  
**Зависимости:** 016, 015

## Описание

Механизмы завершения диалога со стороны AI: не жёсткое правило, а эвристики и опциональный tool call для завершения при токсичности или сильном негативе.

## Scope

- **Таймаут неактивности:** 2ч без сообщений → auto-end, emit `dialog.ended` reason=`timeout`
- **Пинг:** 10 мин без ответа → «Ты ещё здесь?» (один раз)
- **Короткие ответы:** 3 сообщения подряд < 5 символов → мягкий вопрос «всё ок?»
- **Tool `end_conversation`:** LLM может вызвать с reason (`toxic`, `user_bored`, `inappropriate`)
  - Лимит: не чаще 1 раза за dialog
  - Только при confidence / явных триггерах в промпте
- **Негатив-классификатор:** lightweight check на «бот», «фейк», оскорбления → не end, а смена тона в промпте

## Acceptance criteria

- [ ] Auto-end после 2ч неактивности работает
- [ ] Tool end_conversation закрывает dialog и возвращает user в меню
- [ ] Tool нельзя вызвать дважды в одном dialog
- [ ] Эвристики логируются в events metadata
- [ ] User-initiated end (015) имеет приоритет над AI heuristics

## Технические заметки

- Background job (cron / asyncio task) для timeout check
- `end_conversation` в OpenAI tools schema:
  ```json
  { "name": "end_conversation", "parameters": { "reason": "string" } }
  ```
- Не завершать диалог агрессивно — лучше недержать, чем перержать
- Метрики: `ai_initiated_end_rate` per persona

## Out of scope

- ML-модель churn prediction
- Бан пользователя за токсичность (можно v2)
