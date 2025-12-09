#!/bin/sh
set -eu

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"

ENV_FILE="${ROOT_DIR}/.env"
TEMPLATE="${ROOT_DIR}/monitoring/alertmanager/alertmanager.tmpl.yml"
OUTPUT="${ROOT_DIR}/monitoring/alertmanager/alertmanager.yml"

if [ ! -f "$ENV_FILE" ]; then
  echo "ERROR: .env not found at $ENV_FILE" >&2
  exit 1
fi

if [ ! -f "$TEMPLATE" ]; then
  echo "ERROR: template not found at $TEMPLATE" >&2
  exit 1
fi

set -a
. "$ENV_FILE"
set +a

: "${ALERT_TELEGRAM_BOT_TOKEN:?ALERT_TELEGRAM_BOT_TOKEN is not set in .env}"
: "${ALERT_TELEGRAM_CHAT_ID:?ALERT_TELEGRAM_CHAT_ID is not set in .env}"

if ! command -v envsubst >/dev/null 2>&1; then
  echo "ERROR: envsubst not found. Install gettext/gettext-base on host." >&2
  exit 1
fi

envsubst < "$TEMPLATE" > "$OUTPUT"

echo "Generated $OUTPUT from $TEMPLATE using $ENV_FILE"
