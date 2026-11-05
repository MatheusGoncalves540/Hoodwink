# Revisão 01: varredura do prompt 10 (T27)

Status: **concluída**. As 21 telas de jogo existem e foram abertas no navegador em 390×844 e 1440×900, nos dois temas, com o console limpo. Gates: `npm test` (110 testes, 19 arquivos) e `npm run build` verdes.

Método: greps do prompt 10 em `web/src`, leitura dos CSS Modules e conferência das capturas com o app rodando (Chrome DevTools MCP) em `?cena=…&fixture=…`. Sem Playwright; o Lighthouse foi rodado em snapshot na mesa.

## 1. Cores fora dos tokens

Comando: `grep -rnE "#[0-9a-fA-F]{3,8}|rgb\(" web/src`, descontando `web/src/styles/` (cópia do handoff, fonte dos tokens) e as cores fixas permitidas em `02-design-system.md`.

| Local | Ocorrência | Classificação |
| --- | --- | --- |
| `features/table/Seat.module.css` | listras do camarote fora de cena em `#5A1820` / `#7A2430` | **corrigido**: `var(--hw-veludo)` + `color-mix` com `--hw-sobre-veludo` |
| `features/turn/ActionButton.*` e `RoleEyes.tsx` | fundo/faixas/texto da arte dos papéis | **aceitável**: “faixa de arte das ações” é cor fixa prevista |
| `features/endgame/SuddenDeath*`, `styles/sudden-death.css` | fogo, cortina, holofote | **aceitável**: “toda a morte súbita” é cor fixa prevista |
| `ui/Card/Card.module.css`, `ui/Coin/Coin.module.css` | `#1e1512` | **aceitável**: texto das moedas, cor fixa prevista |
| `features/endgame/Finale.module.css`, `features/turn/DeclaredCard.module.css` | `rgba(185,138,46,…)` (ouro com alfa) e sombra | **aceitável**: ornamento/sombra, sem variação de tema |
| `componentsTSX/styles/loginModal.css` | `#3b82f6`, `#2563eb`, `#666`, `#333`, `#e5e7eb`, `#d1d5db`, `rgba(0,0,0,.5)` | **inferido**: o modal de login não está nas pranchas; fica o visual legado até o produto decidir |

## 2. Movimento (`transform`/`opacity`)

Nenhum `transition` ou `animation` de tela nova anima propriedade fora de `transform`/`opacity`.

- `ui/Sheet` e as entradas `m-*` animam `transform`/`opacity`.
- `ui/Timer` usa `transform: scaleX(var(--de))` via `m-timer`.
- A morte súbita usa `opacity`, `transform`, e `visibility` só em `.ms-some` (a cortina deixa de existir aos 3,3 s), previsto na linha do tempo.
- Exceção legada: `componentsTSX/styles/loginModal.css` tem `transition: background-color`; fica no relatório junto do modal, fora das 21 telas.
- Movimento reduzido conferido no navegador: a cortina não aparece, o fogo fica parado no primeiro quadro e o aviso entra com fade.

## 3. Texto fora de `game/copy.ts`

- **Corrigido**: as 5 regiões com `aria-label` literal ganharam chave em `copy.ts`:
  - `features/lobby/Lobby.tsx` — “Cartaz” e “Jogar” (`copy.cartazRegion`, `copy.playRegion`)
  - `features/room/Cast.tsx` — “Elenco” (`copy.castRegion`)
  - `features/room/Invite.tsx` — “Convite” (`copy.inviteRegion`)
  - `features/room/WaitingRoom.tsx` — “Começar” (`copy.startRegion`)
- `web/src/ui/`: nenhum texto literal.
- `app/Amostras.tsx`: textos literais, mas é a página de amostras de desenvolvimento (`?cena=amostras`), fora das 21 telas. **Aceitável**.
- `componentsTSX/loginModal.tsx` (“Login”, “Fechar”, descrição): legado do login, mantido pela T26. **Inferido**.
- O login de depuração que ficava no `App.tsx` foi removido na T25; a entrada agora é o `Lobby`.

