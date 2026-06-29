# 016. Photo intent classifier

**Статус:** todo  
**Фаза:** ai  
**Зависимости:** 013

## Описание

Определить, просит ли пользователь фотографию, и извлечь теги для подбора из каталога. В v1 — через LLM structured output, без отдельной embedding-модели.

## Scope

- Классификация на каждом user message (или batch с ответом персоны)
- Structured output schema:
  ```json
  {
    "is_photo_request": true,
    "tags": ["selfie", "smile"],
    "nsfw_hint": "safe" | "adult" | "none"
  }
  ```
- Если `is_photo_request` — trigger photo delivery pipeline (задача 018)
- Персона в текстовом ответе соглашается/играет роль («ловлю на слове, держи 😊»)
- Emit `photo.requested` с tags
- Fallback: keyword list («фото», «скинь», «покажи», «pic», «photo») если LLM fail

## Acceptance criteria

- [ ] «Скинь селфи» → is_photo_request=true, tags содержит selfie
- [ ] «Как дела?» → is_photo_request=false
- [ ] EN: «send me a pic» → корректная классификация
- [ ] Latency классификации < 2 сек (можно в том же LLM call)
- [ ] False positive rate < 10% на тестовом наборе из 50 фраз

## Технические заметки

- Оптимизация: один LLM call → `{ reply, photo_intent }` JSON mode
- Альтернатива: отдельный cheap call (gpt-4o-mini) только для intent
- Embedding-модель — отложить до 100k+ msg/day
- nsfw_hint влияет на выбор safe vs adult из каталога

## Out of scope

- Embedding-based classifier
- Computer vision на входящих фото пользователя
- Генерация фото on-the-fly
