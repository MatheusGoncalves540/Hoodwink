# Makefile raiz: ambiente único de desenvolvimento (backend + gameServer)
#
# O backend (API REST/PostgreSQL) e o gameServer (WebSocket/Redis) continuam
# sendo módulos Go independentes. Este arquivo apenas orquestra os dois,
# reutilizando os containers definidos em backend/makefile.

# O make nativo do Windows usa cmd.exe por padrão, o que quebra as receitas
# POSIX deste arquivo. Detectamos o sh do Git for Windows e o usamos como
# SHELL, garantindo também rm/mkdir/sed no PATH.
ifeq ($(OS),Windows_NT)
  GIT_SH  := $(wildcard C:/Program\ Files/Git/usr/bin/sh.exe)
  GIT_USR := C:/Program Files/Git/usr/bin
  ifeq ($(GIT_SH),)
    GIT_SH  := $(wildcard C:/Program\ Files/Git/bin/sh.exe)
    GIT_USR := C:/Program Files/Git/bin
  endif
  ifeq ($(GIT_SH),)
    $(error Git Bash nao encontrado em "C:\Program Files\Git". Instale o Git for Windows ou rode o make pelo Git Bash)
  endif
  SHELL := $(GIT_SH)
  export PATH := $(GIT_USR);$(PATH)
else
  SHELL := sh
endif
.ONESHELL:
.DEFAULT_GOAL := dev

# Serviços
BACKEND_DIR    := backend
GAMESERVER_DIR := gameServer
BIN_DIR        := bin

# Containers gerenciados por backend/makefile
POSTGRES_CONTAINER := postgresdb
REDIS_CONTAINER    := redisdb

# Runner: air (live reload) quando instalado, senão go run
AIR := $(shell command -v air 2>/dev/null)
ifeq ($(AIR),)
RUNNER := go run .
else
RUNNER := $(AIR)
endif

.PHONY: all dev backend game check-env infra-up infra-down infra-clean \
        deps setup build test goose-check migrate migrate-down migrate-redo \
        migrate-status migrate-create clean help

## Alias para o alvo padrão
all: dev

## Sobe o ambiente completo: containers, backend e gameServer
dev: check-env infra-up
	@set -e
	pids=""
	cleanup() { kill $$pids 2>/dev/null || true; }
	trap cleanup INT TERM EXIT
	echo '>> Subindo backend e gameServer (runner: $(RUNNER))'
	( cd $(BACKEND_DIR) && exec $(RUNNER) ) & pids="$$pids $$!"
	( cd $(GAMESERVER_DIR) && exec $(RUNNER) ) & pids="$$pids $$!"
	echo '>> Ambiente de dev no ar. Ctrl+C encerra os dois servicos.'
	wait

## Sobe apenas o backend (API REST)
backend: check-env infra-up
	@cd $(BACKEND_DIR) && exec $(RUNNER)

## Sobe apenas o gameServer (WebSocket)
game: check-env infra-up
	@cd $(GAMESERVER_DIR) && exec $(RUNNER)

## Valida que cada serviço tem seu .env (godotenv carrega a partir do CWD)
check-env:
	@set -e
	test -f $(BACKEND_DIR)/.env || { echo 'ERRO: $(BACKEND_DIR)/.env nao encontrado. Copie de $(BACKEND_DIR)/.env.example'; exit 1; }
	test -f $(GAMESERVER_DIR)/.env || { echo 'ERRO: $(GAMESERVER_DIR)/.env nao encontrado. Copie de $(GAMESERVER_DIR)/.env.example'; exit 1; }

## Sobe PostgreSQL e Redis (cria os containers se ainda não existirem)
infra-up:
	@set -e
	podman info >/dev/null 2>&1 || { echo 'ERRO: Podman nao esta respondendo. Inicie a maquina: podman machine start'; exit 1; }
	podman container exists $(POSTGRES_CONTAINER) >/dev/null 2>&1 || $(MAKE) -C $(BACKEND_DIR) SHELL="$(SHELL)" db-up
	podman container exists $(REDIS_CONTAINER) >/dev/null 2>&1 || $(MAKE) -C $(BACKEND_DIR) SHELL="$(SHELL)" redis-up
	$(MAKE) -C $(BACKEND_DIR) SHELL="$(SHELL)" start
	printf 'Aguardando PostgreSQL'
	i=0
	until podman exec $(POSTGRES_CONTAINER) pg_isready -U postgres -q 2>/dev/null; do
		i=$$((i+1))
		[ $$i -ge 30 ] && { echo ' timeout!'; exit 1; }
		printf '.'
		sleep 1
	done
	echo ' ok'

