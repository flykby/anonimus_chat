# Handoff — anonimus_chat

Документ для продолжения работы в новом чате / новым агентом.  
Обновлено: 2026-07-06.

---

## Проект

Telegram-бот для анонимного общения и обмена фотографиями.

| | |
|---|---|
| **Репозиторий** | `github.com/flykby/anonimus_chat` |
| **Prod VM** | `/opt/anonimus_chat` |
| **Backlog** | [`tasks/README.md`](../tasks/README.md) |
| **Milestone 2** | [`tasks/MILESTONE-2-full-bot.md`](../tasks/MILESTONE-2-full-bot.md) |
| **Deploy** | [`docs/deploy.md`](deploy.md) |

### Стек

| Компонент | Технология |
|-----------|------------|
| bot, api, ai | Go 1.24 |
| БД | Postgres + [goose](https://github.com/pressly/goose) (SQL-миграции) |
| Кэш/очереди | Redis |
| Деплой | Docker на VM, internal registry (GHCR) |
| Inference (будущее) | RunPod HTTP API |

### Архитектура

```
Telegram → bot → api → postgres / redis
                  ↓
                 ai → RunPod (позже)
```

**CI/CD:** push в `main` → lint/test → docker build → push registry → SSH deploy на VM.  
**Деплой:** `scripts/remote-deploy.sh` → `scripts/deploy.sh` → postgres healthy → **goose migrate up** → compose up → bot healthy.  
Если health после `compose up` падает — авто-откат миграций + образов с последнего успешного tag (`.deploy/current`), CI exit ≠ 0.

> **⚠️ Главное правило git:** все изменения **пушить сразу в `main`**.  
> Не создавать feature branches и PR — CI auto-deploy срабатывает только на push в `main`.  
> Workflow: `git checkout main` → правки → `go test ./...` → commit → `git push origin main`.

---

## Milestone 2 — ✅ feature-complete (по коду)

Весь пользовательский функционал бота реализован end-to-end. AI = echo-заглушка (038), real LLM — позже (016+).

### Сделанные задачи (последние коммиты на `main`)

| # | Задача | Коммит | Что сделано |
|---|--------|--------|-------------|
| 012 | i18n RU/EN | `d524725` | `locales.T()`, `ru.yaml`/`en.yaml`, рефакторинг строк бота |
| — | CI fix | `4e9cf78`, `00122fc` | gofmt + pin `go-internal v1.14.1` для Go 1.24 |
| 027 | Edit profile | `05e6b25` | `PATCH /users/me/profile` (age/gender/seeking), bot FSM `edit:*`, event `user.profile_updated` |
| 028 | Change language | `e61de20` | PATCH с `language`, bot flow `lang:ru`/`lang:en` из профиля |
| 029 | Delete profile | `7585ea5` | Soft-delete, double confirm, end active dialog, notify P2P partner |
| — | Docs | `a0456cc` | Обновлён backlog, M2 закрыт по коду |
| — | Nav screen UX | `61169a5` | Замена экранов вместо накопления сообщений (PR #8) |

### Ранее в M2 (уже на main)

- **001–004** — CI/CD, deploy
- **005–008** — infra, events
- **010–011** — регистрация FSM, главное меню
- **013–015** — match routing, queue UX, end dialog
- **024–025** — P2P matchmaking + relay/moderation
- **026** — profile view
- **030** — rules page
- **038** — AI echo stub

---

## Nav screen UX (done, PR #8 merged)

При навигации предыдущее сообщение бота удаляется, показывается новый экран:

```
Главное меню → Правила → (меню стёрлось) → Правила
Назад → (правила стёрлись) → Главное меню
```

- `internal/redis/navscreen/` — Redis store ID сообщений per user
- `internal/bot/handlers/navigation.go` — `showNavScreen`, `clearNavScreen`, `deleteUserMessage`
- Подключено: menu, rules, profile, edit, language, delete, queue, end dialog
- Reply-кнопки пользователя («Профиль», «Правила» и т.д.) тоже удаляются
- **Не затронуто:** регистрация, диалоги, P2P relay

---

## Маршрутизация матчинга (актуальная)

| Пол | Ищу | Маршрут |
|-----|-----|---------|
| M | F | AI (echo stub) |
| M | M | P2P |
| F | F | AI (echo stub) |
| F | M | P2P |

> **F seeking M = P2P** (как M seeking M), не AI.

---

## Правила и договорённости

### База данных и удаление профиля

1. **Soft-delete только через `deleted_at`**, НЕ через boolean `is_deleted`.
2. Миграция `00003_soft_delete_telegram_unique.sql`: partial unique index `telegram_id WHERE deleted_at IS NULL` — позволяет re-register после удаления.
3. После удаления `/start` → новая регистрация с тем же `telegram_id`.
4. **Premium не восстанавливается** на новом профиле.
5. `deletion_benefits` привязан к **telegram_id**, не к user row — пересоздание профиля не сбрасывает флаг.
6. UI бесплатного unlock adult-фото при удалении — **отложен до задачи 021** (инфра `deletion_benefits` уже есть).
7. Активный dialog при удалении → force end + уведомление P2P-партнёра.

### Миграции

- **Автоматически прогоняются при деплое** через `scripts/deploy.sh` → `migrate-prod.sh up`.
- Флаг `--skip-migrate` только для emergency.
- Новые миграции — goose SQL в `migrations/`.

### i18n

- Все UI-строки бота — через `locales.T(key, lang, params)` и YAML `internal/bot/locales/ru.yaml`, `en.yaml`.
- Fallback: RU если ключ не найден в EN.
- Правила (030) — отдельные markdown per lang.
- `menu.LabelsFor(lang)` строится из locales.
- Нет хардкода строк в handlers.

### Bot UX / навигация

- **Nav screen:** один «экран» за раз — старое сообщение бота удаляется перед показом нового.
- Redis key: `anonimus:navscreen:{telegramID}`.
- Многочастные правила — все message ID хранятся и удаляются вместе.
- Диалоги и P2P relay **не используют** nav screen (отдельная логика сообщений).

### События (event logging)

- Emit events при значимых действиях: `user.profile_updated`, `user.deleted`, dialog events и т.д.
- Схема в `internal/events/`.

### CI / деплой

- Go **1.24**, `go-internal` pinned `v1.14.1`.
- Перед push: `gofmt`, `make lint`, `make test`, `make build`.
- Деплой только из `main` (CI auto-deploy через SSH).
- При падении CI — сначала логи GitHub Actions; типичные проблемы: gofmt, lint.

### Git workflow

- **Пушить сразу в `main`** — без feature branches и без PR.
- Перед push: `gofmt`, `make test` (или `go test ./...`), commit с понятным сообщением.
- После push в `main` — CI автоматически деплоит на prod.
- Минимальный scope diff — не трогать несвязанный код.

### Код-стиль

- Следовать существующим паттернам проекта (handlers, apiclient, redis stores).
- Комментарии только для неочевидной логики.
- Тесты — для новой бизнес-логики и redis stores; не писать тривиальные тесты.
- Не over-engineer: простое решение лучше абстракций на 2 строки.

### Структура ключевых файлов

```
cmd/bot/main.go              — wiring bot + redis stores
cmd/api/main.go              — HTTP API
internal/bot/handlers/       — все bot handlers
internal/bot/locales/        — i18n YAML
internal/bot/menu/           — клавиатуры, labels, actions
internal/api/                — REST handlers (users, match, dialogs)
internal/db/                 — postgres queries
internal/redis/              — redis stores (fsm, navscreen, matchqueue, ...)
migrations/                  — goose SQL
tasks/                       — backlog (README.md — главный индекс)
scripts/deploy.sh            — prod deploy + migrate
```

---

## Что не закрыто

### Milestone 2 closure

- [ ] **E2E smoke на проде:**
  - все 4 комбинации gender/seeking → рабочий dialog (AI echo или P2P)
  - end dialog для AI и P2P (partner notified в P2P)
  - profile flow: view → edit → language → delete → re-register
  - nav screen UX (Главное меню ↔ Правила ↔ Профиль ↔ Назад)

### Отложено (не в scope M2)

| Блок | Задачи | Заметки |
|------|--------|---------|
| Real AI | 036 → 016 → 017–019 | RunPod LLM вместо echo |
| Фото + Stars | 020–023, 031 | Каталог, blur, оплата, premium UI |
| Delete unlock UI | часть 029 + 021 | Бесплатный unlock adult-фото |
| Live F priority | 037 | Приоритет живых F в очереди M→F |
| Webhook | 009 | ✅ Реализован, использовать по желанию |
| Ops/Launch | 032–035 | Personas, метрики, launch checklist |

---

## Milestone 3 — рекомендуемый порядок

| Приоритет | Блок | Задачи | Зачем |
|-----------|------|--------|-------|
| **A** | Real AI | 036 → 016 → 017 → 018 → 019 | Echo → настоящие диалоги с персонами |
| **B** | Монетизация | 020 → 021 → 022 → 023 → 031 | Фото, blur, Stars, premium |
| **C** | Match UX | 037 | Live F priority для M→F |
| **D** | Ops | 009, 032–035 | Webhook, метрики, launch checklist |

**Рекомендуемый next step после закрытия M2:**

1. E2E smoke на проде (включая nav screen UX)
2. Выбрать M3: **036+016 (Real AI)** или **020–023 (Photos/Stars)**

---

## Как работать с пользователем

- Подтверждает задачи коротко: «да», «давай», «делаем».
- Одна задача = один focused commit **в `main`**.
- **Не делать PR** — commit + push в `main` напрямую.
- При падении деплоя — сначала диагностика CI, потом fix.
- Документацию (`tasks/`) обновлять после завершения milestone/крупных задач.
- Отвечать на русском, если пользователь пишет на русском.

---

## Быстрые команды

```bash
# Тесты
make test          # или go test ./...

# Линт + формат
gofmt -w .
make lint

# Локально
docker compose up

# Деплой (на VM)
bash scripts/remote-deploy.sh <tag>
```

---

## Промпт для нового чата

Скопируй в начало нового чата:

> Продолжаем разработку **anonimus_chat**. Прочитай `docs/HANDOFF.md`, `tasks/README.md` и `tasks/MILESTONE-2-full-bot.md`. Следуй всем правилам из HANDOFF — **особенно: пушить сразу в `main`, без PR**. Текущий фокус: [указать задачу, например «E2E smoke на проде» или «задача 036 RunPod»].
