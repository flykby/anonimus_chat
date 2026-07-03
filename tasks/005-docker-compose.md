# 005. Docker Compose (dev + prod stack)

**Статус:** done  
**Фаза:** infra  
**Зависимости:** 001

## Описание

Docker Compose для dev и prod VM: PostgreSQL, Redis, **Go**-сервисы bot/api/ai.

## Scope

- `docker-compose.yml` — postgres, redis, bot, api, ai (Go images, local build)
- `docker-compose.prod.yml` — registry images, restart policies
- Volumes, healthchecks, internal network
- RunPod env → `ai` service
- Dev override example (без uvicorn — сервисы на Go)

## Acceptance criteria

- [x] `docker compose up` поднимает postgres + redis
- [x] API и AI отвечают `GET /health` → 200
- [x] Postgres data persists (named volume)
- [x] Prod compose merge работает с deploy.sh
- [x] RunPod URL'ы в env `ai`, не hardcode

## Технические заметки

- Dockerfiles: `docker/{bot,api,ai}.Dockerfile` — multi-stage Go build из корня
- Healthchecks: `wget` на `/health`

## Out of scope

- Kubernetes
