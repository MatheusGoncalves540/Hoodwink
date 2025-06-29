# Makefile para subir PostgreSQL e Redis Stack com Podman

# Diretórios locais
POSTGRES_VOLUME := ./postgres
REDIS_CONF := ./redis.conf

# Containers
POSTGRES_CONTAINER := postgresdb
REDIS_CONTAINER := redisdb

# Imagens
POSTGRES_IMAGE := docker.io/library/postgres
REDIS_IMAGE := docker.io/redis/redis-stack-server:latest

.PHONY: all db-up redis-up start stop clean redis-conf install dev

all: db-up redis-up install dev

## Cria o diretório e sobe o container do PostgreSQL
db-up:
	@mkdir -p $(POSTGRES_VOLUME)
	podman create --name $(POSTGRES_CONTAINER) \
	  -e POSTGRES_USER=postgres \
	  -e POSTGRES_PASSWORD=postgres \
	  -e POSTGRES_DB=hoodwink \
	  -p 5432:5432 \
	  -v $(POSTGRES_VOLUME):/data/db \
	  $(POSTGRES_IMAGE)
	podman start $(POSTGRES_CONTAINER)

## Gera o arquivo redis.conf se não existir
redis-conf:
	@test -f $(REDIS_CONF) || printf '%s\n' \
'requirepass redis' \
'appendonly yes' \
'appendfsync everysec' \
'loadmodule /opt/redis-stack/lib/redisearch.so MAXSEARCHRESULTS 10000 MAXAGGREGATERESULTS 10000' \
'loadmodule /opt/redis-stack/lib/rediscompat.so' \
'loadmodule /opt/redis-stack/lib/redisbloom.so' \
'loadmodule /opt/redis-stack/lib/redistimeseries.so' \
'loadmodule /opt/redis-stack/lib/redisgears.so v8-plugin-path /opt/redis-stack/lib/libredisgears_v8_plugin.so' \
'loadmodule /opt/redis-stack/lib/rejson.so' > $(REDIS_CONF)

## Sobe o container do Redis Stack
redis-up: redis-conf
	podman create --name $(REDIS_CONTAINER) \
	  -p 6379:6379 \
	  -v $(CURDIR)/$(REDIS_CONF):/etc/redis.conf:ro \
	  $(REDIS_IMAGE) \
	  redis-server /etc/redis.conf
	podman start $(REDIS_CONTAINER)

## Inicia os containers caso já existam
start:
	-podman start $(POSTGRES_CONTAINER)
	-podman start $(REDIS_CONTAINER)

## Para os containers
stop:
	-podman stop $(POSTGRES_CONTAINER)
	-podman stop $(REDIS_CONTAINER)

## Remove os containers
clean: stop
	-podman rm $(POSTGRES_CONTAINER)
	-podman rm $(REDIS_CONTAINER)

## Inicia o projeto
install:
	-go mod tidy

## Inicia o projeto
dev:
	-air
