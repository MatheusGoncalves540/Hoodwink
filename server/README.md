# 🎭 Hoodwink

> Um jogo de blefe estratégico e imprevisível, onde cartas são vidas, moedas são poder, e a mentira pode ser sua maior arma.

## 📌 Descrição

**Hoodwink** é um jogo de cartas baseado em blefe, onde os jogadores utilizam habilidades especiais para manipular a partida, eliminar oponentes e tentar ser o último sobrevivente. As cartas não são jogadas como em um baralho comum — elas representam as "vidas" dos jogadores, e cada uma possui habilidades únicas que podem ser utilizadas com ou sem possuí-las de fato.

---

## 🧠 Conceitos Principais

- Cada jogador começa com **2 cartas** (vidas) e **3 moedas** (definido em `gameServer/game/rules/modes/ClassicRules.yaml`).
- É possível **usar habilidades de cartas sem necessariamente tê-las**, blefando.
- Os outros jogadores podem **contestar** ações suspeitas.
- **Moedas** são utilizadas para ações básicas e ativação de habilidades.
- O objetivo do jogo é ser o **último jogador com cartas vivas na mão**.

---

## ▶️ Ações Básicas por Turno

Em cada turno, o jogador pode realizar **uma das ações abaixo**:

1. Passar o turno (deixar o tempo acabar).
2. **Coletar 1 moeda + taxas** (nada pode impedir).
3. **Pagar 6 + (taxas / 2, arredondado para baixo)** para matar 1 carta de um adversário (pode ser bloqueado pela Guardiã).
4. **Usar uma habilidade de carta** (mesmo sem possuí-la, blefando).
5. **Trocar carta(s) com o baralho**: pagando metade das suas moedas (arredondada para cima) por 1 carta, ou todas as moedas para trocar as 2. É preciso ter no mínimo **6 moedas**.

---

## ⚠️ Contestações

- Jogadores vivos podem contestar o uso de uma carta.
- Se o jogador **possuía a carta**: ela é revelada, o jogador escolhe uma carta do contestador para morrer e, depois, descarta a carta revelada e compra outra do baralho.
- Se o jogador **não possuía a carta**: o contestador escolhe uma carta do jogador para morrer e o efeito da carta é cancelado (as moedas gastas não são devolvidas).

---

## 💰 Limite de Moedas

- O limite máximo de moedas por jogador é **20**.
- Ao receber mais que isso, o excedente é **desconsiderado** e o jogador **perde automaticamente uma carta** (nada pode impedir essa perda).

---

## 📈 Sistema de Taxas

- As **taxas** aumentam o custo de uso das habilidades de cartas e são usadas em algumas ações básicas.
- No modo clássico elas não mudam aleatoriamente: só sobem com o **Político** e descem com o **Rebelde**.
- As taxas variam de **-5 a +5** e todos os cálculos são arredondados para baixo.
- Kamikaze, Político e Rebelde não são afetados pelas taxas.

---

## 🃏 Habilidades das Cartas

- Cada carta possui **uma habilidade especial**, com interações variadas entre elas.
- Algumas cartas possuem efeitos passivos, outras reativas, e outras ofensivas ou defensivas.

Exemplos:
- **Político**: aumenta as taxas em 1 (efeito acumulativo).
- **Assassino**: mata uma carta por 3 moedas (+ taxas).
- **Guardiã**: impede ataques.
- **Trapaceiro**: rouba moedas.
- **Egoísta**: impede que o adversário receba moedas ou cartas.
- **Kamikaze**: leva uma carta inimiga ao morrer.

Os valores e limites de cada carta ficam em `gameServer/game/rules/modes/ClassicRules.yaml`, que é a fonte de verdade do modo clássico.

---

## 🏁 Condições de Vitória

- Um jogador é **eliminado** ao perder todas as suas cartas.
- O jogo termina quando **restar apenas um jogador com cartas vivas**.

---

## 🛠️ Desenvolvimento

O ambiente de dev sobe o `backend` (API REST + PostgreSQL) e o `gameServer` (WebSocket + Redis) juntos a partir do `makefile` da raiz.

Pré-requisitos: Go 1.24+, Podman (máquina iniciada com `podman machine start`) e, opcionalmente, `goose` para migrations e `air` para live reload. No Windows, o `makefile` detecta e usa automaticamente o `sh.exe` do Git for Windows (instalação padrão em `C:\Program Files\Git`), então os alvos funcionam direto do PowerShell.

```bash
# Primeira execução: dependências Go + containers
make setup

# Sobe o ambiente completo (Ctrl+C encerra os dois serviços)
make dev
```

Cada serviço carrega o próprio `.env` (`backend/.env` e `gameServer/.env`), então mantenha-os criados a partir dos respectivos `.env.example`.

Outros alvos úteis:

- `make backend` / `make game` — sobe apenas um dos serviços
- `make infra-up` / `make infra-down` / `make infra-clean` — gerencia PostgreSQL e Redis
- `make migrate` — aplica as migrations pendentes (goose)
- `make build` / `make test` — compila ou testa os dois módulos
- `make help` — lista todos os alvos

---

## 📂 Estrutura do Projeto

> Este repositório contém o código-fonte, regras completas e eventuais assets do jogo Hoodwink.

---

## 👨‍💻 Contato

Para dúvidas, sugestões ou contribuições, entre em contato com o mantenedor do projeto.

---

