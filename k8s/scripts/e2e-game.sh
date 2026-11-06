#!/usr/bin/env bash
# E2E manual do jogo (não roda no CI: cria um usuário de teste no banco).
#   bash k8s/scripts/e2e-game.sh
# Percorre JWT do backend -> /newRoom -> /getTicket -> handshake WebSocket 101.
set -euo pipefail

NS="${NS:-hoodwink}"
ENV_FILE="${ENV_FILE:-$(cd "$(dirname "$0")/.." && pwd)/.env.production}"
export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"
# shellcheck disable=SC1090
. "$ENV_FILE"

PLAYER_ID="00000000-0000-0000-0000-000000000001"

cleanup() {
  kubectl -n "$NS" exec deploy/hoodwink-postgres -- \
    psql -U postgres -d hoodwink -c "DELETE FROM users WHERE id = '$PLAYER_ID';" >/dev/null 2>&1 || true
}
trap cleanup EXIT

kubectl -n "$NS" exec deploy/hoodwink-postgres -- psql -U postgres -d hoodwink -v ON_ERROR_STOP=1 -c "
INSERT INTO users (id, email, provider, username)
VALUES ('$PLAYER_ID', 'smoke@example.com', 'smoke', 'smoke')
ON CONFLICT (id) DO NOTHING;" >/dev/null

TOKEN=$(python3 - "$JWT_SECRET" <<'PY'
import base64, hashlib, hmac, json, sys, time

def b64(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).rstrip(b"=").decode()

secret = sys.argv[1].encode()
header = b64(json.dumps({"alg": "HS256", "typ": "JWT"}, separators=(",", ":")).encode())
claims = {
    "id": "00000000-0000-0000-0000-000000000001",
    "username": "smoke",
    "provider": "smoke",
    "email": "smoke@example.com",
    "temp": False,
    "exp": int(time.time()) + 600,
}
payload = b64(json.dumps(claims, separators=(",", ":")).encode())
sig = b64(hmac.new(secret, f"{header}.{payload}".encode(), hashlib.sha256).digest())
print(f"{header}.{payload}.{sig}")
PY
)

ROOM=$(curl -fsS -X POST http://127.0.0.1:4444/newRoom \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"roomName":"smoke-test","maxPlayers":2}' \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["RoomId"])')
echo "ok  newRoom -> $ROOM"

TICKET=$(curl -fsS -X POST "http://127.0.0.1:4444/getTicket/$ROOM" \
  -H "Authorization: Bearer $TOKEN" \
  | python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["ticket"])')
echo "ok  getTicket"

WS_HEAD=$(curl -s -i --max-time 3 \
  -H 'Upgrade: websocket' -H 'Connection: Upgrade' \
  -H 'Sec-WebSocket-Version: 13' -H 'Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==' \
  "http://127.0.0.1:4444/game?Ticket=$TICKET" 2>/dev/null || true)
echo "$WS_HEAD" | grep -q "101 Switching Protocols" || {
  echo "FALHOU WebSocket:" >&2
  echo "$WS_HEAD" >&2
  exit 1
}
echo "ok  WebSocket -> 101"
