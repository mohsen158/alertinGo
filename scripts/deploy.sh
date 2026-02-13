#!/usr/bin/env bash
set -euo pipefail

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

log "Starting deploy..."

log "Pulling latest changes..."
git pull

log "Rebuilding and restarting containers..."
docker compose up --build -d

log "Deploy finished."
