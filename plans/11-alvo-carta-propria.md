# PLANO 11: Nunca oferecer as próprias cartas como alvo
Status: pendente
Objetivo: Impedir que a interface abra as cartas do próprio jogador ao mirar outra pessoa, cobrindo Matar, Assassino, Clarividente, Kamikaze e a escolha da carta que cai na contestação.
Fora do escopo: mudar a validação do servidor; redesenhar o fluxo de contestação; alterar as regras de alvo; criar novos painéis.
Premissas e decisões: Clientes: não se aplica (repo sem `task clientes`; módulos tocados: client/web) · decisões do usuário: não há certeza se foi em Matar ou em Assassino/Clarividente/Kamikaze (7a) · o servidor já recusa alvo próprio em `ValidateContestPenaltyProtocol` e nas validações de carta; o furo é de UI · a trava vale para todo alvo de carta: camarote do próprio jogador e jogador fora de cena não abrem a escolha · `selectDecision` devolve `none` quando o estado de contestação aponta o mesmo jogador como quem escolhe e como alvo · `ContestPick` devolve `null` se o alvo da escolha for `view.me` · o resto do fluxo `pickCardToKill` e `waitPick` continua igual.
Pré-requisitos: `client/web` com `npm install`; fixtures `ContestaBlefe` e `ContestaVerdade`; referência de UI: `client/tech-docs/travas-de-design.md` (seção 7).

## Tarefas
- [ ] T1 [Frontend] Travar alvo próprio no fluxo de carta e desabilitar camarote fora de cena durante o alvo · arquivos: `client/web/src/features/table/Table.tsx`, `client/web/src/features/turn/ContestPick.tsx`, `client/web/src/game/selectors.ts`, `client/web/src/features/table/Seat.tsx`, `client/web/src/features/table/SeatRing.tsx` · depende: — · verificar: `cd client/web && npm test -- Table Turn selectors`
- [ ] T2 [QA] Cobrir as quatro combinações de `hasCard` e os alvos Matar, Assassino, Clarividente e Kamikaze, provando que o alvo nunca é `me` · arquivos: `client/web/src/features/table/Table.test.tsx`, `client/web/src/features/turn/Turn.test.tsx`, `client/web/src/game/selectors.test.ts` · depende: T1 · verificar: `cd client/web && npm test -- Table Turn selectors`

## Aceite final
- [ ] `cd client/web && npm test` verde.
- [ ] Em `ContestaBlefe` e `ContestaVerdade`, o jogador que escolhe a carta nunca vê as próprias cartas como opção (`npm test -- Turn`).
- [ ] Em Matar, Assassino, Clarividente e Kamikaze, clicar em um camarote fora de cena não abre a escolha de carta (`npm test -- Table`).
- [ ] Um estado de contestação em que quem escolhe e o alvo são o mesmo jogador não renderiza `pickCardToKill` (`npm test -- selectors`).

## Riscos
- A trava esconder um caso legítimo de escolher a própria carta: nenhum papel atual permite esse alvo; o servidor também recusaria.
- O camarote fora de cena perder o anel de alvo: ele continua visível, só deixa de ser clicável.
- O `ContestPick` ficar em branco com estado inconsistente: mostrar o aviso de estado inválido em vez de `null` silencioso.

Esforço: P
