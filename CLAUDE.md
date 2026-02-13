# AlertinGo Project

## Overview
Health check monitoring & alerting system in Go. Servers send heartbeats; if they stop, Telegram alerts fire.

## Tech Stack
- **Go 1.25** (Alpine Docker), **Gin** web framework, **pgx/v5** for Postgres, **godotenv**
- **Postgres 16** (Docker), **Air** for hot reload
- Telegram Bot API for notifications

## Architecture
Flat structure, no internal/repository/service split. Handlers talk to `db` package directly.

```
cmd/main.go           → entry point, routes
cmd/admin/main.go     → admin CLI (generate API key)
handler/heartbeat.go  → POST /heartbeat (upsert monitor)
handler/monitor.go    → GET/PUT/DELETE monitors
handler/channel.go           → GET/POST/DELETE channels
handler/api_key.go           → GET/POST/DELETE api-keys
handler/notification_log.go  → GET /notification-logs
middleware/auth.go           → API key auth middleware (X-API-Key header, SHA-256 validated)
model/models.go              → Monitor, NotificationChannel, AlertState, NotificationLog, ApiKey structs
db/db.go              → pgxpool connection, all SQL queries, migrations
watcher/watcher.go    → background goroutine (10s tick): checks overdue + recovered
notifier/telegram.go  → sends messages via Telegram Bot API
migrations/001_initial.sql   → schema (monitors, notification_channels, alert_states, notification_logs)
migrations/002_api_keys.sql  → api_keys table
```

## Key Concepts
- **Monitor**: unique by (monitor_name, check_type). Auto-created on first heartbeat as inactive.
- **NotificationChannel**: has Telegram chat ID. Assigned to monitors by admin.
- **AlertState**: tracks firing/resolved state, last_alerted_at for re-alerting.
- **ApiKey**: SHA-256 hashed key for authenticating heartbeat requests. Key prefix stored for identification. Admin-only: created via `cmd/admin` CLI.
- **NotificationLog**: every sent notification (alert, re_alert, recovered) is logged with success/error status.
- Watcher checks every 10s: overdue active monitors with channels → alert/re-alert. Recovered monitors → resolve alert + notify.

## DB Access
`postgres://alerting:alerting@localhost:5432/alerting`

## API Routes (all under /api/v1)
- GET /health
- POST /heartbeat (requires X-API-Key header)
- GET /api-keys
- DELETE /api-keys/:id
- GET /monitors
- GET /monitors/:id
- PUT /monitors/:id
- DELETE /monitors/:id
- GET /channels
- POST /channels
- DELETE /channels/:id
- GET /notification-logs

## Docker
`docker compose up --build` — app + postgres. Port 8080 (app), 5432 (db).

## Admin CLI
```
go run cmd/admin/main.go --name "my-server"
```
Generates an API key (printed once). The monitor is auto-created on first heartbeat.

## Env Vars
DATABASE_URL, TELEGRAM_BOT_TOKEN, PORT (default 8080)