## Para PostgreSQL e Redis
infra-down:
	@$(MAKE) -C $(BACKEND_DIR) SHELL="$(SHELL)" stop

## Remove os containers de infraestrutura
infra-clean:
	@$(MAKE) -C $(BACKEND_DIR) SHELL="$(SHELL)" clean

## Baixa as dependências Go dos dois módulos
deps:
	@set -e
	( cd $(BACKEND_DIR) && go mod download )
	( cd $(GAMESERVER_DIR) && go mod download )

## Prepara o ambiente: dependências + containers
setup: deps infra-up

## Compila os dois serviços em ./bin
build:
	@set -e
	mkdir -p $(BIN_DIR)
	( cd $(BACKEND_DIR) && go build -o ../$(BIN_DIR)/backend . )
	( cd $(GAMESERVER_DIR) && go build -o ../$(BIN_DIR)/gameserver . )

## Roda os testes dos dois módulos
test:
	@set -e
	( cd $(BACKEND_DIR) && go test ./... )
	( cd $(GAMESERVER_DIR) && go test ./... )

# Goose / migrations (delegado ao backend/makefile)
goose-check:
	@command -v goose >/dev/null 2>&1 || { echo 'ERRO: goose nao encontrado. Instale com: go install github.com/pressly/goose/v3/cmd/goose@latest'; exit 1; }

## Aplica as migrations pendentes
migrate: goose-check
	@$(MAKE) -C $(BACKEND_DIR) SHELL="$(SHELL)" migrate

## Desfaz a última migration
migrate-down: goose-check
	@$(MAKE) -C $(BACKEND_DIR) SHELL="$(SHELL)" migrate-down

## Refaz a última migration
migrate-redo: goose-check
	@$(MAKE) -C $(BACKEND_DIR) SHELL="$(SHELL)" migrate-redo

## Mostra o status das migrations
migrate-status: goose-check
	@$(MAKE) -C $(BACKEND_DIR) SHELL="$(SHELL)" migrate-status

## Cria uma migration (ex: make migrate-create name=create_users_table)
migrate-create: goose-check
	@$(MAKE) -C $(BACKEND_DIR) SHELL="$(SHELL)" migrate-create name="$(name)"

## Remove containers e binários
clean: infra-clean
	@rm -rf $(BIN_DIR)

## Lista os alvos disponíveis
help:
	@echo 'Hoodwink - ambiente de desenvolvimento'
	@echo ''
	@echo '  make dev                       -> Sobe containers + backend e gameServer (Ctrl+C encerra tudo)'
	@echo '  make backend                   -> Sobe apenas o backend (API REST)'
	@echo '  make game                      -> Sobe apenas o gameServer (WebSocket)'
	@echo ''
	@echo '  make setup                     -> Baixa dependencias Go e sobe PostgreSQL + Redis'
	@echo '  make deps                      -> go mod download nos dois modulos'
	@echo '  make infra-up                  -> Sobe PostgreSQL + Redis'
	@echo '  make infra-down                -> Para PostgreSQL + Redis'
	@echo '  make infra-clean               -> Remove os containers de infraestrutura'
	@echo ''
	@echo '  make migrate                   -> Aplica migrations pendentes (goose)'
	@echo '  make migrate-down              -> Desfaz a ultima migration'
	@echo '  make migrate-redo              -> Refaz a ultima migration'
	@echo '  make migrate-status            -> Mostra o status das migrations'
	@echo '  make migrate-create name=xxx   -> Cria um arquivo de migration SQL'
	@echo ''
	@echo '  make build                     -> Compila os dois servicos em ./bin'
	@echo '  make test                      -> Roda os testes dos dois modulos'
	@echo '  make clean                     -> Remove containers e binarios'
	@echo ''
	@echo 'Runner: air (live reload) se instalado, senao "go run ."'
