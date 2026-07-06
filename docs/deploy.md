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

GitHub Actions builds and pushes images to **GHCR**, then deploys on the VM over SSH.

#### Checklist

**A. GitHub repository secrets** (`Settings → Secrets and variables → Actions`):

| Secret | Value | Required |
|--------|-------|----------|
| `REGISTRY_URL` | `ghcr.io/flykby/anonimus` | yes |
| `REGISTRY_USER` | leave empty (CI uses `GITHUB_TOKEN`) | no |
| `REGISTRY_PASSWORD` | leave empty for GHCR push from CI | no |
| `DEPLOY_HOST` | VM public IP or hostname | yes |
| `DEPLOY_USER` | SSH user, e.g. `root` | yes |
| `DEPLOY_SSH_KEY` | private SSH key for deploy | yes |

CI push to GHCR uses the built-in `GITHUB_TOKEN` (`packages: write` is set in the workflow).  
Only `REGISTRY_URL` is required for the registry side in GitHub secrets.

**B. SSH key for deploy**

On your laptop:

```bash
ssh-keygen -t ed25519 -f deploy_key -N ""
ssh-copy-id -i deploy_key.pub root@YOUR_VM_IP   # or append deploy_key.pub to authorized_keys
```

Copy private key to GitHub secret `DEPLOY_SSH_KEY`:

```bash
# Recommended: set secret from file (preserves newlines)
gh secret set DEPLOY_SSH_KEY < deploy_key

# Or copy manually — must be the PRIVATE key (not .pub):
cat deploy_key
```

The secret must include the full file:

```
-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----
```

Do not wrap in quotes. If deploy fails with "not a valid private key", delete the secret and run `gh secret set DEPLOY_SSH_KEY < deploy_key` again.

**C. VM one-time setup**

```bash
cd /opt/anonimus_chat
git pull

# Interactive: docker login ghcr.io + update .env
GHCR_PAT=ghp_xxxx bash scripts/setup-vm-ghcr.sh
# or: read -rsp prompt inside the script
bash scripts/setup-vm-ghcr.sh
```

The VM needs a PAT with **`read:packages`** to pull private images.  
After the first CI push, make packages visible: GitHub → Packages → each package → **Change visibility** (public is simplest for a solo project).

**D. First automated deploy**

```bash
git push origin main
```

Watch **Actions** tab: `Push to registry` → `Deploy to VM`.

Image tags: full SHA, short SHA (e.g. `a4f2b6a`), and `latest` on `main`.

#### Manual redeploy on VM

```bash
bash scripts/remote-deploy.sh $(git rev-parse --short HEAD)
# or:
bash scripts/deploy.sh --tag latest
```

## Rollback

Rollback to the previous successful tag (< 2 min). Also rolls back DB schema one step (`goose down-to`):

```bash
bash scripts/deploy.sh --rollback
# or explicit tag:
bash scripts/deploy.sh --tag <previous-sha>
```

Previous tag is stored in `.deploy/previous` after each successful deploy.  
Previous schema version — in `.deploy/migration_previous`.

## Webhook mode (optional)

By default, the bot uses **long polling**. For production with payments, **webhook mode** is recommended — Telegram will retry failed deliveries.

### Option 1: Self-signed certificate with IP address

Use your **public** IP (what Telegram can reach from the internet), not the internal VM subnet:

```bash
curl -4 ifconfig.me   # e.g. 109.122.197.235
./scripts/gen-webhook-cert.sh YOUR_PUBLIC_IP ./certs
```

Add to `.env`:

```env
WEBHOOK_URL=https://YOUR_PUBLIC_IP:8443/telegram/webhook
WEBHOOK_SECRET=your-random-secret-here
WEBHOOK_CERT_PATH=/app/certs/webhook.pem
WEBHOOK_KEY_PATH=/app/certs/webhook.key
```

Prod compose maps host `8443` → container `8080` and mounts `./certs` read-only.

If the bot was already running with a bad key permission, fix without regenerating:

```bash
chmod 644 certs/webhook.pem certs/webhook.key
docker compose -f docker-compose.yml -f docker-compose.prod.yml restart bot
```

### Option 2: Domain with reverse proxy

With a domain and Caddy/nginx for TLS termination:

```env
WEBHOOK_URL=https://bot.example.com/telegram/webhook
WEBHOOK_SECRET=your-random-secret-here
# No cert paths needed — proxy handles TLS
```

### Switching back to polling

Remove or clear `WEBHOOK_URL` from `.env` — the bot will use long polling.

### Telegram webhook ports

Telegram only accepts webhooks on ports: **443, 80, 88, 8443**

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
| Webhook: `permission denied` on `webhook.key` | Bot runs as uid 65534; key must be world-readable: `chmod 644 certs/webhook.key` then restart bot |
| Webhook: bot silent, health OK | Check `WEBHOOK_URL` uses **public** IP (`curl -4 ifconfig.me`), cert CN must match; open port 8443 in provider firewall |
| Webhook: `too many requests` on register | Telegram rate-limits `setWebhook`; wait 1–2 min, fix cert/permissions, then restart once |
| No Telegram reply | Long polling: VM needs outbound HTTPS to `api.telegram.org`. Webhook: see rows above |
| Rollback missing | `.deploy/previous` exists only after ≥1 successful deploy |

## Database migrations (prod)

Migrations run **automatically** during `scripts/deploy.sh` (and CI deploy):

1. Start postgres + redis
2. `goose up` (or `goose down-to` on `--rollback`)
3. Start/update app containers
4. On deploy failure after migrate — auto rollback schema to pre-deploy version

State files in `.deploy/`:

| File | Purpose |
|------|---------|
| `migration_current` | Applied goose version after last successful deploy |
| `migration_previous` | Version to restore on `--rollback` (one step, like image tags) |

Manual commands (rarely needed):

```bash
bash scripts/migrate-prod.sh status
bash scripts/migrate-prod.sh version
bash scripts/deploy.sh --skip-migrate --tag latest   # emergency: deploy without migrate
```

## Next steps

- **006** — database schema and goose migrations
- **009** — webhook + HTTPS via Caddy
