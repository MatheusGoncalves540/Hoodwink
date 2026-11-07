# PLANO 08: Carta revelada do Clarividente fica no "o que a mesa já viu"
Status: concluído
Objetivo: A carta revelada para todos entra no histórico do jogador revelado e aparece na gaveta dele com o turno, sem transformar a revelação em estado permanente da carta.
Fora do escopo: mostrar a revelação privada na gaveta do alvo; mudar o servidor para manter a carta aberta; refazer o replay; alterar o palco do evento `REVEALED_CARD`.
Premissas e decisões: Clientes: não se aplica (repo sem `task clientes`; módulos tocados: client/web) · decisões do usuário: a memória vale só quando "mostrar para todos" (3a) · a revelação privada fica só no histórico de quem jogou o Clarividente · na gaveta do alvo entra um selo "Carta revelada no turno N" e uma linha no "o que a mesa já viu" (4a) · o estado da carta continua escondido; a memória é do diário, não da carta · o evento já chega ao cliente com `targetPlayer`, `targetCardIndex`, `revealedCard` e `showToAllPlayers` · nenhuma migração ou mudança de protocolo.
Pré-requisitos: `client/web` com `npm install`; referência de UI: `client/tech-docs/travas-de-design.md` e a prancha "Camarote"; fixture `Mesa6` para o caso de teste.

## Tarefas
- [x] T1 [Frontend] Guardar no `LogEntry` a carta, o alvo, o índice e se a revelação foi para todos · arquivos: `client/web/src/game/log.ts`, `client/web/src/game/log.test.ts` · depende: — · verificar: `cd client/web && npm test -- log`
- [x] T2 [Frontend] Expor as cartas reveladas para todos no `selectSeatSheet` e acrescentar a linha derivada no diário do alvo · arquivos: `client/web/src/game/selectors.ts`, `client/web/src/game/selectors.test.ts` · depende: T1 · verificar: `cd client/web && npm test -- selectors`
- [x] T3 [Frontend] Renderizar o bloco "Carta revelada no turno N" e a linha no "o que a mesa já viu" na gaveta do camarote · arquivos: `client/web/src/features/sheets/SeatSheet.tsx`, `client/web/src/features/sheets/SeatSheet.module.css`, `client/web/src/features/sheets/Sheets.test.tsx`, `client/web/src/game/copy.ts` · depende: T2 · verificar: `cd client/web && npm test -- Sheets`
- [x] T4 [QA] Provar que a revelação pública aparece no alvo e a privada não aparece no alvo · arquivos: `client/web/src/features/sheets/Sheets.test.tsx`, `client/web/src/game/selectors.test.ts`, `client/web/src/game/log.test.ts` · depende: T3 · verificar: `cd client/web && npm test -- Sheets selectors log`

## Aceite final
- [x] `cd client/web && npm test` verde.
- [x] Em `Mesa6`, revelar a carta de Bruno para todos e abrir o camarote dele mostra o selo e a linha no "o que a mesa já viu" (`npm test -- Sheets`).
- [x] A mesma revelação com `showToAllPlayers: false` não acrescenta nada à gaveta do alvo.
- [x] A carta continua com o estado visual de escondida no `SheetCard` fora do bloco de memória.

## Riscos
- Duas revelações da mesma carta duplicarem o selo: agrupar por alvo e índice, mantendo o turno mais recente.
- O diário derivado sair fora de ordem: ordenar por turno decrescente junto com `selectLog`.
- A revelação privada vazar para outros clientes: ela já chega só ao jogador do Clarividente; o filtro `showToAllPlayers` é a segunda barreira.

Esforço: M
