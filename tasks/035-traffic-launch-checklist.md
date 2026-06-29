# 035. Traffic launch checklist

**Статус:** todo  
**Фаза:** launch  
**Зависимости:** 032, 033, 034, 030

## Описание

Чеклист готовности к запуску трафика: VM + Docker инфраструктура, RunPod, мониторинг, лимиты, rollback.

## Scope

### Infrastructure (VM + Docker)
- [ ] Prod webhook HTTPS на стабильном домене (009, 004)
- [ ] Postgres + Redis в Docker на VM, daily backups
- [ ] Internal registry доступен, documented image tags
- [ ] CI pipeline (003) green on main
- [ ] Secrets в `.env` на VM, не в repo
- [ ] RunPod LLM + embedding pods: uptime, keep-alive policy, cost cap

### Monitoring
- [ ] Uptime check на `/health`
- [ ] Alerting: error rate > 5%, RunPod latency p95 > 10s
- [ ] Log aggregation (stdout → Loki или аналог)
- [ ] Dashboard 033 подключен

### Product
- [ ] 5 personas active, фото загружены (032)
- [ ] Rules (030) опубликованы
- [ ] Premium pricing финализирован
- [ ] P2P moderation channel настроен

### Safety & Legal
- [ ] 18+ gate работает
- [ ] Telegram ToS compliance review
- [ ] Report flow протестирован
- [ ] Privacy: retention 30d

### Limits
- [ ] Rate limits: messages, match starts, photo requests
- [ ] Max concurrent dialogs per user = 1
- [ ] RunPod cost cap per day (circuit breaker в 036)

### Rollback
- [ ] `active=false` personas → maintenance message
- [ ] Previous docker image tag documented (`deploy.sh --tag`)
- [ ] DB migration rollback plan

### Load smoke test
- [ ] 10 concurrent AI dialogs stable (RunPod + VM)
- [ ] 5 concurrent P2P pairs stable

## Acceptance criteria

- [ ] Все пункты пройдены или явно waived с причиной
- [ ] Sign-off документирован
- [ ] Первые 100 users мониторятся 48h

## Технические заметки

- Soft launch: invite-only перед open traffic
- Kill switch: `MAINTENANCE_MODE=true`
- Cost: RunPod $/hour + $/dialog vs Stars revenue

## Out of scope

- Marketing campaign
- International expansion beyond RU/EN
