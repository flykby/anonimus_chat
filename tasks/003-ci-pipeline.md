# 003. CI pipeline (build, test, lint)

**Статус:** done  
**Фаза:** milestone-1  
**Зависимости:** 001, 002

## Описание

CI на VM через Docker-контейнер-сборщик: lint, тесты, сборка Go-образов, push в internal registry.

## Scope

- `docker/ci.Dockerfile` — Go 1.22, golangci-lint, docker-cli, goose
- `Makefile`: `lint`, `test`, `build`, `build-docker`, `push`
- `.github/workflows/ci.yml` + `scripts/ci.sh`
- Pipeline: `golangci-lint` → `go test ./...` → `docker build` → push
- Tag: `registry.internal/anonimus/{bot,api,ai}:$GIT_SHA`

## Acceptance criteria

- [x] Push в main запускает CI
- [x] Lint и тесты блокируют push при падении
- [x] Успешный прогон публикует образы в registry
- [x] Smoke test bot image (`BOT_HEALTH_ONLY`)
- [x] Секреты не в образе и git

## Технические заметки

- **Go-only CI** — без Python/ruff
- Build context — корень monorepo (`go.mod`)
- `CI=true` включает golangci-lint в `make lint`

## Out of scope

- Deploy на prod VM (004)
