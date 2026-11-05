# Regras v2 (modo Clássico): spec de handoff

Fonte: `HOODWINK.pdf` (regras do autor, nov/2026) + decisões do chat: Trapaceiro fica com a metade maior; sala de 2 a 10 atores.
Este arquivo é o contrato entre **Hoodwink-server** (gameServer) e **Hoodwink-Web**. Os dois lados implementam a mesma tabela; se algo divergir, este arquivo vence até ser atualizado.

Notação: `t` = taxa atual. `⌊x⌋` = arredondar para baixo (vale para TODO cálculo de taxa, ex.: 3/2 = 1). `clamp(v, mín, máx)`.

---

## 1. Regras gerais

| Regra | Valor |
|---|---|
| Atores por sala | 2 a 10 |
| Cartas por ator | 2 |
| Taxa | começa em 0, limite **−5 a +5** |
| Arredondamento | sempre para baixo (⌊⌋) |
| Papéis imunes à taxa | Kamikaze, Político, Rebelde |
| Limite de moedas | 20. Ao receber além disso: o excedente é descartado **e** o ator perde uma carta. Nada impede essa perda. |
| Carta que morre | vai para o baralho morto |
| Sem cartas vivas | vira espectador; não interage mais com a partida |
| Vitória | o último ator com cartas vivas |

## 2. Ações do turno (escolha 1)

Todas podem ser feitas **sem ter carta nenhuma**.

| Ação | Custo / ganho | Quem impede | Contestável |
|---|---|---|---|
| **Pegar moeda** | recebe `1 + t` | ninguém | não |
| **Matar (básico)** | paga `max(3, 6 + ⌊t/2⌋)` se `t > 0`, senão `max(3, 6 + t)` (taxa positiva entra pela metade; zero ou negativa entra inteira; **mínimo 3**). Ex.: t=−5 → 3, t=−3 → 3, t=−2 → 4, t=0 → 6, t=3 → 7, t=5 → 8. Mata 1 carta de um adversário | Guardiã | não |
| **Usar um papel** | ver seção 3 | conforme o papel | sim |
| **Trocar cartas** | paga `⌈moedas/2⌉`, descarta 1 carta da mão e puxa 1 do baralho. Exige ter **≥ 6 moedas**. Trocar as 2 de uma vez custa **todas** as moedas. | ninguém | não |

## 3. Papéis

| Papel | Quando | Alvo | Efeito | Preço | Limites | Quem impede |
|---|---|---|---|---|---|---|
| Político | seu turno | mesa | taxa +1 | 2, dobra a cada uso para todos até o fim da partida (2→4→8) | máx 8; a cada 3 turnos sem uso cai pela metade, mín 2 | — |
| Rebelde | seu turno | mesa | taxa −1 | igual ao Político | igual | — |
| Assassino | seu turno | carta de adversário | mata a carta | `3 + t` | mín 2 | Guardiã |
| Coveiro | seu turno | carta morta **sua** | revive **a mesma** carta que você tinha | `5 + t` | mín 3; não usa com a mão cheia | Egoísta |
| Crupiê | seu turno | carta sua (viva ou morta) | descarta e puxa uma nova aleatória do baralho | `6 + t` | mín 4; não usa com a mão cheia (ver D5) | Egoísta |
| Trapaceiro | seu turno | adversário | tira `X = clamp(2 + t, 2, 5)` dele e fica com `⌈X/2⌉` | — | tirado: 2 a 5 | Egoísta, se o ganho > 1 |
| Clarividente | seu turno | carta de adversário | revela a carta para todos **ou** só para quem usou; depois ela volta a ficar escondida, no mesmo lugar | `4 + t` | mín 2 | Guardiã |
| Trilionário | seu turno | — | recebe `3 + t` | — | mín 1 recebido | Egoísta, se o ganho ≥ limiar dela |
| Kamikaze | quando morre / seu turno | carta de adversário | ao morrer, mata uma carta de outro ator; pode ser morto de propósito no seu turno | 0 | imune à taxa | ninguém |
| Guardiã | reação ou seu turno | carta sua | impede um ataque a uma carta sua | reação: `5 + t`, mín 4 · contra Clarividente: `4 + t`, mín 2 · no seu turno (protege o próximo ataque): `6 + t`, mín 5 (D6) | uma carta protegida por vez; proteger outra tira a anterior | — |
| Egoísta | só fora do seu turno | adversário | a) por **1**, impede de receber `2 + t` ou mais moedas · b) por `5 + t`, impede de receber 1 carta (Crupiê, Coveiro) · c) impede um investimento | a) 1 fixo · b) `5 + t` | b) mín 3, máx 8 | — |
| Investidor | seu turno | — | investe **1** moeda (1 por turno) | a moeda investida | investido no máximo `clamp(4 + t, 2, 6)` por ator | Egoísta impede o investimento, nunca o rendimento |

### Investidor em detalhe
- Rendimento passivo: em todo turno do dono, recebe `⌊investido/2⌋` (só moedas inteiras).
- Cada moeda investida fica 6 turnos e depois volta para a mão do dono.
- Se a taxa cair e o investido passar do novo máximo, o excedente é **descartado** (não volta).
- Só é contestável no momento do investimento.

## 4. Contestação

Qualquer ator ativo pode contestar quando um **papel** é usado (ações básicas não são contestáveis).

- **Ele tinha a carta:** revela para todos → escolhe 1 carta de quem contestou para morrer → descarta a carta revelada e puxa outra do baralho. A jogada acontece.
- **Ele não tinha:** quem contestou escolhe 1 carta dele para morrer. A jogada não acontece.
- A Guardiã não impede morte por contestação.

