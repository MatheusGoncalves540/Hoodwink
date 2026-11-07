# Deploy — Hoodwink (k3s + Helm na VM Oracle)

## Arquitetura

```
Internet → nginx :443 (TLS certbot) → 127.0.0.1
                                        ├─ :30081  web (SPA React, NodePort)
                                        ├─ :8088   backend (API Go)
                                        └─ :4444   gameServer (HTTP + WebSocket)
                                                     ├─ Postgres :5432 (pod)
                                                     └─ Redis    :6379 (pod)
```

Tudo roda no namespace `hoodwink` do k3s da mesma VM do Summus
(`144.33.18.156`). O TLS termina no nginx do host, não no Traefik.

## Decisões críticas (não reverter)

- **`hostNetwork: true`** em backend, gameServer, Postgres e Redis — a rede de
  pods do flannel não tem egress nesta VM Oracle. Com `dnsPolicy: Default` o pod
  usa o DNS do nó; o backend alcança `oauth2.googleapis.com` para validar o
  id_token do Google e fala com Postgres/Redis por `127.0.0.1`.
  Implica `replicaCount: 1` e `strategy: Recreate`.
- **Web não usa hostNetwork**: é só arquivo estático. O Service é `NodePort`
  `30081` e o nginx do host faz proxy para `127.0.0.1:30081`.
- **Mesmo domínio, três prefixos**: `/api/` → API, `/gs/` → gameServer (com
  upgrade de WebSocket), `/` → SPA. As URLs públicas ficam em
  `client/web/.env.production` e são gravadas no bundle pelo `vite build`.
- **Imagem buildada na VM** (arm64 nativo) + `k3s ctr images import`. Sem
  GHCR/PAT.
- **Segredos fora do Git**: `k8s/.env.production` na VM (ignorado pelo Git) vira
  o Secret `hoodwink-secrets`.
- **Schema**: `AUTOMIGRATE=false` em produção; um initContainer do backend
  (imagem `postgres:16-alpine`) aplica `001_init.sql` de forma idempotente e
  semeia a API key do gameServer.

## DNS (Cloudflare)

Registro necessário em `hermis.com.br`:

| Tipo | Nome     | Conteúdo        | Proxy    |
| ---- | -------- | --------------- | -------- |
| A    | `hoodwink` | `144.33.18.156` | **DNS only** (nuvem cinza) |

Painel: <https://dash.cloudflare.com/?to=/:account/hermis.com.br/dns/records>

Use **DNS only** na primeira emissão ou deixe o proxy laranja ligado: o
`edge-nginx.sh --tls` sobe um 443 temporário para o desafio ACME chegar no
origin (a Cloudflare em Full ignora o cert temporário) e emite o certificado
válido. Depois da emissão, configure **SSL/TLS → Full (strict)** na Cloudflare;
o origin tem certificado Let's Encrypt próprio para `hoodwink.hermis.com.br`.

## Google OAuth

O login usa Google Identity Services com o client ID
`58863270996-...apps.googleusercontent.com`. Em **Google Cloud Console → APIs &
Services → Credentials → OAuth 2.0 Client**, adicione em *Authorized JavaScript
origins*:

```
https://hoodwink.hermis.com.br
```

Sem isso o popup/One Tap recusa a origem (o backend já está preparado para
validar o token).

## Deploy automático

O workflow `.github/workflows/ci.yml` roda testes no PR e, em **push para
`production`**, builda as três imagens na VM e faz `helm upgrade --install`.

Secrets/vars do repositório (os mesmos nomes do Summus):

| Tipo   | Nome         | Valor            |
| ------ | ------------ | ---------------- |
| var    | `VM_HOST`    | `144.33.18.156`  |
| secret | `VM_USER`    | `ubuntu`         |
| secret | `VM_SSH_KEY` | chave privada SSH |

Os segredos da aplicação vivem no **environment `production` do GitHub**. O job
de deploy cria/atualiza diretamente o Secret `hoodwink-secrets` no cluster com
`kubectl --from-literal`; **não existe arquivo `.env.production` na VM**.

