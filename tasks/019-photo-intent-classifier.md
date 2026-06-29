# 019. Photo intent classifier

**Статус:** todo  
**Фаза:** ai  
**Зависимости:** 016

## Описание

Определить, просит ли пользователь фотографию, и извлечь теги/семантику для подбора из каталога (020). v1 — LLM structured output на RunPod; v2 — комбинация LLM + embedding similarity.

## Scope

- Классификация на каждом user message (или batch с ответом персоны)
- Structured output schema:
  ```json
  {
    "is_photo_request": true,
    "tags": ["selfie", "smile"],
    "semantic_query": "selfie near window",
    "nsfw_hint": "safe" | "adult" | "none"
  }
  ```
- Если `is_photo_request` — trigger photo delivery (021)
- Персона в текстовом ответе соглашается («держи 😊»)
- Emit `photo.requested` с tags + semantic_query
- Fallback: keyword list если LLM fail

## Acceptance criteria

- [ ] «Скинь селфи» → is_photo_request=true
- [ ] «Как дела?» → is_photo_request=false
- [ ] EN: «send me a pic» → корректная классификация
- [ ] Latency < 2 сек (можно в том же LLM call на RunPod)
- [ ] semantic_query используется в 020 для embedding search

## Технические заметки

- Один LLM call → `{ reply, photo_intent }` JSON mode (RunPod, 036)
- `semantic_query` → embedding через RunPod → search в 020
- nsfw_hint влияет на safe vs adult в каталоге

## Out of scope

- Computer vision на входящих фото пользователя
- Генерация фото on-the-fly