## 5. Decisões em aberto (o autor responde antes de fechar)

- **D1.** O PDF fala em "1 de 5 ações básicas", mas lista 4. Qual é a quinta? (Hipótese: sacrificar o Kamikaze.)
- **D2.** "Pegar 1 moeda + taxas" com taxa negativa dá 0 ou menos. Proposta: mínimo 1.
- **D3.** (Resolvida para o Matar: com t ≤ 0 a taxa entra inteira, sem metade.) Se outra regra usar ⌊t/2⌋ com t negativo: −5/2 é −3 (para baixo de verdade) ou −2 (truncado, que é o que o `/` do Go faz)? Proposta: −3, para seguir "sempre para baixo".
- **D4.** Clarividente: quem escolhe se a revelação é para todos ou só para quem usou? Proposta: quem usa, no momento de declarar.
- **D5.** Crupiê: se ele troca uma carta (inclusive morta), a regra "não pode com a mão cheia" vale? Proposta: retirar essa restrição do Crupiê.
- **D6.** Guardiã no seu turno: o mínimo é 5 (4 + 1)? Proposta: sim.
- **D7.** A taxa aleatória (hoje 5% por vez que a vez passa) continua? O PDF não fala.
- **D8.** Espectador ainda conversa no chat? O PDF diz "não interage mais com a partida"; hoje o front deixa conversar. Proposta: chat sim, jogada não.
- **D9.** Os tempos (10 s por deixa, morte súbita com 5 s) e as 3 moedas iniciais continuam? O PDF não fala.

## 6. Contrato de mensagens novas (proposta)

As três ações básicas já existem no servidor (`gameServer/game/structs/playerPlay.go`) com os nomes abaixo. **Estes são os nomes oficiais**; o front usa exatamente estes. Payload no mesmo formato dos papéis (`net/actions.ts → play`):

| Tipo | Payload |
|---|---|
| `TAKE_COINS` | `{}` |
| `KILL_CARD` | `{ targetPlayer, targetCardIndex }` |
| `SWAP_CARDS` | `{ targetCardIndex, useOnTwoCards? }` (`useOnTwoCards: true` troca as duas e custa todas as moedas) |
| `CLAIRVOYANT` | `showToAllPlayers: boolean` (já existe no servidor) |
| `GUARDIAN` (mudança) | o servidor diferencia reação, contra Clarividente e no seu turno, para cobrar o preço certo |

Eventos de apresentação esperados no front: `COIN_TAKEN`, `CARD_KILLED_PAID`, `CARDS_SWAPPED`, `INVESTMENT_RETURNED`, `INVESTMENT_DISCARDED`, `GREED_LIMIT` (passou de 20).

---

## 7. O que muda no servidor (Hoodwink-server/gameServer)

- `rules/modes/ClassicRules.yaml`: `maxPlayers 10`, `maxTax 5`, `minTax −5`; valores e mínimos/máximos da tabela 3 (Coveiro 5/mín 3, Crupiê mín 4, Guardiã mín 4, Egoísta 5/mín 3/máx 8, Investidor 4/2..6, sem `taxMaximum` onde o PDF não define). Kamikaze, Político e Rebelde com `affectedByTaxes: false`.
- `effects/stealCoinsEffect.go`: `earnedCoins = (LostCoins + 1) / 2`.
- Coveiro: revive a mesma carta morta do próprio ator (hoje é aleatória do baralho morto).
- Crupiê: descarta a carta escolhida e puxa uma nova (hoje "ressuscita nova carta").
- Ações básicas novas (seção 2) + validações (≥ 6 moedas para trocar; trocar 2 custa tudo).
- Egoísta: limiar de moedas `2 + t` (hoje é 2 fixo).
- Investidor: devolução após 6 turnos; descarte do excedente quando a taxa cai.
- Limite de 20: descartar o excedente antes de matar a carta.
- Arredondamento ⌊⌋ em todo cálculo de taxa (D3).

## 8. O que muda no front (Hoodwink-Web/web/src)

- `game/rules.classic.ts`: espelhar o YAML novo (o comentário no topo já pede isso).
- `game/pricing.ts` + `pricing.test.ts`: fórmulas novas (Trapaceiro ⌈X/2⌉, Guardiã com 3 preços, Egoísta com limiar `2 + t`, Investidor com teto investido) e preço das ações básicas.
- `store/sessionStore.ts`: `DEFAULT_MAX_PLAYERS = 10`; testar a mesa (`SeatRing`) com 10 camarotes a 390 px.
- `features/turn/ActionBoard.tsx`: grupo novo **Ações do turno** no topo (Pegar moeda, Matar, Trocar cartas), com "Precisa de 6 moedas" desabilitado e aviso "custa todas as suas moedas" ao trocar as duas.
- `features/turn/TargetBar.tsx`: alvo de carta própria para Trocar (1 ou 2 cartas); opção "para todos / só para mim" no Clarividente.
- `net/actions.ts`: `takeCoins` (`TAKE_COINS`), `killCard` (`KILL_CARD`), `swapCards` (`SWAP_CARDS`).
- `game/copy.ts`: textos dos papéis e do diário (seção 3 e eventos da seção 6).
- Cena nova **Como se joga** (canvas "Hoodwink — Refino v2"), ligada aos links `#como-se-joga` de `Lobby.tsx:139` e `RoomMenu.tsx:39`. A ficha de cada papel lê `pricing.ts`, então o texto e o preço nunca divergem do jogo.