## 4. Preço fixo fora de `pricing.ts`

Nenhuma ocorrência nas telas: o painel de ações usa `priceOf`, e os efeitos da ribalta vêm de `selectFootlights`. Os números de `app/Amostras.tsx` (“por 6 moedas”, moeda 7) são da página de amostras.

## 5. Revisão tela a tela

Conferidas contra as capturas (`referencias/capturas/`), no app rodando:

| Tela | Capturas | Resultado |
| --- | --- | --- |
| Entrada (T14) | `Lobby`, `LobbyDesk`, `LobbyDeskEscuro` | ok |
| Sala de espera (T15) | `Sala`, `SalaDesk`, `SalaCheia` | ok |
| Mesa mobile (T16) | `Mesa6`, `Mesa10` | ok |
| Mesa desktop (T17) | `Mesa6Desk`, `Mesa6DeskEscuro` | ok |
| Camarote e Você (T18) | `Camarote`, `Voce`, `VoceDesk` | ok |
| Menu e conversa (T19) | `Menu`, `ConversaCheiaDesk` | ok |
| Trilho (T20) | `Mesa6Desk` | ok |
| Seu turno (T21) | `Acoes`, `AcoesDesk` | ok (preço e “na mão” conferem) |
| Reação (T22) | `Main`, `MainDesk` | ver diferenças abaixo |
| Contestações (T22) | `ContestaBlefe`, `ContestaVerdade`, `*Desk` | ok; corrigido o estado da carta protegida |
| Fora de cena (T23) | `ForaDeCena`, `ForaDeCenaDesk` | ok |
| Fim (T23) | `Fim`, `FimDesk` | ok |
| Morte súbita (T24) | `MorteSubita`, `MorteSubitaDesk` | ok, incluindo holofote a 3,45 s |
| Mesa em morte súbita (T24) | `MesaSubita`, `MesaSubitaDesk` | ok |

Diferenças encontradas e classificação:

- **Defeito corrigido:** em `ContestaVerdade`, a carta protegida aparecia como “na mão”. O `ContestPick` lia `c.protected`, campo que não existe no `PickCard`; passou a ler `c.state` e a mostrar a face das suas cartas conhecidas com o rótulo “na mão” quando não estão reveladas. Coberto em `Turn.test.tsx`.
- **Defeito bloqueado pelo servidor:** em `Main`, a prancha mostra “Bloquear com Guardiã” na reação ao Assassino. `guardianValidation.go` só aceita Guardiã em `WAITING_FIRST_ACTION`, `CARD_KILLED` e `CARD_PLAYED_CLAIRVOYANT`; a jogada seria recusada. Mantido só o bloqueio do Clarividente e o pedido ao servidor (seção 10).
- **Inferido:** a prancha escreve “Ela paga 4 moedas…”, “Ela não tinha a carta…”, “Ela mostra a carta…”. O app usa o primeiro nome do ator (“Ana paga…”) para não errar o gênero de atores desconhecidos; o layout e o resto da frase conferem.
- **Inferido:** `waitPickHint` diz “Depois disso a jogada dela ainda mira você.”; a prancha diz “o Assassino dela”. A frase genérica cobre papéis que não atacam.
- **Aceitável:** o rodapé “Voltar à mesa” aparece com o anel de foco quando a gaveta abre no desktop (o Radix foca o primeiro elemento focável); a prancha é estática e não mostra foco.
- **Aceitável:** a captura de `ConversaCheiaDesk` mostra 5 linhas de diário e conversa vazia; a fixture tem 3 linhas e conversa. A estrutura do trilho confere.
- **Aceitável:** `FimDesk` tem o rótulo “FIM DO ESPETÁCULO” um pouco mais à esquerda e “Voltar à entrada” sem sublinhado; diferenças de 1–2 px/estilo de link.

