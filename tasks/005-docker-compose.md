# 005. Docker Compose (dev + prod stack)

**Статус:** todo  
**Фаза:** infra  
**Зависимости:** 001

## Описание

Единый Docker Compose для dev и prod VM: PostgreSQL, Redis и контейнеры приложений. Postgres и Redis крутятся на той же виртуалке, что и bot/api/ai — без managed cloud DB.

## Scope

- `docker-compose.yml` — базовый стек: postgres, redis, api, bot, ai
- `docker-compose.prod.yml` — extends/overrides для prod VM (без dev mounts)
- Volumes для персистентности Postgres и Redis (AOF optional)
- Health checks: postgres, redis, bot, api, ai
- `docker/*.Dockerfile` — multi-stage, slim images (сборка через 003)
- Dev: hot-reload (volume mount + uvicorn --reload)
- Internal network: сервисы обращаются друг к другу по имени

## Acceptance criteria

- [ ] `docker compose up` поднимает postgres + redis без ошибок
- [ ] API и AI stubs отвечают на `GET /health` → `200 OK`
- [ ] Данные Postgres сохраняются между перезапусками (named volume)
- [ ] Тот же compose (с prod override) работает на VM после pull из registry (004)
- [ ] RunPod URL'ы прокидываются в `ai` через env, не hardcode

## Технические заметки

- Postgres 16, Redis 7 — контейнеры на prod VM
- `DATABASE_URL=postgresql://anonimus:anonimus@postgres:5432/anonimus`
- `REDIS_URL=redis://redis:6379/0`
- Dev webhook: ngrok / cloudflared → `WEBHOOK_URL`
- `docker-compose.override.yml` для локальных настроек (не коммитить секреты)
- GPU **не** нужен на VM — inference на RunPod

## Out of scope

- Kubernetes / swarm
- Managed Postgres/Redis в облаке
- Мониторинг — см. 035
