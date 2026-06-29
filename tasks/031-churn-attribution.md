# 031. Churn attribution

**Статус:** todo  
**Фаза:** metrics  
**Зависимости:** 005, 030

## Описание

Эвристическая атрибуция причин ухода пользователя из диалога: несовпадение характера, типажа после фото, исследование других персон.

## Scope

- Классификация при `dialog.ended`:
  - `persona_mismatch` — ended before first photo AND messages < 10
  - `appearance_mismatch` — ended within 2 min after first photo AND no unlock
  - `exploring` — ended after photo AND new dialog started within 24h
  - `quality_churn` — no return 7d AND only 1 dialog ever (weak signal)
  - `normal_end` — user confirmed after long dialog
  - `unknown` — fallback
- Batch job: daily attribution pass over ended dialogs
- Store in `events` или `dialog_attribution` table
- Admin report: % per reason per persona

## Acceptance criteria

- [ ] Каждый ended AI dialog получает attribution label
- [ ] Правила documented и versioned (`attribution_rules_v1`)
- [ ] Report показывает breakdown per persona
- [ ] «София -50% median» можно декомпозировать: % persona_mismatch vs appearance_mismatch
- [ ] False positives acceptable — метрика эвристическая

## Технические заметки

- Не показывать attribution пользователю
- Пороги в config для tuning без деплоя
- Cross-check: если photo_request_rate низкий но persona_mismatch высокий → проблема в промпте
- Если post_photo_churn высокий → фото/типаж

Пример decision tree:
```
if no_photo and msg_count < 10 → persona_mismatch
elif photo_sent and ended_within_120s and not unlocked → appearance_mismatch
elif photo_sent and new_dialog_24h → exploring
elif no_activity_7d and total_dialogs == 1 → quality_churn
else → normal_end or unknown
```

## Out of scope

- Прямое измерение «удалил приложение»
- User surveys
- Causal inference / ML