```dotenv
PG_PASS=<senha do Postgres>
JWT_SECRET=<segredo do JWT do backend>
GAME_JWT_SECRET=<segredo legado do ticket; sai num deploy seguinte>
BACKEND_API_KEY=<uuid usado entre gameServer e backend>
REDIS_PASSWORD=<senha do Redis>
TICKET_PRIVATE_KEY=<base64 do PEM PKCS8 Ed25519>
TICKET_PUBLIC_KEY=<base64 do PEM PKIX Ed25519>
REPLAY_SIGNING_KEY=<base64 do PEM PKCS8 Ed25519>
REPLAY_PUBLIC_KEY_HEX=<64 hex da chave pública Ed25519>
```

Cadastre com `gh secret set <NOME> --env production` (o valor é pedido no
prompt, sem ficar no histórico). O `init-secrets.sh` é legado: só serve para
gerar pares localmente, não é usado no deploy. O `REPLAY_PUBLIC_KEY_HEX` é
passado no build do web (`--build-arg VITE_REPLAY_PUBLIC_KEY=...`); sem ele o
bundle usa a chave de dev
e os replays de produção aparecem como "não oficial" (ainda reproduzem).

## Chaves do ticket

O ticket do WebSocket é assinado pelo backend com **Ed25519** e verificado pelo
gameServer só com a chave pública. As chaves são geradas pelo
`init-secrets.sh`; para gerar manualmente:

```bash
openssl genpkey -algorithm ed25519 | base64 -w0   # TICKET_PRIVATE_KEY (PKCS8)
openssl pkey -pubout | base64 -w0                  # TICKET_PUBLIC_KEY (PKIX)
```

- **Formato**: base64 do PEM em uma linha (privada PKCS8, pública PKIX). O
  código faz `base64 -d` → `pem.Decode` → PKCS8/PKIX e sobe com erro se faltar
  ou for inválida (sem imprimir o valor).
- **Rotação**: troque o par nos dois serviços no mesmo deploy (o
  `helm upgrade --atomic` sobe backend e gameServer juntos). Tickets vivos de
  até 2h (`TICKET_EXPIRATION`) caem e o cliente pede outro no próximo
  `/getTicket`.
- **Ordem de deploy**: cadastre as chaves novas no environment `production`
  do GitHub ANTES do push e rode `sudo bash k8s/scripts/edge-nginx.sh`
  (bloqueia `/api/internal/` com 404). Backend e gameServer saem no mesmo
  `helm upgrade --atomic` — nunca só um deles, ou a entrada nas salas quebra.

## Deploy manual

```bash
cd ~/Hoodwink && git fetch origin && git checkout -f origin/production
SHA=$(git rev-parse --short HEAD)

sudo docker build -f server/backend/Dockerfile -t ghcr.io/matheusgoncalves540/hoodwink-backend:$SHA server
sudo docker build -f server/gameServer/Dockerfile -t ghcr.io/matheusgoncalves540/hoodwink-gameserver:$SHA server
set -a; . k8s/.env.production; set +a
sudo docker build \
  --build-arg "VITE_REPLAY_PUBLIC_KEY=${REPLAY_PUBLIC_KEY_HEX:-}" \
  -t ghcr.io/matheusgoncalves540/hoodwink-web:$SHA client/web
for image in backend gameserver web; do
  sudo docker save "ghcr.io/matheusgoncalves540/hoodwink-$image:$SHA" -o "/tmp/$image.tar"
  sudo k3s ctr images import "/tmp/$image.tar"
  sudo rm -f "/tmp/$image.tar"
done

kubectl apply -f k8s/namespace.yaml
kubectl -n hoodwink create secret generic hoodwink-secrets \
  --from-env-file=k8s/.env.production --dry-run=client -o yaml | kubectl apply -f -

helm upgrade --install hoodwink k8s/charts/hoodwink -n hoodwink \
  -f k8s/charts/hoodwink/values.production.yaml \
  --set imageTag="$SHA" --wait --atomic --timeout 10m
```

