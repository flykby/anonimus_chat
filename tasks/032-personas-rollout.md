# 032. Personas rollout

**Статус:** todo  
**Фаза:** metrics  
**Зависимости:** 017, 020

## Описание

Запуск 5 AI-персон с разными характерами и внешностью. Offline-оценка качества фото («пластик / не пластик») и загрузка ассетов в бота.

## Scope

- 5 персон: имена, промпты, gender=female
- Разные типажи: характер (застенчивая, дерзкая, etc.) + набор фото
- Assignment strategy: round-robin или weighted random для равномерного трафика
- Offline benchmark:
  - Датасет 50–100 фото
  - Ручная или модельная оценка «looks plastic»
  - Порог качества для prod
- Offline контент-пайплайн:
  - Генерация фото (вне codebase)
  - Upload script → telegram file_id → БД (020)
  - Минимум 10 safe + 10 adult на персону
- A/B: `persona_id` в каждом `dialog.started` для метрик

## Acceptance criteria

- [ ] 5 active personas в production
- [ ] Каждая persona имеет уникальный prompt_version
- [ ] Фото-каталог заполнен для всех 5
- [ ] Benchmark документирован: % plastic per persona
- [ ] Трафик распределяется между персонами (±10% за неделю)
- [ ] Одна persona может быть disabled без деплоя

## Технические заметки

- Имена примеры: Алиса, Василиса, София, Катя, Мила
- Не коммитить сгенерированные изображения в git (только scripts)
- `personas.weight` для controlled rollout
- Качество «пластик» — субъективно, использовать median_dialog_duration как главную метрику (033)

## Out of scope

- Real-time photo generation
- User-facing persona picker
- Male AI personas
