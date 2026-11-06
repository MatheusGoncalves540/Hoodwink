#!/usr/bin/env bash
# Borda do Hoodwink na VM: nginx do host -> pods (hostNetwork/NodePort).
#
#   sudo bash k8s/scripts/edge-nginx.sh          # grava a config HTTP e recarrega
#   sudo bash k8s/scripts/edge-nginx.sh --tls    # + emite/instala o certificado
#
# O TLS é terminado aqui (como no Summus), não no Traefik. Rode --tls somente
# depois que o registro DNS apontar para esta VM.
set -euo pipefail

DOMAIN="${DOMAIN:-hoodwink.hermis.com.br}"
WEB_PORT="${WEB_PORT:-30081}"
API_PORT="${API_PORT:-8088}"
GAME_PORT="${GAME_PORT:-4444}"
CONF="/etc/nginx/sites-available/$DOMAIN"
ENABLED="/etc/nginx/sites-enabled/$DOMAIN"

install -d -m 755 /var/www/certbot

if [ "${FORCE_CONFIG:-0}" = "1" ] || [ ! -f "$CONF" ] || ! grep -q "ssl_certificate" "$CONF"; then
  cat > "$CONF" <<'EOF'
server {
    listen 80;
    listen [::]:80;
    server_name @DOMAIN@;

    client_max_body_size 25M;

    location ^~ /.well-known/acme-challenge/ {
        root /var/www/certbot;
        default_type "text/plain";
        try_files $uri =404;
    }

    # API Go (o prefixo /api/ é removido pelo proxy_pass com barra final).
    location /api/ {
        proxy_pass http://127.0.0.1:@API_PORT@/;

        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
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
}
EOF
  sed -i \
    -e "s|@DOMAIN@|$DOMAIN|g" \
    -e "s|@API_PORT@|$API_PORT|g" \
    -e "s|@GAME_PORT@|$GAME_PORT|g" \
    -e "s|@WEB_PORT@|$WEB_PORT|g" \
    "$CONF"
fi

ln -sf "$CONF" "$ENABLED"
nginx -t
systemctl reload nginx
echo "==> nginx servindo $DOMAIN (HTTP)"

if [ "${1:-}" = "--tls" ]; then
  echo "==> Emitindo/instalando certificado para $DOMAIN"
  certbot --nginx -d "$DOMAIN" --non-interactive --agree-tos --redirect --keep-until-expiring
  systemctl reload nginx
  echo "==> TLS pronto: https://$DOMAIN"
fi
