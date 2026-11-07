#!/usr/bin/env bash
# Atualiza as faixas da Cloudflare usadas pelo http_realip_module do nginx.
#
#   sudo bash k8s/scripts/cloudflare-realip.sh
#
# Baixa https://www.cloudflare.com/ips-v4 e /ips-v6, recusa lista vazia ou com
# linha que não seja CIDR, grava /etc/nginx/snippets/cloudflare-realip.conf e só
# recarrega o nginx depois de um `nginx -t` verde. Se o teste falhar, restaura o
# arquivo anterior. Idempotente; sem segredo; roda no edge-nginx.sh e num timer
# systemd semanal.
set -euo pipefail

SNIPPET="${SNIPPET:-/etc/nginx/snippets/cloudflare-realip.conf}"
CF_URL_V4="${CF_URL_V4:-https://www.cloudflare.com/ips-v4}"
CF_URL_V6="${CF_URL_V6:-https://www.cloudflare.com/ips-v6}"

DOWNLOADS="$(mktemp)"
OUTPUT="$(mktemp)"
BACKUP=""

cleanup() {
  rm -f "$DOWNLOADS" "$OUTPUT" "$BACKUP"
}
trap cleanup EXIT

is_ipv4_cidr() {
  local line="$1"
  [[ "$line" =~ ^([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})/([0-9]{1,2})$ ]] || return 1
  local a="${BASH_REMATCH[1]}" b="${BASH_REMATCH[2]}" c="${BASH_REMATCH[3]}" d="${BASH_REMATCH[4]}" p="${BASH_REMATCH[5]}"
  (( a <= 255 && b <= 255 && c <= 255 && d <= 255 && p <= 32 )) || return 1
}

is_ipv6_cidr() {
  local line="$1"
  [[ "$line" =~ ^[0-9a-fA-F:]+/[0-9]{1,3}$ ]] || return 1
  local p="${line##*/}"
  (( p <= 128 )) || return 1
}

# 1) Baixa as duas listas. Falha aqui aborta sem tocar em arquivo algum.
curl -fsSL "$CF_URL_V4" > "$DOWNLOADS" || { echo "Falha ao baixar $CF_URL_V4" >&2; exit 1; }
curl -fsSL "$CF_URL_V6" >> "$DOWNLOADS" || { echo "Falha ao baixar $CF_URL_V6" >&2; exit 1; }

if [ ! -s "$DOWNLOADS" ]; then
  echo "Lista da Cloudflare veio vazia; abortando." >&2
  exit 1
fi

# 2) Valida cada linha como CIDR e monta o include.
while IFS= read -r line; do
  [ -n "$line" ] || continue
  if ! is_ipv4_cidr "$line" && ! is_ipv6_cidr "$line"; then
    echo "Linha inválida (não é CIDR): $line" >&2
    exit 1
  fi
  echo "set_real_ip_from $line;"
done < "$DOWNLOADS" > "$OUTPUT"

# 3) Grava mantendo o anterior para rollback, testa e só então recarrega.
install -d -m 755 "$(dirname "$SNIPPET")"
if [ -f "$SNIPPET" ]; then
  BACKUP="$(mktemp)"
  cp "$SNIPPET" "$BACKUP"
fi

install -m 644 "$OUTPUT" "$SNIPPET"
if nginx -t >/dev/null 2>&1; then
  systemctl reload nginx
  echo "Faixas da Cloudflare atualizadas em $SNIPPET."
else
  if [ -n "$BACKUP" ]; then
    cp "$BACKUP" "$SNIPPET"
  fi
  echo "nginx -t falhou; arquivo anterior restaurado." >&2
  exit 1
fi
