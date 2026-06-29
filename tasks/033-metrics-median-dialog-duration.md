# 033. Metrics: median dialog duration

**Статус:** todo  
**Фаза:** metrics  
**Зависимости:** 008, 032

## Описание

Главная продуктовая метрика: медианная длительность диалога (`median_dialog_duration`) в разрезе persona_id. Сравнение удержания между персонами.

## Scope

- SQL view или scheduled query:
  ```sql
  SELECT persona_id,
         percentile_cont(0.5) WITHIN GROUP (ORDER BY duration_sec) AS median_duration,
         percentile_cont(0.9) WITHIN GROUP (ORDER BY duration_sec) AS p90_duration,
         COUNT(*) AS dialog_count
  FROM dialogs
  WHERE type = 'ai' AND ended_at IS NOT NULL
    AND message_count >= 3
    AND started_at > NOW() - INTERVAL '30 days'
  GROUP BY persona_id
  ```
- Дополнительные метрики per persona:
  - `messages_per_dialog` (avg)
  - `photo_request_rate`
  - `post_photo_churn_rate` (end within 2 min after first photo)
  - `unlock_rate`
- Grafana dashboard или простой admin HTML
- Alert: persona median < 50% от лучшей за 7 дней

## Acceptance criteria

- [ ] median_dialog_duration считается per persona_id
- [ ] Фильтр min_messages >= 3 применяется
- [ ] Dashboard обновляется минимум daily
- [ ] Можно сравнить Алису vs Василису за выбранный период
- [ ] P2P диалоги исключены из persona metrics

## Технические заметки

- `duration_sec` и `message_count` в `dialog.ended` metadata или columns
- Окна: 7d, 30d rolling
- Baseline persona = max median за период
- Export CSV для offline analysis

## Out of scope

- Real-time dashboard sub-second
- ML anomaly detection
