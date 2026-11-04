# Revisão 01: varredura do prompt 10 (T27)

Status: **em andamento**. A varredura automática está feita; a revisão tela a tela depende das telas da T21–T24, que ainda não existem. Reexecutar os itens marcados como pendentes quando elas entrarem.

Método: greps do prompt 10 em `web/src`, leitura dos CSS Modules e conferência das telas já prontas contra `design-handoff/referencias/capturas/`. Gates rodados neste estado: `npm test` (99 testes, 15 arquivos) e `npm run build` verdes. Sem Playwright (opcional no prompt).

## 1. Cores fora dos tokens

Comando: `grep -rnE "#[0-9a-fA-F]{3,8}|rgb\(" web/src`, descontando `web/src/styles/` (cópia do handoff, fonte dos tokens) e as cores fixas permitidas em `02-design-system.md`.

| Local | Ocorrência | Classificação |
| --- | --- | --- |
| `web/src/features/table/Seat.module.css:30-31` | `#5A1820` e `#7A2430` nas listras do veludo do camarote | **defeito**: usar `var(--hw-veludo)` e derivar a listra clara com `color-mix` (correção de 1 linha por regra) |
| `web/src/componentsTSX/styles/loginModal.css` | `#3b82f6`, `#2563eb`, `#666`, `#333`, `#e5e7eb`, `#d1d5db`, `rgba(0,0,0,.5)` | **inferido**: o modal de login não está nas pranchas; precisa de decisão do produto antes de restilizar |
| `web/src/ui/Card/Card.module.css:66` e `web/src/ui/Coin/Coin.module.css:9` | `#1e1512` | **aceitável**: cor fixa prevista em `02-design-system.md` (“texto das moedas”) |
| `web/src/main.css` | várias cores do legado | **removido**: arquivo não era importado por ninguém |

## 2. Movimento (`transform`/`opacity`)

Nenhum `transition` ou `animation` de componente anima propriedade fora de `transform`/`opacity`.

- `web/src/ui/Sheet/Sheet.module.css:36` (`rail`): anima `transform`.
- `web/src/componentsTSX/styles/loginModal.css` (`overlayShow`, `contentShow`): anima `opacity` e `transform`.
- `web/src/styles/motion.css` e `web/src/styles/sudden-death.css` (cópias do handoff): só `transform`/`opacity`, com a exceção de `ms-some` (`visibility`), que é a cortina da morte súbita.
- `ui/Timer` usa `transform: scaleX(var(--de))` via `m-timer`; correto.
- Movimento reduzido: regras em `styles/index.css` cobrem `data-motion="reduced"`; conferido no navegador nas telas prontas.

## 3. Texto fora de `game/copy.ts`

- 5 `aria-label` literais em `features/` (cada um pede uma chave nova em `copy.ts` + uma linha no componente):
  - `web/src/features/lobby/Lobby.tsx:84` — “Cartaz”
  - `web/src/features/lobby/Lobby.tsx:114` — “Jogar”
  - `web/src/features/room/Cast.tsx:19` — “Elenco”
  - `web/src/features/room/Invite.tsx:37` — “Convite”
  - `web/src/features/room/WaitingRoom.tsx:54` — “Começar”
- `web/src/ui/`: nenhum texto literal.
- `web/src/app/Amostras.tsx`: textos literais, mas é a página de amostras de desenvolvimento (`?cena=amostras`), fora das 21 telas. **Aceitável**.
- `web/src/app/App.tsx:60` (“Login”) e `web/src/componentsTSX/loginModal.tsx` (“Login”, “Fechar”, descrição): legado do login; o `App` sai na T25 e o modal continua. Revisar junto com a T25.

## 4. Preço fixo fora de `pricing.ts`

Nenhuma ocorrência. `copy.roleDesc` e `game/selectors.ts` recebem `Price` de `game/pricing.ts`; não há número de preço escrito em tela. Revisar de novo quando a T21 (painel de ações) entrar.

## 5. Revisão tela a tela

Conferidas contra as capturas nas tarefas correspondentes:

| Tela | Capturas |
| --- | --- |
| Entrada (T14) | `Lobby`, `LobbyDesk`, `LobbyDeskEscuro` |
| Sala de espera (T15) | `Sala`, `SalaDesk`, `SalaCheia` |
| Mesa mobile (T16) | `Mesa6`, `Mesa10` |
| Mesa desktop (T17) | `Mesa6Desk` |
| Camarote e Você (T18) | `Camarote`, `Voce` |
| Menu e conversa (T19) | `Menu` |
| Trilho (T20) | `Mesa6Desk` |

Pendentes (tela ainda não construída):

- T21 — `Acoes`
- T22 — `Main`, `ContestaBlefe`, `ContestaVerdade`
- T23 — `ForaDeCena`, `Fim`
- T24 — `MorteSubita`, `MesaSubita`

## 6. Acessibilidade

- `<html lang="pt-BR">`: ok (`web/index.html`).
- Palco `aria-live="polite"`: ok (`web/src/features/table/Stage.tsx:11`).
- Conversa `role="log"`: ok (`web/src/features/sheets/ChatSheet.tsx:20`).
- Aviso de morte súbita `role="alert"`: a implementar na T24.
- Gavetas: foco entra, Esc fecha e o foco volta (Radix; verificado na T13 no navegador).
- Áreas de toque e contraste: conferidos nas telas prontas; revisar T21–T24.
- Lighthouse de acessibilidade: não rodado nesta passada (depende das telas finais).

