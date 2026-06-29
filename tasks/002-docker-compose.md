# 002. Docker Compose

**Статус:** todo  
**Фаза:** infra  
**Зависимости:** 001

## Описание

Поднять локальное окружение для разработки: PostgreSQL, Redis и контейнеры приложений. Единая команда для старта всего стека.

## Scope

- `docker-compose.yml`: postgres, redis, api, bot, ai
- Volumes для персистентности Postgres
- Health checks для postgres и redis
- Проброс портов: API `8000`, AI `8001` (опционально)
- `docker/Dockerfile` для каждого сервиса (multi-stage, slim image)
- Hot-reload для dev (volume mount + uvicorn --reload)

## Acceptance criteria

- [ ] `docker compose up` поднимает postgres + redis без ошибок
- [ ] API отвечает на `GET /health` → `200 OK`
- [ ] Сервисы видят друг друга по internal network (`api`, `redis`, `postgres`)
- [ ] Данные Postgres сохраняются между перезапусками (volume)

## Технические заметки

- Postgres 16, Redis 7
- `DATABASE_URL=postgresql://anonimus:anonimus@postgres:5432/anonimus`
- `REDIS_URL=redis://redis:6379/0`
- Для webhook в dev: ngrok / cloudflared tunnel → `WEBHOOK_URL`
- Отдельный `docker-compose.override.yml` для dev-настроек (не коммитить секреты)

## Out of scope

- Production orchestration (K8s, managed DB)
- Мониторинг (Prometheus/Grafana) — см. задачу 032
