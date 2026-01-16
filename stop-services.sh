#!/bin/sh

set -e

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

log "=== Stopping all services ==="

SERVICES="user order payment notification product"

# Останавливаем микросервисы
for svc in $SERVICES; do
    if [ -d "$svc" ]; then
        log "Stopping $svc"
        cd "$svc" && docker compose stop && cd - > /dev/null
    fi
done

# Останавливаем инфраструктуру
if [ -d "init" ]; then
    log "Stopping infrastructure"
    cd init && docker compose stop && cd - > /dev/null
fi

log "=== All containers stopped ==="