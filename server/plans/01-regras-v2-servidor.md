# Regras v2 no servidor: plano de execução

Contrato completo (vale para os dois lados): `Hoodwink-Web/plans/02-regras-v2.md`. Se este plano e o 02 divergirem, o 02 vence.
Escopo: só `gameServer/`. O front tem o próprio plano (`Hoodwink-Web/plans/04-handoff-refino-v2.md`).

## Antes de começar

- O repositório tem trabalho não commitado de outro agente (≈75 arquivos). **Não descarte, não reverta, não reformate.** Faça só as mudanças deste plano.
- Rode `go test ./...` em `gameServer/` antes de mexer, para saber o que já estava vermelho.
- Cada fase: teste que falha primeiro, depois o código, depois `go test ./...` verde.

## Já existe (não refazer)

| Item | Onde |
|---|---|
| `TAKE_COINS`, `KILL_CARD`, `SWAP_CARDS` | `game/structs/playerPlay.go`, causas em `structs/effects.go` |
| Payload da troca `{targetCardIndex, useOnTwoCards?}` | `structs/payloads.go` (`SwapCardsPayload`) |
| Clarividente para todos (`showToAllPlayers`) | `structs/payloads.go` |
| Custo da troca: `⌈moedas/2⌉`, as duas = todas | `structs/rooms/roomCards.go` `GetBasicSwapCost` |
| Taxa de −5 a +5 | `rules/modes/ClassicRules.yaml` |
| Trapaceiro fica com `⌈X/2⌉` | `engine/effects/stealCoinsEffect.go` (`CeilDiv`) |

## Fase 1. Preço do Matar (decidido)

Regra nova: `custo = 6 + (t > 0 ? ⌊t/2⌋ : t)`, **mínimo 3**.

- `game/structs/rooms/roomCards.go`, `GetBasicKillCardCost`: metade da taxa só quando ela é positiva; com taxa zero ou negativa ela entra inteira. Troque o piso `0` por `3`. Atualize o comentário da função.
- `game/structs/rooms/classicRules_test.go`, `TestBasicActionCosts`: tabela com
  `t: −5→3, −4→3, −3→3, −2→4, −1→5, 0→6, 1→6, 3→7, 5→8`.
  Hoje −4, −3, −2 e −1 dão 4, 4, 5 e 6; o teste tem que falhar antes da mudança.
- O mínimo 3 pode virar campo no YAML (`basicKillCardMinimum: 3`) se `GeneralRules` já seguir esse padrão; senão, constante com comentário.

## Fase 2. Sala até 10

- `ClassicRules.yaml`: `maxPlayers: 10`.
- Procure limites fixos em 6 fora do YAML (tamanho de sala, assentos, arrays). Teste: sala aceita o 10º e recusa o 11º.

## Fase 3. Conferir contra o 02 §3 e corrigir o que divergir

Para cada item: leia o código, compare com a tabela da seção 3 do 02. Se já bate, anote como feito no fim deste arquivo; se não, teste que falha + correção.

1. **Guardiã** (D6, decidido como proposto): reação `5 + t` mín 4 · contra Clarividente `4 + t` mín 2 · no próprio turno `6 + t` mín 5. O YAML hoje tem `taxMinimum: 2` e `futureProtectionTaxMinimum: 4`, que não batem. O servidor precisa saber qual dos três casos está cobrando.
2. **Coveiro**: revive **a mesma** carta morta do próprio ator (`engine/effects/cardEffects/gravediggerEffect.go`). Preço `5 + t` mín 3; não usa com a mão cheia.
3. **Crupiê**: descarta a carta escolhida (viva ou morta) e puxa uma nova do baralho. Preço `6 + t` mín 4.
4. **Egoísta**: limiar de moedas `2 + t` (o YAML tem `coinDenialThreshold: 2`; confira se o código soma a taxa). Bloqueio de carta `5 + t`, mín 3, máx 8.
5. **Investidor**: teto investido `clamp(4 + t, 2, 6)`; devolução depois de 6 turnos (`gracePeriod`); quando a taxa cai e o investido passa do teto, o excedente é **descartado**.
6. **Limite de 20 moedas**: descarta o excedente **e** o ator perde uma carta (`players/playerCoins.go` `AddCoins` devolve `breakLimit`; confira quem trata).
7. **Imunes à taxa**: Kamikaze, Político e Rebelde com `affectedByTaxes: false` no YAML.

## Fase 4. Eventos para o front

O front espera estes eventos no diário: `COIN_TAKEN`, `CARD_KILLED_PAID`, `CARDS_SWAPPED`, `INVESTMENT_RETURNED`, `INVESTMENT_DISCARDED`, `GREED_LIMIT`.
Se o servidor já emite algo equivalente com outro nome, **não renomeie**: escreva o nome real na seção 6 do `Hoodwink-Web/plans/02-regras-v2.md` (é o contrato).

## Fora deste plano (esperando o autor)

D1 (quinta ação básica), D2 (mínimo de 1 ao pegar moeda com taxa negativa; hoje dá 0), D4, D5, D7, D8, D9 do 02 §5. Não implemente sem resposta.

## Pronto quando

- `go test ./...` verde em `gameServer/`.
- Testes novos cobrindo as fases 1 e 2 e cada correção da fase 3.
- Nada do trabalho não commitado anterior perdido.
- Lista no fim deste arquivo: o que já batia, o que foi corrigido.

## Status

Status: em andamento

- [x] Fase 1. Preço do Matar
- [ ] Fase 2. Sala até 10
- [ ] Fase 3. Conferir contra o 02 §3 e corrigir o que divergir
- [ ] Fase 4. Eventos para o front
