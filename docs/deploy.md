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
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8000/health
curl http://127.0.0.1:8001/health
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
| Bot unhealthy | `docker logs anonimus-bot`, verify `BOT_TOKEN` |
| No Telegram reply | Bot uses long polling; VM needs outbound HTTPS to `api.telegram.org` |
| Rollback missing | `.deploy/previous` exists only after ≥1 successful deploy |

## Next steps

- **006** — database schema and Alembic migrations
- **009** — webhook + HTTPS via Caddy
