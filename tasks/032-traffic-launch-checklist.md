# 032. Traffic launch checklist

**Статус:** todo  
**Фаза:** launch  
**Зависимости:** 029, 030, 031, 027

## Описание

Чеклист готовности к запуску трафика: инфраструктура, мониторинг, лимиты, legal, rollback plan.

## Scope

### Infrastructure
- [ ] Production webhook HTTPS на стабильном домене
- [ ] Managed Postgres + backups (daily)
- [ ] Redis persistent или managed
- [ ] Secrets в env/vault, не в repo
- [ ] LLM API rate limits и fallback provider

### Monitoring
- [ ] Uptime check на `/health`
- [ ] Alerting: error rate > 5%, LLM latency p95 > 10s
- [ ] Log aggregation (stdout → Loki/CloudWatch)
- [ ] Dashboard 030 подключен

### Product
- [ ] 5 personas active, фото загружены
- [ ] Rules (027) опубликованы
- [ ] Premium pricing финализирован
- [ ] P2P moderation channel настроен

### Safety & Legal
- [ ] 18+ gate работает
- [ ] Telegram ToS compliance review
- [ ] Report flow протестирован
- [ ] Privacy: что логируем, retention 30d

### Limits
- [ ] Rate limits: messages, match starts, photo requests
- [ ] Max concurrent dialogs per user = 1
- [ ] LLM cost cap per day (circuit breaker)

### Rollback
- [ ] `active=false` все personas → maintenance message
- [ ] Previous docker image tag documented
- [ ] DB migration rollback plan

### Load smoke test
- [ ] 10 concurrent AI dialogs stable
- [ ] 5 concurrent P2P pairs stable

## Acceptance criteria

- [ ] Все пункты чеклиста пройдены или явно waived с причиной
- [ ] Sign-off документирован (дата, версия)
- [ ] Первые 100 users мониторятся вручную 48h

## Технические заметки

- Soft launch: invite-only channel перед open traffic
- Kill switch env: `MAINTENANCE_MODE=true`
- Cost estimate: LLM $/dialog, Stars revenue target

## Out of scope

- Marketing campaign
- App Store (Telegram only)
- International expansion beyond RU/EN
