# Backlog — anonimus_chat

Telegram-бот для анонимного общения и обмена фотографиями.

## Стек

| Компонент | Технология |
|-----------|------------|
| **bot, api, ai** | Go 1.22+ |
| **Миграции БД** | [goose](https://github.com/pressly/goose) (SQL) |
| **Postgres / Redis** | Docker на VM |
| **Inference** | RunPod (HTTP) |

## Принципы деплоя

- **Prod VM:** все runtime-сервисы (bot, api, ai, **postgres, redis**) — Docker-контейнеры на виртуалке
- **CI:** отдельный build-контейнер на VM — lint, test, `docker build`, push в **internal registry**
- **Inference:** chat LLM и embedding-модель — **RunPod** (HTTP API), не на prod VM

## Архитектура

```mermaid
flowchart TD
    subgraph vm [Production VM — Docker]
        Proxy[nginx/caddy]
        Bot[bot]
        API[api]
        AI[ai]
        PG[(postgres)]
        Redis[(redis)]
    end
    subgraph cicd [CI on VM]
        CI[ci runner container]
        Reg[(internal registry)]
    end
    subgraph runpod [RunPod]
        LLM[Chat LLM pod]
        EMB[Embedding pod]
    end
    Git[git push] --> CI
    CI --> Reg
    Reg --> vm
    TG[Telegram] --> Proxy --> Bot
    Bot --> API
    API --> PG
    API --> Redis
    API --> AI
    AI --> LLM
    AI --> EMB
```

## Шаблон задачи

Каждый файл: статус, фаза, зависимости, описание, scope, acceptance criteria, технические заметки, out of scope.

---

## Milestone 1 — Echo bot + CI/CD (старт здесь)

| # | Задача | Статус |
|---|--------|--------|
| 001 | [Project scaffold](001-project-scaffold.md) | done |
| 002 | [Echo bot](002-echo-bot.md) | done |
| 003 | [CI pipeline (build, test, lint)](003-ci-pipeline.md) | done |
| 004 | [VM deploy + internal registry](004-vm-deploy-registry.md) | done |

**Результат milestone 1:** echo-бот в Telegram, образ собирается в CI, деплоится на VM из registry.

---

## Milestone 2 — Полный бот ✅ feature-complete

> Весь функционал бота end-to-end реализован. AI = echo-заглушка (038).  
> Осталась E2E-верификация на проде. Детали: [MILESTONE-2-full-bot.md](MILESTONE-2-full-bot.md)

| # | Задача | Статус |
|---|--------|--------|
| 012 | [i18n RU/EN](012-i18n-ru-en.md) | done |
| 013–015 | Match + queue + end | done |
| 024 | [P2P matchmaking](024-p2p-matchmaking.md) | done |
| 025 | [P2P relay](025-p2p-relay-moderation.md) | done |
| 026 | [Profile view](026-profile-view.md) | done |
| 027–029 | Edit / language / delete | done |
| 030 | [Rules page](030-rules-page.md) | done |
| 038 | [AI echo stub](038-ai-echo-stub.md) | done |

---

## Milestone 3 — что дальше (выбрать направление)

Рекомендуемый порядок после закрытия M2:

| Приоритет | Блок | Задачи | Зачем |
|-----------|------|--------|-------|
| **A** | Real AI | 036 → 016 → 017 → 018 → 019 (+ **039** prompts toolkit) | Echo → настоящие диалоги с персонами |
| **B** | Монетизация | 020 → 021 → 022 → 023 → 031 | Фото, blur, Stars, premium |
| **C** | Match UX | 037 | Live F priority для M→F |
| **D** | Ops | 009, 032–035 | Webhook, метрики, launch checklist |

```mermaid
flowchart LR
    M2[M2 Full Bot done] --> E2E[E2E smoke prod]
    E2E --> A[036+016 Real AI]
    E2E --> B[020+ Photos Stars]
    A --> Launch[035 Launch]
    B --> Launch
```

**Текущий фокус:** закрыть E2E Milestone 2, затем **036 RunPod + 016 AI dialog** (если приоритет — умные диалоги) или **020–023** (если приоритет — фото и деньги).

---

## Фаза 0 — Инфраструктура данных (005–008)

| # | Задача | Статус |
|---|--------|--------|
| 005 | [Docker Compose (dev + prod stack)](005-docker-compose.md) | done |
| 006 | [Database schema](006-database-schema.md) | done |
| 007 | [Redis queues & sessions](007-redis-queues-sessions.md) | done |
| 008 | [Event logging](008-event-logging.md) | done |

---

## Фаза 1 — Telegram Bot (009–012)

| # | Задача | Статус |
|---|--------|--------|
| 009 | [Telegram bot webhook](009-telegram-bot-webhook.md) | done |
| 010 | [Registration FSM](010-registration-fsm.md) | done |
| 011 | [Main menu](011-main-menu.md) | done |
| 012 | [i18n RU/EN](012-i18n-ru-en.md) | done |

---

## Фаза 2 — Диалоги и матчинг (013–015)

| # | Задача | Статус |
|---|--------|--------|
| 013 | [Match routing](013-match-routing.md) | done |
| 014 | [Queue UX](014-queue-ux.md) | done |
| 015 | [End dialog flow](015-end-dialog-flow.md) | done |
| 037 | [Live F priority for M→F](037-live-f-priority.md) | todo |

---

## Фаза 3 — AI + RunPod (036, 016–019)

| # | Задача | Статус |
|---|--------|--------|
| 036 | [RunPod inference (LLM + embeddings)](036-runpod-inference.md) | todo |
| 016 | [AI dialog service](016-ai-dialog-service.md) | todo |
| 017 | [Persona prompts](017-persona-prompts.md) | todo |
| 039 | [Persona generation prompts & templates](039-persona-generation-prompts.md) | todo |
| 018 | [AI end dialog heuristics](018-ai-end-dialog-heuristics.md) | todo |
| 019 | [Photo intent classifier](019-photo-intent-classifier.md) | todo |

---

## Фаза 4 — Фото и монетизация (020–023)

| # | Задача | Статус |
|---|--------|--------|
| 020 | [Photo catalog (+ embedding search)](020-photo-catalog.md) | todo |
| 021 | [Photo delivery & blur](021-photo-delivery-blur.md) | todo |
| 022 | [Telegram Stars payments](022-telegram-stars-payments.md) | done |
| 023 | [Premium logic](023-premium-logic.md) | done |

---

## Фаза 5 — P2P (024–025)

| # | Задача | Статус |
|---|--------|--------|
| 024 | [P2P matchmaking](024-p2p-matchmaking.md) | done |
| 025 | [P2P relay & moderation](025-p2p-relay-moderation.md) | done |

---

## Фаза 6 — Профиль и правила (026–031)

| # | Задача | Статус |
|---|--------|--------|
| 026 | [Profile view](026-profile-view.md) | done |
| 027 | [Edit profile](027-edit-profile.md) | done |
| 028 | [Change language](028-change-language.md) | done |
| 029 | [Delete profile anti-abuse](029-delete-profile-antiabuse.md) | done |
| 030 | [Rules page](030-rules-page.md) | done |
| 031 | [Premium purchase menu](031-premium-purchase-menu.md) | done |

---

## Фаза 7 — Персоны и метрики (032–034)

| # | Задача | Статус |
|---|--------|--------|
| 032 | [Personas rollout](032-personas-rollout.md) | todo |
| 033 | [Metrics: median dialog duration](033-metrics-median-dialog-duration.md) | todo |
| 034 | [Churn attribution](034-churn-attribution.md) | todo |

---

## Фаза 8 — Запуск (035)

| # | Задача | Статус |
|---|--------|--------|
| 035 | [Traffic launch checklist](035-traffic-launch-checklist.md) | todo |

---

## Зависимости фаз

```mermaid
flowchart LR
    M2[Milestone 2 Full Bot] --> M3[Real AI 016+036]
    M2 --> M4[Photos 020+]
    M2 --> M37[037 Live F]
    M3 --> M7[Metrics 032+]
    M4 --> M7
    M7 --> Launch[035 Launch]
```

**Текущий фокус:** Milestone 2 закрыт по коду — E2E smoke, затем Milestone 3 (Real AI или Photos).  
P2P (024–025) идёт **до** RunPod (036). Live F (037) — после M2.