## 6. Acessibilidade

- `<html lang="pt-BR">`: ok.
- Palco `aria-live="polite"`: ok; o palco declarado usa `aria-live="assertive"`.
- Conversa `role="log"`: ok.
- Aviso de morte súbita `role="alert"`: implementado na T24, com tudo o que é decorativo em `aria-hidden`.
- Gavetas: foco entra, Esc fecha e o foco volta (Radix; conferido na T13 e no navegador).
- Lighthouse (snapshot, mesa desktop): **Acessibilidade 100**, Melhores práticas 100. Aparece o aviso `label-content-name-mismatch` nos camarotes porque o `aria-label` da especificação (“Ver camarote de Ana, 7 moedas”) não contém o texto visível “Ana 7 ···”. O rótulo é o da prancha e foi mantido; se o produto preferir zerar o aviso, o nome acessível precisa começar pelo texto visível.
- Áreas de toque e contraste: conferidos nas telas; os botões de papel medem 48 px+ e o contraste dos tokens passa nos dois temas.

## 7. Arquitetura

- As telas de `features/` leem `room` cru em 3 pontos, fora dos seletores: `sheets/ChatSheet.tsx` (`chat.length`), `sheets/RoomMenu.tsx` (`id`, `turn`) e `sheets/SceneLog.tsx` (`turn`). **Reavaliado:** são leituras primitivas, sem regra duplicada; um seletor só para `turn` não pagaria o custo. Mantido.
- Seletores do Zustand devolvem valores primitivos/referências (`me`, `room`, `log`), sem objeto novo por render; não precisam de `useShallow`.
- `App.tsx` escolhe a cena por `selectScene({ me, room, log })`; `?cena=&fixture=` só existe quando `import.meta.env.DEV` e não há ticket.

## 8. Protocolo (conferido contra o servidor)

Conferência de `web/src/net/actions.ts` com `game/structs/validatePlayerPlayPayload.go` e os handlers:

- `ASSASSIN`, `KAMIKAZE`, `CLAIRVOYANT`, `GUARDIAN`, `TRICKSTER`, `GRAVEDIGGER` e `CROUPIER` mandam os campos que o `DecodeStrictJSON` exige, com os nomes certos (`targetPlayer`, `targetCardIndex`, `targetAllyCardIndex`, `useOnTwoCards`).
- `CONTEST_PENALTY` manda `TargetCardIndex` com T maiúsculo, como no cliente antigo; o `encoding/json` do Go casa o campo `targetCardIndex` sem diferenciar caixa.
- `TRILLIONAIRE`, `POLITICAL`, `REBEL`, `INVESTOR` e `SELFISH` vão com `payload: null`; o servidor decodifica `null` como struct vazia e deriva o alvo do evento (`protocols/selfish.go`), então não falta campo.
- `CONTEST` e `COMMAND_*` vão com `{}`.
- `CHAT_MESSAGE` só é enviado com texto não vazio (`ChatSheet` faz `trim` e desabilita o botão).

Nenhum ajuste necessário no `net/`.

## 9. Correções da auditoria de lógica

Achados confirmados no servidor e corrigidos nesta passada (cada um com teste):

- `selectScene`/`selectFootlights` usam `startTime`: o servidor manda `gameEvent: null` entre deixas, e a tela não volta mais para a sala no meio da partida.
- `applyRoom` e `reset` descartam diário, fila e gavetas ao trocar de sala; `GameSocket` define `me` pelo ticket e limpa o estado ao montar.
- `play()` lança antes de enviar se faltar `targetPlayer`/`targetCardIndex`.
- Guardiã é oferecida na morte causada por você mesmo (Greed), como `guardianValidation.go` permite.
- Egoísta não é oferecido em `STEAL_COINS` com `lostCoins/2 <= 1`.
- `CONTEST_PENALTY` sem `CONTEST` (reconexão) fecha a declaração já com o desfecho.
- `CHANGED_CARD` com `cause: CONTEST_PENALTY` não é narrado como Crupiê.
- `selectYou.timer.total` usa a janela do evento (10/6/8 s; 5/3/4 na morte súbita).
- JWT lido como UTF-8 (nome com acento não vira mojibake).
- `ContestPick` mostra “protegida” na carta protegida (antes lia um campo inexistente e dizia “na mão”).

