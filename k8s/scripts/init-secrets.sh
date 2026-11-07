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

# Par Ed25519 do ticket: privada PKCS8 e pública PKIX, base64 do PEM em uma linha.
TICKET_KEY_PEM="$(mktemp)"
# Par Ed25519 do replay: o backend assina e o bundle web confere a pública.
REPLAY_KEY_PEM="$(mktemp)"
trap 'rm -f "$TICKET_KEY_PEM" "$REPLAY_KEY_PEM"' EXIT
openssl genpkey -algorithm ed25519 -out "$TICKET_KEY_PEM" 2>/dev/null
TICKET_PRIVATE_KEY="$(base64 -w0 "$TICKET_KEY_PEM")"
TICKET_PUBLIC_KEY="$(openssl pkey -in "$TICKET_KEY_PEM" -pubout 2>/dev/null | base64 -w0)"

openssl genpkey -algorithm ed25519 -out "$REPLAY_KEY_PEM" 2>/dev/null
REPLAY_SIGNING_KEY="$(base64 -w0 "$REPLAY_KEY_PEM")"
# 32 bytes finais do SPKI DER = chave pública crua; é o que o WASM verifica.
REPLAY_PUBLIC_KEY_HEX="$(openssl pkey -in "$REPLAY_KEY_PEM" -pubout -outform DER 2>/dev/null | tail -c 32 | od -An -tx1 | tr -d ' \n')"

cat > "$DEST" <<EOF
PG_PASS=$(openssl rand -hex 24)
JWT_SECRET=$(openssl rand -hex 32)
GAME_JWT_SECRET=$(openssl rand -hex 32)
BACKEND_API_KEY=$(cat /proc/sys/kernel/random/uuid)
REDIS_PASSWORD=$(openssl rand -hex 24)
TICKET_PRIVATE_KEY=$TICKET_PRIVATE_KEY
TICKET_PUBLIC_KEY=$TICKET_PUBLIC_KEY
REPLAY_SIGNING_KEY=$REPLAY_SIGNING_KEY
REPLAY_PUBLIC_KEY_HEX=$REPLAY_PUBLIC_KEY_HEX
EOF
chmod 600 "$DEST"
echo "segredos gravados em $DEST"
