#!/bin/sh

set -e  # Завершать при любой ошибке

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

log "=== Creating shared network ==="
docker network create --attachable common-net 2>/dev/null || true

log "=== Starting infrastructure (rmq, temporal) ==="
cd init
log "Building and starting infrastructure..."
if ! docker compose up --build -d; then
    log "ERROR: Failed to start infrastructure"
    exit 1
fi
cd - > /dev/null

# Ждём готовности RabbitMQ
log "Waiting for RabbitMQ to be ready..."
while ! docker inspect rmq --format='{{.State.Health.Status}}' 2>/dev/null | grep -q 'healthy'; do
    STATUS=$(docker inspect rmq --format='{{.State.Status}}' 2>/dev/null || echo "unknown")
    if [ "$STATUS" = "exited" ]; then
        log "ERROR: RabbitMQ container exited unexpectedly"
        docker logs rmq
        exit 1
    fi
    sleep 2
done
log "RabbitMQ is ready"

# Ждём готовности Temporal
log "Waiting for Temporal to be ready..."
while ! docker inspect temporal --format='{{.State.Health.Status}}' 2>/dev/null | grep -q 'healthy'; do
    STATUS=$(docker inspect temporal --format='{{.State.Status}}' 2>/dev/null || echo "unknown")
    if [ "$STATUS" = "exited" ]; then
        log "ERROR: Temporal container exited unexpectedly"
        docker logs temporal
        exit 1
    fi
    sleep 2
done
log "Temporal is ready"

# Запускаем микросервисы
SERVICES="user payment notification"

for svc in $SERVICES; do
    log "=== Starting service: $svc ==="
    if [ ! -d "$svc" ]; then
        log "ERROR: Directory '$svc' not found"
        exit 1
    fi
    cd "$svc"
    log "Building $svc..."
    brewkit build
    log "Starting $svc containers..."
    if ! docker compose up --build -d; then
        log "ERROR: Failed to start $svc"
        exit 1
    fi
    cd - > /dev/null
    log "Service $svc started successfully"
done

log "=== All services are running! ==="
log "RabbitMQ UI: http://localhost:15672"
log "Temporal UI: http://localhost:8090"
log "Prometheus: http://localhost:8082/metrics"