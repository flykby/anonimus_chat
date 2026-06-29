# 003. CI pipeline (build, test, lint)

**Статус:** todo  
**Фаза:** milestone-1  
**Зависимости:** 001, 002

## Описание

Настроить CI на виртуалке через отдельный Docker-контейнер-сборщик: lint, тесты, сборка образов приложения, push в внутренний registry. Контейнер CI — изолированное пространство сборки, не смешивается с runtime-контейнерами бота.

## Scope

- `docker/ci.Dockerfile` — образ с Python 3.12, ruff, pytest, docker-cli (dind sidecar или socket mount)
- `Makefile` targets: `lint`, `test`, `build`, `push`
- `.github/workflows/ci.yml` **или** `scripts/ci.sh` + cron/webhook на VM (git push → запуск CI-контейнера)
- Pipeline stages:
  1. `lint` — ruff check + format check
  2. `test` — pytest (минимум smoke-тест echo handler)
  3. `build` — `docker build` для bot (и заготовки api/ai)
  4. `push` — tag `registry.internal/anonimus/bot:$GIT_SHA` + `:latest`
- Кеш pip/docker layers между прогонами
- Fail fast: lint → test → build → push

## Acceptance criteria

- [ ] Push в main/master запускает pipeline на VM (или через GitHub Actions runner на VM)
- [ ] Lint и тесты блокируют push образа при падении
- [ ] Успешный прогон публикует образ bot в internal registry
- [ ] Образ из registry запускается и проходит echo smoke test
- [ ] Секреты (registry credentials, BOT_TOKEN для e2e) не попадают в образ и git

## Технические заметки

### Схема на VM

```mermaid
flowchart LR
    Git[Git push] --> CI[CI runner container]
    CI --> Lint[ruff]
    CI --> Test[pytest]
    CI --> Build[docker build]
    Build --> Reg[(Internal registry)]
```

- **Registry:** self-hosted (Harbor, GitLab Registry, или `registry:2` в Docker) — URL в `REGISTRY_URL`
- **CI runner:** контейнер с `-v /var/run/docker.sock` для sibling builds **или** Docker-in-Docker
- Альтернатива без dind: `docker build` на хосте VM, CI-контейнер только lint/test, build через shell на хосте
- Tagging: `$GIT_SHA`, `$GIT_BRANCH`, `latest` только для main
- `.env.ci.example`: `REGISTRY_URL`, `REGISTRY_USER`, `REGISTRY_PASSWORD`

## Out of scope

- Деплой на prod VM (задача 004)
- Multi-arch builds (arm64) — при необходимости позже
- SAST/dependency scanning — добавить в 035
