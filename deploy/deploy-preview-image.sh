#!/usr/bin/env bash
set -Eeuo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.preview.yml}"
ENV_FILE="${ENV_FILE:-.env.preview}"
SERVICE="${SERVICE:-sub2api}"
HEALTH_RETRIES="${HEALTH_RETRIES:-45}"
HEALTH_INTERVAL="${HEALTH_INTERVAL:-2}"

cd "$(dirname "$0")"

if [ ! -f "$ENV_FILE" ]; then
  echo "missing env file: $ENV_FILE" >&2
  exit 1
fi

if [ -z "${HEALTH_URL:-}" ]; then
  env_server_port="$(
    awk -F= '/^[[:space:]]*SERVER_PORT[[:space:]]*=/{print $2}' "$ENV_FILE" |
      tail -n 1 |
      tr -d '[:space:]' |
      tr -d '"' |
      tr -d "'"
  )"
  HEALTH_URL="http://127.0.0.1:${SERVER_PORT:-${env_server_port:-3000}}/health"
fi

compose=(docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE")

echo "--- pull image ---"
"${compose[@]}" pull "$SERVICE"

echo "--- restart service ---"
"${compose[@]}" up -d --no-deps --force-recreate "$SERVICE"

echo "--- wait health ---"
for i in $(seq 1 "$HEALTH_RETRIES"); do
  if curl -fsS "$HEALTH_URL"; then
    echo
    echo "health ok attempt=$i"
    break
  fi

  if [ "$i" -eq "$HEALTH_RETRIES" ]; then
    echo
    echo "health failed after ${HEALTH_RETRIES} attempts" >&2
    "${compose[@]}" ps "$SERVICE" >&2 || true
    exit 1
  fi

  sleep "$HEALTH_INTERVAL"
done

echo "--- service status ---"
"${compose[@]}" ps "$SERVICE"
