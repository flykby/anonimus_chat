# 027. Edit profile

**Статус:** todo  
**Фаза:** profile  
**Зависимости:** 026

## Описание

Редактирование полей анкеты: возраст, пол, кого ищет. Язык редактируется отдельно (задача 028).

## Scope

- Меню выбора поля: Возраст / Пол / Ищу
- FSM edit flow per field (аналогично регистрации)
- API: `PATCH /users/me/profile` partial update
- Валидация возраста 18–99
- Emit `user.profile_updated` с `{ field, old, new }`
- Предупреждение при смене seeking во время активного dialog — сначала завершить диалог
- Пересчёт match route при смене пол/ищу

## Acceptance criteria

- [ ] Каждое поле редактируется независимо
- [ ] Язык недоступен в этом меню (только через 025)
- [ ] Изменения сразу видны в профиле
- [ ] Невалидный возраст отклоняется
- [ ] Смена seeking с F+M на F+F меняет доступные маршруты

## Технические заметки

- Callback prefix: `edit:age`, `edit:gender`, `edit:seeking`
- Не сбрасывать premium при edit
- История изменений optional в events metadata

## Out of scope

- Смена языка (028)
- Смена public_uuid
