#!/usr/bin/env bash
# Smoke test pós-deploy (rodado na VM pelo CI).
# Testa os três serviços, o schema e a API key semeada; não cria dados.
set -euo pipefail

NS="${NS:-hoodwink}"
BASE="${BASE:-127.0.0.1}"
export KUBECONFIG="${KUBECONFIG:-$HOME/.kube/config}"

check() {
  local url="$1"
  local code
  code=$(curl -s -o /dev/null --max-time 10 -w '%{http_code}' "$url")
  [ "$code" = "200" ] || { echo "FALHOU $url -> $code" >&2; exit 1; }
  echo "ok  $url -> $code"
}

check "http://$BASE:8088/alive"
check "http://$BASE:4444/alive"
check "http://$BASE:30081/"

keys=$(kubectl -n "$NS" exec deploy/hoodwink-postgres -- \
  psql -U postgres -d hoodwink -tAc "SELECT count(*) FROM api_keys WHERE system_name = 'gameServer'")
[ "$keys" = "1" ] || { echo "FALHOU api_keys seed -> $keys" >&2; exit 1; }
echo "ok  schema/API key"
