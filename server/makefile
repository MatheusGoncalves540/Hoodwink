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

# Replay WASM (plano 05): um .wasm por versão do motor + manifest do bundle.
REPLAY_DIR      := replay
WEB_DIR         := ../client/web
REPLAY_PUBLIC   := $(WEB_DIR)/public/replay
REPLAY_MANIFEST := $(WEB_DIR)/src/replay/manifest.gen.ts

# Runner: air (live reload) quando instalado, senão go run
AIR := $(shell command -v air 2>/dev/null)
ifeq ($(AIR),)
RUNNER := go run .
else
RUNNER := $(AIR)
endif

.PHONY: all dev backend game check-env infra-up infra-down infra-clean \
        deps setup build test goose-check migrate migrate-down migrate-redo \
        migrate-status migrate-create clean engine-check engine-bump \
        replay-wasm replay-wasm-check help

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

## Verifica a versão e o fecho do motor (hash, time.Now/rand., dependências, cópias vN)
engine-check:
	@set -e
	( cd tools && go run ./enginecheck )

## Sobe a versão e congela o fecho do motor em replay/engines/vN
engine-bump:
	@set -e
	( cd tools && go run ./enginebump )

## Compila um WASM do motor por versão (Go padrão, brotli) e gera o manifest do bundle
##
## O .wasm e o .wasm.br ficam em client/web/public/replay (o Vite serve a pasta
## public) e o manifest em client/web/src/replay/manifest.gen.ts. Nome com hash,
## sha256 dos bytes descomprimidos (o que o loader confere antes de instanciar),
## sha256 do .wasm.br servido pelo nginx e sha256 do wasm_exec.js da toolchain.
replay-wasm:
	@set -e
	command -v go >/dev/null 2>&1 || { echo 'ERRO: go nao encontrado no PATH.'; exit 1; }
	command -v brotli >/dev/null 2>&1 || { echo 'ERRO: brotli nao encontrado. O T12 exige brotli para comprimir o .wasm (apt-get install brotli / brew install brotli).'; exit 1; }
	command -v sha256sum >/dev/null 2>&1 || { echo 'ERRO: sha256sum nao encontrado no PATH.'; exit 1; }
	versions="$$(grep -o 'Number: [0-9]*' $(REPLAY_DIR)/versions/versions.go | grep -o '[0-9][0-9]*')"
	test -n "$$versions" || { echo 'ERRO: nenhuma versao encontrada em $(REPLAY_DIR)/versions/versions.go'; exit 1; }
	mkdir -p $(REPLAY_PUBLIC) $(WEB_DIR)/src/replay
	rm -f $(REPLAY_PUBLIC)/replay-v*.wasm.br $(REPLAY_PUBLIC)/replay-v*.wasm
	tmp="$$(mktemp -d)"
	trap 'rm -rf "$$tmp"' EXIT
	goroot="$$(cd $(REPLAY_DIR) && go env GOROOT)"
	wasm_exec="$$goroot/lib/wasm/wasm_exec.js"
	test -f "$$wasm_exec" || wasm_exec="$$goroot/misc/wasm/wasm_exec.js"
	test -f "$$wasm_exec" || { echo "ERRO: wasm_exec.js nao encontrado em $$goroot/lib/wasm nem em $$goroot/misc/wasm (toolchain Go sem wasm_exec.js?)"; exit 1; }
	cp "$$wasm_exec" $(REPLAY_PUBLIC)/wasm_exec.js
	exec_sha="$$(sha256sum $(REPLAY_PUBLIC)/wasm_exec.js | cut -d' ' -f1)"
	echo ">> wasm_exec.js ($$(cd $(REPLAY_DIR) && go env GOVERSION)) sha256 $$exec_sha"
	printf '%b' '// Code generated por `make replay-wasm` (server/makefile). NAO edite a mao.\n//\n// Um .wasm por versao do motor congelado. O `sha256` e dos bytes DESCOMPRIMIDOS\n// do .wasm - e o que o loader confere antes de WebAssembly.instantiate. Em\n// producao o nginx serve o .wasm.br com Content-Encoding: br, entao o fetch ja\n// entrega os bytes descomprimidos. O `sha256Br` e do arquivo .wasm.br em si\n// (integridade do artefato publicado; nao e o hash usado no instantiate).\n\nexport type ReplayEngineAsset = {\n  version: number;\n  /** Nome do arquivo em public/replay/, comprimido com brotli. */\n  file: string;\n  /** sha256 (hex) dos bytes descomprimidos do .wasm. */\n  sha256: string;\n  /** sha256 (hex) do arquivo .wasm.br servido pelo nginx. */\n  sha256Br: string;\n  /** sha256 (hex) do wasm_exec.js da toolchain Go usada no build. */\n  wasmExecSha256: string;\n  brotli: true;\n};\n\nexport const REPLAY_WASM_BASE = "/replay/";\nexport const REPLAY_WASM_EXEC_FILE = "wasm_exec.js";\n\nexport const replayEngines: readonly ReplayEngineAsset[] = [\n' > $(REPLAY_MANIFEST)
	for v in $$versions; do
		wasm="$$tmp/replay-v$$v.wasm"
		( cd $(REPLAY_DIR) && GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w -X main.buildEngineVersion=$$v" -o "$$wasm" ./cmd/wasm )
		sha="$$(sha256sum "$$wasm" | cut -d' ' -f1)"
		short="$$(printf '%s' "$$sha" | cut -c1-8)"
		file="replay-v$$v-$$short.wasm.br"
		raw="replay-v$$v-$$short.wasm"
		cp "$$wasm" "$(REPLAY_PUBLIC)/$$raw"
		brotli -q 11 -f -o "$(REPLAY_PUBLIC)/$$file" "$$wasm"
		brsha="$$(sha256sum "$(REPLAY_PUBLIC)/$$file" | cut -d' ' -f1)"
		bytes="$$(wc -c < "$$wasm" | tr -d ' ')"
		brbytes="$$(wc -c < "$(REPLAY_PUBLIC)/$$file" | tr -d ' ')"
		printf '>> v%s: %s + %s (%s B brotli, %s B wasm, sha256 %s)\n' "$$v" "$$raw" "$$file" "$$brbytes" "$$bytes" "$$short"
		printf '  { version: %s, file: "%s", sha256: "%s", sha256Br: "%s", wasmExecSha256: "%s", brotli: true },\n' "$$v" "$$file" "$$sha" "$$brsha" "$$exec_sha" >> $(REPLAY_MANIFEST)
	done
	printf '%b' '];\n\nexport const replayEngineByVersion = (version: number): ReplayEngineAsset | undefined =>\n  replayEngines.find((engine) => engine.version === version);\n\nexport const replayAssetUrl = (file: string): string => REPLAY_WASM_BASE + file;\n' >> $(REPLAY_MANIFEST)
	echo ">> manifest: $(REPLAY_MANIFEST)"

## Confere os assets do replay: arquivos existem, sha256 bate com o manifest e tamanho no teto
replay-wasm-check:
	@set -e
	command -v node >/dev/null 2>&1 || { echo 'ERRO: node nao encontrado (necessario para conferir o manifest).'; exit 1; }
	node $(REPLAY_DIR)/cmd/wasm/check_manifest.mjs $(WEB_DIR)

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
	@echo '  make engine-check              -> Verifica a versao/fecho do motor (replay)'
	@echo '  make engine-bump               -> Congela o motor em replay/engines/vN'
	@echo '  make replay-wasm               -> Compila o WASM por versao + manifest do bundle'
	@echo '  make replay-wasm-check         -> Confere sha256/tamanho dos assets do replay'
	@echo '  make clean                     -> Remove containers e binarios'
	@echo ''
	@echo 'Runner: air (live reload) se instalado, senao "go run ."'