## 7. Arquitetura

- As telas de `features/` leem `room` cru em 3 pontos, fora dos seletores: `sheets/ChatSheet.tsx:48` (`chat.length`), `sheets/RoomMenu.tsx:78-79` (`id`, `turn`) e `sheets/SceneLog.tsx:10` (`turn`). Trocar por seletores em `game/selectors.ts` quando essas gavetas forem revistas.
- Seletores do Zustand devolvem valores primitivos/referências (`me`, `room`, `log`), sem objeto novo por render; não precisam de `useShallow` no estado atual.

## 8. Protocolo (conferido contra o servidor)

Conferência de `web/src/net/actions.ts` com `game/structs/validatePlayerPlayPayload.go` e os handlers:

- `ASSASSIN`, `KAMIKAZE`, `CLAIRVOYANT`, `GUARDIAN`, `TRICKSTER`, `GRAVEDIGGER` e `CROUPIER` mandam os campos que o `DecodeStrictJSON` exige, com os nomes certos (`targetPlayer`, `targetCardIndex`, `targetAllyCardIndex`, `useOnTwoCards`).
- `CONTEST_PENALTY` manda `TargetCardIndex` com T maiúsculo, como no cliente antigo; o `encoding/json` do Go casa o campo `targetCardIndex` sem diferenciar caixa.
- `TRILLIONAIRE`, `POLITICAL`, `REBEL`, `INVESTOR` e `SELFISH` vão com `payload: null`; o servidor decodifica `null` como struct vazia e deriva o alvo do evento (`protocols/selfish.go`), então não falta campo.
- `CONTEST` e `COMMAND_*` vão com `{}`.
- `CHAT_MESSAGE` só é enviado com texto não vazio (`ChatSheet` faz `trim` e desabilita o botão).

Nenhum ajuste necessário no `net/`; nada a pedir ao servidor aqui.

## 9. Correções da auditoria de lógica

Achados confirmados no servidor e corrigidos nesta passada (cada um com teste):

- `selectScene`/`selectFootlights` passam a usar `startTime` (`protocol.ts`, `fixtures.ts`): o servidor manda `gameEvent: null` entre deixas, e a tela não volta mais para a sala no meio da partida.
- `applyRoom` e `reset` descartam diário, fila e gavetas ao trocar de sala; `GameSocket` define `me` pelo ticket e limpa o estado ao montar.
- `play()` lança antes de enviar se faltar `targetPlayer`/`targetCardIndex` (o servidor fecha o socket ou age na carta 0 quando o campo some).
- Guardiã é oferecida na morte causada por você mesmo (Greed), como `guardianValidation.go` permite.
- Egoísta não é oferecido em `STEAL_COINS` com `lostCoins/2 <= 1` (`selfishValidation.go`).
- `CONTEST_PENALTY` sem `CONTEST` (reconexão) fecha a declaração já com o desfecho.
- `CHANGED_CARD` com `cause: CONTEST_PENALTY` não é mais narrado como Crupiê.
- `selectYou.timer.total` usa a janela do evento: ação 10 s, mensagem 6 s, importante 8 s (3/4/5 na morte súbita), conforme `ClassicRules.yaml`.
- JWT lido como UTF-8 (nome com acento não vira mojibake).

## 10. Pedidos ao servidor (não alterar aqui)

- `players/playerConversion.go` não preenche `moveTime`, `investments` nem `pendingInvestmentCoin` no jogador público. Sem `moveTime`, a ordem de jogo (e o anfitrião) cai no id; investimentos nunca aparecem.
- A versão pública das cartas não manda `name` nem para cartas reveladas; o `REVEALED_CARD` não vira face na tela dos outros (o cliente pode cachear o evento, mas o certo é o servidor mandar).
- `PublicRoomForUpdates` não manda `maxPlayers`; a sala mostra sempre 6 lugares.
- `GameEvent` não tem tags JSON (chega `PlayerID`, `Type`, `ExpiresAt`, `Payload`); o cliente normaliza em `net/protocol.ts`. O ideal é camelCase.
- `PublicRoomForUpdates` manda `pendingEffects` (o cliente ignora) e não manda `PendingPresentationEvents` (interno, correto).
- **Discrepância de regra, confirmar com o produto:** o servidor credita `LostCoins / 2` ao Trapaceiro (`stealCoinsEffect.go`), então “rouba 3” vira 1; a prancha `Acoes` e `pricing.ts` dizem “fica com 2”. `pricing.ts` ficou como a prancha; não troquei por uma suposição.

## 11. Decisões inferidas (não desenhadas)

- Fluxo de alvo da T21 (camarotes elegíveis com anel, demais a 40%, “☜ Voltar às ações”), conforme `05-protocolo.md`.
- Modal de login (visual legado, não está nas pranchas).
- Mensagens de erro de sala e de conexão (`copy.roomNotFound`, `copy.roomFull`, `copy.invalidCode`, `copy.connectionLost`).
