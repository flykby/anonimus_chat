# Production deploy on VM

Milestone 1: echo-бот из internal registry через Docker Compose.

## Prerequisites

- Linux VM with Docker Engine and Docker Compose v2
- Access to internal registry (from CI task 003)
- `BOT_TOKEN` from [@BotFather](https://t.me/BotFather)
- Git clone of this repo on VM (recommended path: `/opt/anonimus_chat`)

## First-time setup

```bash
git clone https://github.com/flykby/anonimus_chat.git /opt/anonimus_chat
cd /opt/anonimus_chat

cp .env.prod.example .env
# Edit .env: BOT_TOKEN, REGISTRY_URL, REGISTRY_USER, REGISTRY_PASSWORD, IMAGE_TAG

bash scripts/deploy.sh --tag latest
```

Verify:

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml ps
docker compose -f docker-compose.yml -f docker-compose.prod.yml exec bot wget -qO- http://127.0.0.1:8080/health
docker compose -f docker-compose.yml -f docker-compose.prod.yml exec api wget -qO- http://127.0.0.1:8000/health
docker compose -f docker-compose.yml -f docker-compose.prod.yml exec ai wget -qO- http://127.0.0.1:8001/health
```

## Deploy new version

After CI pushes a new image tag:

```bash
cd /opt/anonimus_chat
git pull   # updates compose/scripts if needed

bash scripts/deploy.sh --tag <git-sha-short>
# or rely on IMAGE_TAG in .env:
bash scripts/deploy.sh
```

Flow: **registry login → docker pull → compose up → health check**.

## Deploy via GitHub Actions

Push to `main` runs CI: lint/test → docker build → push images → **SSH deploy on VM** (when secrets are configured).

```mermaid
flowchart LR
    Push[git push main] --> CI[GitHub Actions]
    CI --> Reg[(Registry)]
    CI --> SSH[SSH to VM]
    SSH --> Deploy[remote-deploy.sh]
    Reg --> Deploy
    Deploy --> Stack[docker compose up]
```

### Option A — GitHub-hosted CI + SSH deploy (recommended)

GitHub Actions builds and pushes images to a registry the VM can reach (**not** `127.0.0.1:5000` on the VM — GitHub runners cannot push there). Use **GHCR** or any registry with a reachable endpoint.

#### 1. Registry (example: GHCR)

In GitHub → **Settings → Secrets and variables → Actions**, add:

| Secret | Example |
|--------|---------|
| `REGISTRY_URL` | `ghcr.io/flykby/anonimus` |
| `REGISTRY_USER` | your GitHub username |
| `REGISTRY_PASSWORD` | GitHub PAT with `write:packages`, or `GITHUB_TOKEN` if packages are in the same org/repo |

On the VM, point `.env` at the same registry and log in once:

```bash
# /opt/anonimus_chat/.env
REGISTRY_URL=ghcr.io/flykby/anonimus
IMAGE_TAG=latest   # deploy.sh uses --tag from CI; this is fallback only

echo "$GITHUB_PAT" | docker login ghcr.io -u YOUR_GITHUB_USER --password-stdin
```

Make GHCR packages pullable from the VM (package visibility / org settings).

Image tags pushed by CI: full SHA, short SHA (e.g. `2889586`), and `latest` on `main`.

#### 2. SSH deploy secrets

Generate a deploy key (on your laptop or VM):

```bash
ssh-keygen -t ed25519 -f deploy_key -N ""
# Add deploy_key.pub to VM: ~/.ssh/authorized_keys (for deploy user)
```

Add repository secrets:

| Secret | Value |
|--------|-------|
| `DEPLOY_HOST` | VM IP or hostname |
| `DEPLOY_USER` | SSH user (e.g. `root` or dedicated `deploy`) |
| `DEPLOY_SSH_KEY` | private key contents (`deploy_key`) |

#### 3. Flow after setup

```bash
git push origin main
```

CI pushes `ghcr.io/.../bot:2889586` (short SHA), then SSH runs:

```bash
cd /opt/anonimus_chat
bash scripts/remote-deploy.sh 2889586
```

Manual redeploy on VM (same as CI):

```bash
bash scripts/remote-deploy.sh $(git rev-parse --short HEAD)
```

Optional: create GitHub **Environment** `production` and add `environment: production` to the deploy job in `.github/workflows/ci.yml` for manual approval before deploy.

### Option B — Self-hosted runner (keep local registry `127.0.0.1:5000`)

If you want CI to build on the VM and push to `127.0.0.1:5000`:

1. Install a [GitHub Actions self-hosted runner](https://docs.github.com/en/actions/hosting-your-own-runners) on the VM.
2. Set repository **variable** `DEPLOY_SELF_HOSTED=true`.
3. Keep secrets `REGISTRY_URL=127.0.0.1:5000/anonimus` (and login if needed).

Workflow: `.github/workflows/deploy-self-hosted.yml` — lint, test, build, push, deploy on the same machine.

Leave `DEPLOY_HOST` unset so the SSH deploy job in `ci.yml` is skipped.

## Rollback

Rollback to the previous successful tag (< 2 min):

```bash
bash scripts/deploy.sh --rollback
# or explicit tag:
bash scripts/deploy.sh --tag <previous-sha>
```

Previous tag is stored in `.deploy/previous` after each successful deploy.

## HTTPS reverse proxy (stub for task 009)

Optional Caddy profile for TLS termination:

```bash
# Set DOMAIN=bot.example.com in .env
bash scripts/deploy.sh --with-proxy
```

Caddy config: `deploy/caddy/Caddyfile`. Full webhook routing — task 009.

## Systemd (optional)

```bash
sudo cp deploy/systemd/anonimus.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now anonimus
```

Adjust `WorkingDirectory` in the unit file if not using `/opt/anonimus_chat`.

## Secrets

- All secrets live in `.env` on the VM only
- Never commit `.env` or put tokens in compose files
- `.deploy/` stores only image tags, no secrets

## Troubleshooting

| Problem | Check |
|---------|-------|
| `pull access denied` | `REGISTRY_USER` / `REGISTRY_PASSWORD`, `docker login` |
| `address already in use` | Prod does not bind app ports on the host (only optional Caddy `:80/:443`). Stale Docker endpoints often survive `ss`. Run `docker compose -f docker-compose.yml -f docker-compose.prod.yml down --remove-orphans`, `docker rm -f anonimus-postgres anonimus-redis anonimus-api anonimus-ai anonimus-bot`, `git pull`, redeploy. `deploy.sh` also removes containers stuck in `created` state. If it persists: `systemctl restart docker` |
| Bot unhealthy | `docker logs anonimus-bot`, verify `BOT_TOKEN` |
| No Telegram reply | Bot uses long polling; VM needs outbound HTTPS to `api.telegram.org` |
| Rollback missing | `.deploy/previous` exists only after ≥1 successful deploy |

## Database migrations (prod)

Postgres is not exposed on the host in prod. Run goose via the Docker network:

```bash
docker run --rm --network anonimus-prod_default \
  -v /opt/anonimus_chat/migrations:/migrations \
  ghcr.io/pressly/goose:latest \
  -dir /migrations postgres \
  "postgresql://anonimus:YOUR_PASSWORD@postgres:5432/anonimus?sslmode=disable" up
```

## Next steps

- **006** — database schema and goose migrations
- **009** — webhook + HTTPS via Caddy