## Borda (uma vez / quando mudar o proxy)

```bash
sudo bash k8s/scripts/edge-nginx.sh        # config HTTP + reload
sudo bash k8s/scripts/edge-nginx.sh --tls  # depois do DNS: emite o certificado
```

## IP do jogador na borda (Cloudflare realip)

O `edge-nginx.sh` exige `http_realip_module` (`nginx -V 2>&1 | grep -q
http_realip_module`), chama o `cloudflare-realip.sh` e escreve no snippet do
domínio:

- `real_ip_header CF-Connecting-IP;`
- `include /etc/nginx/snippets/cloudflare-realip.conf;`
- `proxy_set_header X-Real-IP $remote_addr;` e `proxy_set_header CF-Connecting-IP "";`
  em todos os `location` de backend (`/api/`) e gameServer (`/gs/`, inclusive o
  upgrade do WebSocket).

O `cloudflare-realip.sh` baixa `https://www.cloudflare.com/ips-v4` e `/ips-v6`,
recusa lista vazia ou linha que não seja CIDR, grava o include, roda `nginx -t`
e só então recarrega (se falhar, restaura o arquivo anterior). Ele também roda
num timer systemd semanal:

```bash
# (manual na VM) ajuste o caminho para onde o repositório estiver
cat > /etc/systemd/system/cloudflare-realip.service <<'EOF'
[Unit]
Description=Atualiza faixas da Cloudflare (realip)
[Service]
Type=oneshot
ExecStart=/bin/bash /root/Hoodwink/k8s/scripts/cloudflare-realip.sh
EOF
cat > /etc/systemd/system/cloudflare-realip.timer <<'EOF'
[Unit]
Description=Atualiza faixas da Cloudflare semanalmente
[Timer]
OnCalendar=weekly
Persistent=true
[Install]
WantedBy=timers.target
EOF
systemctl daemon-reload
systemctl enable --now cloudflare-realip.timer
```

### Portas 8088/4444 fechadas para fora

Backend (8088) e gameServer (4444) usam `hostNetwork` e respondem no IP do nó;
quem fecha o acesso externo é a security list da Oracle. Conferir de fora da VM:

```bash
nc -zv -w3 <IP-público-da-VM> 8088   # não deve conectar
nc -zv -w3 <IP-público-da-VM> 4444   # não deve conectar
```

Se conectarem, feche as portas na security list da Oracle (o `/internal/*` já
fica atrás da API key, mas não deve ser exposto).

## Operação

Testes pós-deploy (o CI roda o smoke automaticamente):

```bash
bash k8s/scripts/smoke-test.sh   # health dos 3 serviços + schema/API key
bash k8s/scripts/e2e-game.sh     # JWT -> sala -> ticket -> WebSocket 101
```

O `e2e-game.sh` cria um usuário `smoke@example.com`, joga uma sala de teste e
remove o usuário no fim; a sala expira pelo `ROOM_TTL`.

```bash
kubectl -n hoodwink get pods,svc,deploy
helm -n hoodwink history hoodwink
kubectl -n hoodwink logs deploy/hoodwink-backend -f
kubectl -n hoodwink logs deploy/hoodwink-gameserver -f
curl https://hoodwink.hermis.com.br/api/alive
curl https://hoodwink.hermis.com.br/gs/alive
```

Rollback:

```bash
helm -n hoodwink rollback hoodwink <REVISION>
```

## Dados persistentes

| Componente | Caminho na VM             |
| ---------- | ------------------------- |
| Postgres   | `/opt/hoodwink/postgres`  |
| Redis      | `/opt/hoodwink/redis`     |

Backup manual do banco:

```bash
sudo docker run --rm --network host -e PGPASSWORD="<PG_PASS>" postgres:16-alpine \
  pg_dump -h 127.0.0.1 -U postgres -d hoodwink -Fc > ~/hoodwink-$(date +%F).dump
```
