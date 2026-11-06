#!/usr/bin/env bash
# Gera k8s/.env.production na VM (uma vez; segredos ficam só lá).
#   bash k8s/scripts/init-secrets.sh
#   FORCE=1 bash k8s/scripts/init-secrets.sh   # sobrescreve
set -euo pipefail

DEST="${1:-$(cd "$(dirname "$0")/.." && pwd)/.env.production}"

if [ -e "$DEST" ] && [ "${FORCE:-0}" != "1" ]; then
  echo "já existe: $DEST (use FORCE=1 para sobrescrever)" >&2
  exit 1
fi

umask 077
mkdir -p "$(dirname "$DEST")"
cat > "$DEST" <<EOF
PG_PASS=$(openssl rand -hex 24)
JWT_SECRET=$(openssl rand -hex 32)
GAME_JWT_SECRET=$(openssl rand -hex 32)
BACKEND_API_KEY=$(cat /proc/sys/kernel/random/uuid)
REDIS_PASSWORD=$(openssl rand -hex 24)
EOF
chmod 600 "$DEST"
echo "segredos gravados em $DEST"