## 10. Pedidos ao servidor (não alterar aqui)

- `players/playerConversion.go` não preenche `moveTime`, `investments` nem `pendingInvestmentCoin` no jogador público. Sem `moveTime`, a ordem de jogo (e o anfitrião) cai no id; investimentos nunca aparecem.
- A versão pública das cartas não manda `name` nem para cartas reveladas; o `REVEALED_CARD` não vira face na tela dos outros.
- `PublicRoomForUpdates` não manda `maxPlayers`; a sala mostra sempre 6 lugares.
- `GameEvent` não tem tags JSON (chega `PlayerID`, `Type`, `ExpiresAt`, `Payload`); o cliente normaliza em `net/protocol.ts`.
- `PublicRoomForUpdates` manda `pendingEffects` (o cliente ignora) e não manda `PendingPresentationEvents` (interno, correto).
- **Guardiã na reação:** a prancha `Main` desenha “Bloquear com Guardiã” sobre uma declaração de Assassino, mas `guardianValidation.go` recusa Guardiã fora de `WAITING_FIRST_ACTION`, `CARD_KILLED` e `CARD_PLAYED_CLAIRVOYANT`. Ou o servidor passa a aceitar a proteção “futura” na declaração, ou a prancha perde o botão. **Confirmar com o produto.**
- **Discrepância de regra, confirmar com o produto:** o servidor credita `LostCoins / 2` ao Trapaceiro (`stealCoinsEffect.go`), então “rouba 3” vira 1; a prancha `Acoes` e `pricing.ts` dizem “fica com 2”. `pricing.ts` ficou como a prancha.

## 11. Decisões inferidas (não desenhadas)

- Fluxo de alvo da T21 (camarotes elegíveis com anel, demais a 40%, “☜ Voltar às ações”).
- Modal de login (visual legado, não está nas pranchas).
- Mensagens de erro de sala e de conexão (`copy.roomNotFound`, `copy.roomFull`, `copy.invalidCode`, `copy.connectionLost`).
- Primeiro nome no efeito da declaração em vez de “Ela”, para não errar o gênero de atores desconhecidos.
- `waitPickHint` genérico (“a jogada dela”) em vez do papel (“o Assassino dela”).
- Fila de apresentação: o `App` consome a cabeça em `ExpiresAt`; os painéis continuam lendo o `gameEvent` da sala (o servidor apresenta um evento por vez). Assim a fila não cresce e a janela de ação do próximo evento não é comida.
- Registro da morte súbita por sala (`uiStore.suddenDeathShown`), para tocar uma vez por partida mesmo com reconexão.

## 12. Integração (T25)

- Entrada: login Google → `Lobby` (`create`/`join`) → `sessionStore.ticket` → `<GameSocket>`; o `me` sai do JWT do ticket.
- Convite `?sala=K7Q2`: o `App` pede o ticket assim que houver login e não houver socket; sem login, o `Lobby` pré-preenche o código. Sair da sala limpa `?sala=`, `?cena=` e `?fixture=` em todos os pontos de saída.
- Reconexão: o socket usa o mesmo ticket e, ao reaplicar a mesma sala, `applyRoom` não duplica diário nem fila (teste em `gameStore.test.ts`).
- Fila de apresentação coberta por teste (`gameStore.test.ts` → ordem, reaplicação e evento vencido) e consumida no `App`.
- Partida local no gameServer não foi executada nesta máquina (servidor Go/Redis parados); a ligação foi validada com as fixtures e o fluxo HTTP/WS está coberto pelos testes de `net/`.
