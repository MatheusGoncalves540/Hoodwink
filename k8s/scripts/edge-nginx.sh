#!/usr/bin/env bash
# Borda do Hoodwink na VM: nginx do host -> pods (hostNetwork/NodePort).
#
#   sudo bash k8s/scripts/edge-nginx.sh          # HTTP + HTTPS se o cert já existir
#   sudo bash k8s/scripts/edge-nginx.sh --tls    # emite/renova o certificado e instala
#
# O TLS é terminado aqui (como no Summus), não no Traefik. Com o proxy da
# Cloudflare ligado, o desafio HTTP-01 chega no origin por HTTPS; por isso o
# script sobe um server 443 temporário (cert do api.hermis.com.br ou self-signed)
# antes de rodar o certbot. Depois troca para o certificado do domínio.
set -euo pipefail

DOMAIN="${DOMAIN:-hoodwink.hermis.com.br}"
WEB_PORT="${WEB_PORT:-30081}"
API_PORT="${API_PORT:-8088}"
GAME_PORT="${GAME_PORT:-4444}"
CONF="/etc/nginx/sites-available/$DOMAIN"
ENABLED="/etc/nginx/sites-enabled/$DOMAIN"
SNIPPET="/etc/nginx/snippets/$DOMAIN-proxy.conf"
LE_DIR="/etc/letsencrypt/live/$DOMAIN"

# Real IP: só funciona se o nginx tiver http_realip_module compilado.
if ! nginx -V 2>&1 | grep -q http_realip_module; then
  echo "nginx sem http_realip_module; recompile com o módulo e rode de novo." >&2
  exit 1
fi

# Atualiza as faixas da Cloudflare (idempotente; roda também num timer semanal).
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
bash "$SCRIPT_DIR/cloudflare-realip.sh"

install -d -m 755 /var/www/certbot /etc/nginx/snippets

cat > "$SNIPPET" <<'EOF'
# Faixas da Cloudflare e header que o realip substitui. A troca só acontece
# quando a conexão vem de uma dessas faixas; um CF-Connecting-IP forjado de
# fora delas é ignorado.
real_ip_header CF-Connecting-IP;
include /etc/nginx/snippets/cloudflare-realip.conf;

location ^~ /.well-known/acme-challenge/ {
    root /var/www/certbot;
    default_type "text/plain";
    try_files $uri =404;
}

# Rotas internas do backend nunca saem pela borda.
location ^~ /api/internal/ {
    return 404;
}

# API Go (o prefixo /api/ é removido pelo proxy_pass com barra final).
location /api/ {
    proxy_pass http://127.0.0.1:@API_PORT@/;

    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header CF-Connecting-IP "";
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}

# gameServer: HTTP (/newRoom, /getTicket) + WebSocket por /gs/game.
location /gs/ {
    proxy_pass http://127.0.0.1:@GAME_PORT@/;

    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header CF-Connecting-IP "";
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 3600s;
    proxy_send_timeout 3600s;
}

# SPA React (Service NodePort -> nginx da imagem web).
location / {
    proxy_pass http://127.0.0.1:@WEB_PORT@;

    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
EOF
sed -i \
  -e "s|@API_PORT@|$API_PORT|g" \
  -e "s|@GAME_PORT@|$GAME_PORT|g" \
  -e "s|@WEB_PORT@|$WEB_PORT|g" \
  "$SNIPPET"

write_http() {
  cat > "$CONF" <<EOF
server {
    listen 80;
    listen [::]:80;
    server_name $DOMAIN;
    client_max_body_size 25M;
    include $SNIPPET;
}
EOF
}

write_https() { # $1 = fullchain; $2 = privkey
  cat >> "$CONF" <<EOF

server {
    listen 443 ssl;
    listen [::]:443 ssl;
    server_name $DOMAIN;
    client_max_body_size 25M;

    ssl_certificate $1;
    ssl_certificate_key $2;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    include $SNIPPET;
}
EOF
}

bootstrap_cert() {
  # Enquanto não há cert do domínio, usa o do api (Cloudflare Full ignora a
  # validação) ou um self-signed, só para o desafio ACME chegar no origin.
  if [ -f /etc/letsencrypt/live/api.hermis.com.br/fullchain.pem ]; then
    BOOT_CERT=/etc/letsencrypt/live/api.hermis.com.br/fullchain.pem
    BOOT_KEY=/etc/letsencrypt/live/api.hermis.com.br/privkey.pem
    return
  fi
  install -d -m 755 /etc/nginx/ssl
  BOOT_CERT="/etc/nginx/ssl/$DOMAIN.pem"
  BOOT_KEY="/etc/nginx/ssl/$DOMAIN.key"
  if [ ! -f "$BOOT_CERT" ] || [ ! -f "$BOOT_KEY" ]; then
    openssl req -x509 -nodes -newkey rsa:2048 -days 2 \
      -subj "/CN=$DOMAIN" -keyout "$BOOT_KEY" -out "$BOOT_CERT" >/dev/null 2>&1
  fi
}

write_http
if [ -f "$LE_DIR/fullchain.pem" ]; then
  write_https "$LE_DIR/fullchain.pem" "$LE_DIR/privkey.pem"
elif [ "${1:-}" = "--tls" ]; then
  bootstrap_cert
  write_https "$BOOT_CERT" "$BOOT_KEY"
fi

ln -sf "$CONF" "$ENABLED"
nginx -t
systemctl reload nginx
echo "==> nginx servindo $DOMAIN (HTTP)"

if [ "${1:-}" = "--tls" ]; then
  echo "==> Emitindo/renovando certificado para $DOMAIN"
  certbot certonly --webroot -w /var/www/certbot -d "$DOMAIN" \
    --non-interactive --agree-tos --keep-until-expiring
  write_http
  write_https "$LE_DIR/fullchain.pem" "$LE_DIR/privkey.pem"
  nginx -t
  systemctl reload nginx
  echo "==> TLS pronto: https://$DOMAIN"
fi
