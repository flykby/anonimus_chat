# Milestone 2 — Полный бот (без real AI)

**Цель:** весь пользовательский функционал бота работает end-to-end.  
**Исключение:** AI-общение — echo-заглушка (038), real LLM позже (016+).

## Уже готово

| # | Задача |
|---|--------|
| 001–004 | CI/CD, deploy |
| 005–008 | Infra, events |
| 010–011 | Регистрация, главное меню |
| 013–015 | Match routing, queue UX, end dialog |
| 024 | P2P matchmaking (M+M, F+M) |
| 025 | P2P relay & moderation |
| 026 | Profile view |
| 027 | Edit profile |
| 028 | Change language |
| 029 | Delete profile |
| 030 | Rules page |
| 012 | i18n RU/EN |
| 038 | AI echo stub |

## Отложено (после Milestone 2)

| Блок | Задачи |
|------|--------|
| Real AI | 036 RunPod, 016–019, 017 personas |
| Live F priority | 037 |
| Фото + Stars | 020–023, 031 |
| Метрики + launch | 032–035 |
| Webhook | 009 (long polling ок) |

## User journey (Milestone 2 done)

```mermaid
flowchart TD
    Start["/start"] --> Reg[Регистрация]
    Reg --> Menu[Главное меню]
    Menu --> Match[Начать разговор]
    Match --> AI[AI route: echo + end]
    Match --> P2P[P2P route: queue → match → relay + end]
    Menu --> Profile[Профиль]
    Menu --> Rules[Правила]
```

## Критерий закрытия Milestone 2

- [ ] Все 4 комбинации gender/seeking дают рабочий dialog (AI echo или P2P relay)
- [x] Профиль и правила — полноценные экраны
- [x] RU/EN переключается
- [ ] End dialog работает для AI и P2P
- [ ] Real AI не требуется
