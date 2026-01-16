# Makefile para subir PostgreSQL e Redis Stack com Podman

# Diretórios locais
POSTGRES_VOLUME := ./postgres
REDIS_CONF := ./redis.conf

# Arquivo de ambiente (pode sobrescrever via: make ENV_FILE=../.env)
ENV_FILE ?= .env

# Carrega variáveis do .env se existir (linhas no formato KEY=VAL)
ifneq (,$(wildcard $(ENV_FILE)))
include $(ENV_FILE)
export $(shell sed -n 's/^\([A-Za-z_][A-Za-z0-9_]*\)=.*/\1/p' $(ENV_FILE))
endif

# URL de conexão construída
DB_URL := postgres://$(PG_USER):$(PG_PASS)@$(PG_HOST):$(PG_PORT)/$(PG_NAME)?sslmode=$(PG_SSLMODE)

# Containers
POSTGRES_CONTAINER := postgresdb
REDIS_CONTAINER := redisdb

# Imagens
POSTGRES_IMAGE := docker.io/library/postgres:18.1-bookworm
REDIS_IMAGE := docker.io/library/redis:8.0

.PHONY: all db-up redis-up start stop clean redis-conf install dev migrate migrate-down migrate-redo migrate-status migrate-create help

all: redis-conf db-up migrate redis-up install dev

## Cria o diretório e sobe o container do PostgreSQL
db-up:
	@mkdir -p $(POSTGRES_VOLUME)
	podman create --name $(POSTGRES_CONTAINER) \
	  -e POSTGRES_USER=postgres \
	  -e POSTGRES_PASSWORD=postgres \
	  -e POSTGRES_DB=hoodwink \
	  -p 5432:5432 \
	  -v $(POSTGRES_VOLUME):/var/lib/postgresql:Z \
	  $(POSTGRES_IMAGE)
	podman start $(POSTGRES_CONTAINER)

## Gera o arquivo redis.conf se não existir
redis-conf:
	@test -f $(REDIS_CONF) || printf '%s\n' \
'bind 127.0.0.1' \
'requirepass redis' \
'protected-mode yes' \
'appendonly yes' \
'rename-command FLUSHALL' "" \
'rename-command FLUSHDB' "" \
'rename-command CONFIG' "" \
'rename-command SHUTDOWN' "" \
'appendfsync everysec' > $(REDIS_CONF)

## Sobe o container do Redis Stack
redis-up: redis-conf
	podman create --name $(REDIS_CONTAINER) \
	  -p 127.0.0.1:6379:6379 \
	  -v $(CURDIR)/$(REDIS_CONF):/usr/local/etc/redis/redis.conf:ro \
	  $(REDIS_IMAGE) \
	  redis-server /usr/local/etc/redis/redis.conf
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

# Goose / migrations
GOOSE ?= goose
MIGRATIONS_DIR ?= $(CURDIR)/db/migrations

## Sobe (aplica) todas as migrations pendentes
migrate:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

## Desfaz a última migration (down 1)
migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

## Refaz a última migration
migrate-redo:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" redo

## Mostra status das migrations
migrate-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

## Cria nova migration (ex: make migrate-create name=create_users_table)
migrate-create:
	@test $(name) || (echo 'Uso: make migrate-create name=nome_da_migration'; exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(name) sql

## Ajuda (lista alvos principais)
help:
	@echo 'Alvos disponíveis:'
	@echo '  make all                -> Sobe redis+postgres e inicia dev'
	@echo '  make db-up              -> Sobe apenas Postgres'
	@echo '  make redis-up           -> Sobe apenas Redis'
	@echo '  make migrate            -> Aplica migrations pendentes'
	@echo '  make migrate-down       -> Desfaz última migration'
	@echo '  make migrate-redo       -> Refaz última migration'
	@echo '  make migrate-status     -> Mostra status'
	@echo '  make migrate-create name=descr -> Cria arquivo de migration SQL'
	@echo 'Variáveis úteis override: DB_HOST DB_PORT DB_USER DB_PASSWORD DB_NAME DB_SSLMODE ENV_FILE'
