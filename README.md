# AlertinGo

Health check monitoring and alerting system built with Go. Monitored servers send heartbeats, and if they stop arriving within a configured timeout, alerts are sent via Telegram.

## How It Works

1. Server sends `POST /api/v1/heartbeat` with monitor name, check type, timeout, etc.
2. If the `(monitor_name, check_type)` pair is new, a monitor is auto-created (inactive, no channel).
3. If it already exists, `last_seen_at` and other fields are updated.
4. A background goroutine checks every 10s: if an **active** monitor with a channel hasn't reported within its timeout, a Telegram alert is sent.
5. If still down after `re_alert_interval`, a re-alert is sent.
6. When heartbeats resume, a recovery notification is sent.
7. Admin generates API keys via `cmd/admin` CLI, then activates monitors and assigns notification channels via API.

## Quick Start

```bash
cp .env.example .env
# Edit .env to set your TELEGRAM_BOT_TOKEN
docker compose up --build
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/health` | Health check |
| POST | `/api/v1/heartbeat` | Receive heartbeat (requires `X-API-Key` header) |
| GET | `/api/v1/api-keys` | List API keys |
| DELETE | `/api/v1/api-keys/:id` | Delete API key |
| GET | `/api/v1/monitors` | List all monitors |
| GET | `/api/v1/monitors/:id` | Get one monitor |
| PUT | `/api/v1/monitors/:id` | Activate + assign channel |
| DELETE | `/api/v1/monitors/:id` | Delete monitor |
| GET | `/api/v1/channels` | List channels |
| POST | `/api/v1/channels` | Create channel |
| DELETE | `/api/v1/channels/:id` | Delete channel |
| GET | `/api/v1/notification-logs` | View notification log (last 100) |

## Usage Example

**1. Generate an API key (admin CLI):**

```bash
go run cmd/admin/main.go --name "payment-service"
```

Save the printed API key — it won't be shown again.

**2. Send a heartbeat:**

```bash
curl -X POST http://localhost:8080/api/v1/heartbeat \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: <your-api-key>' \
  -d '{
    "monitor_name": "payment-service",
    "check_type": "cpu",
    "message": "CPU at 45%",
    "metadata": {"cpu_percent": 45.2},
    "timeout": 60,
    "re_alert_interval": 300
  }'
```

**3. Create a notification channel:**

```bash
curl -X POST http://localhost:8080/api/v1/channels \
  -H 'Content-Type: application/json' \
  -d '{"name": "CPU Alerts", "telegram_chat_id": "-100123456"}'
```

**4. Activate the monitor and assign a channel:**

```bash
curl -X PUT http://localhost:8080/api/v1/monitors/<monitor-id> \
  -H 'Content-Type: application/json' \
  -d '{"is_active": true, "channel_id": "<channel-id>"}'
```

**5. Stop sending heartbeats** → Telegram alert fires after timeout.

**6. Resume heartbeats** → Recovery notification is sent.

## Database Access

Connect from your host machine with any Postgres client:

```
postgres://alerting:alerting@localhost:5432/alerting
```

## Project Structure

```
alertinGo/
├── cmd/
│   ├── main.go              # Entry point
│   ├── admin/main.go        # Admin CLI (generate API key)
│   └── webhook/main.go      # Deploy poller (polls GitHub for new commits)
├── handler/
│   ├── heartbeat.go         # POST /heartbeat
│   ├── monitor.go           # Monitor CRUD
│   ├── channel.go           # Channel CRUD
│   ├── api_key.go           # API key management
│   └── notification_log.go  # Notification logs
├── middleware/auth.go       # API key auth middleware
├── model/models.go          # Data models
├── db/db.go                 # DB connection + queries
├── watcher/watcher.go       # Background timeout checker
├── notifier/telegram.go     # Telegram notifications
├── migrations/
│   ├── 001_initial.sql
│   └── 002_api_keys.sql
├── scripts/
│   └── deploy.sh            # Auto-deploy script
├── docker-compose.yml
├── Dockerfile
├── .env.example
└── .air.toml                # Hot reload config
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | Postgres connection string | — |
| `TELEGRAM_BOT_TOKEN` | Telegram Bot API token | — |
| `PORT` | HTTP server port | `8080` |
| `DEPLOY_DIR` | Project directory on server (for deploy poller) | Working directory |
| `DEPLOY_BRANCH` | Git branch to track | `main` |
| `POLL_INTERVAL` | Seconds between checks for new commits | `30` |

## Auto-Deploy (Git Polling)

The deploy poller runs **outside** Docker Compose (so it can restart the stack without killing itself). It polls GitHub every 30s for new commits — no inbound network access required (works behind NAT).

```bash
# Build the deploy poller
go build -o deploy-poller ./cmd/webhook/

# Run it (or configure as a systemd service)
DEPLOY_DIR=/path/to/alertinGo ./deploy-poller
```

On each cycle it runs `git fetch`, compares local HEAD with `origin/<branch>`, and if there are new commits, runs `scripts/deploy.sh` (git pull + docker compose rebuild).
