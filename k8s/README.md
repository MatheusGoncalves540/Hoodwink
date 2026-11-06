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

Use **DNS only** para o certbot conseguir o desafio HTTP-01. Depois do
certificado emitido, o proxy laranja pode ser habilitado se o SSL da Cloudflare
ficar em **Full (strict)**.

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

Segredos da aplicação em `~/Hoodwink/k8s/.env.production` na VM:

```dotenv
PG_PASS=<senha do Postgres>
JWT_SECRET=<segredo do JWT do backend>
GAME_JWT_SECRET=<segredo do ticket do gameServer>
BACKEND_API_KEY=<uuid usado entre gameServer e backend>
REDIS_PASSWORD=<senha do Redis>
```

## Deploy manual

```bash
cd ~/Hoodwink && git fetch origin && git checkout -f origin/production
SHA=$(git rev-parse --short HEAD)

sudo docker build -t ghcr.io/matheusgoncalves540/hoodwink-backend:$SHA server/backend
sudo docker build -t ghcr.io/matheusgoncalves540/hoodwink-gameserver:$SHA server/gameServer
sudo docker build -t ghcr.io/matheusgoncalves540/hoodwink-web:$SHA client/web
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
