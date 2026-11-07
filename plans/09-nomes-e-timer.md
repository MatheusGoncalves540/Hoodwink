# PLANO 09: Nomes completos e timer sincronizado na mesa
Status: pendente
Objetivo: Mostrar o nick completo sempre que couber e fazer o número e a barra do tempo da jogada derivarem do mesmo relógio, ressincronizando ao voltar de outra aba.
Fora do escopo: mudar a abreviação no perfil e no Fim da partida; refazer o timer do replay; alterar as janelas de tempo do servidor; redesenhar a faixa "Você".
Premissas e decisões: Clientes: não se aplica (repo sem `task clientes`; módulos tocados: client/web) · decisões do usuário: camarotes, palco e diário usam o nick completo; encolher a fonte e permitir 2 linhas antes de abreviar (6a) · o resto do app continua como está · a abreviação vira último recurso do CSS (`text-overflow: ellipsis`), não uma regra de dados · o timer atualiza a cada 250 ms e recalcula em `visibilitychange` e `focus` · número usa `Math.ceil` e barra usa a fração exata restante, ambos do mesmo `now` · a barra passa a ser `transform: scaleX(...)` inline, sem keyframe de duração, para não divergir após alt+tab · `ContestPick` reaproveita o mesmo hook do timer.
Pré-requisitos: `client/web` com `npm install`; referência de UI: `client/tech-docs/travas-de-design.md` (seção 4) e a prancha "Mesa6".

## Tarefas
- [ ] T1 [Frontend] Trocar `firstName` por nome completo nos camarotes, palco, diário e contestação, com fonte adaptativa e 2 linhas antes de reticências · arquivos: `client/web/src/game/selectors.ts`, `client/web/src/game/log.ts`, `client/web/src/features/table/Seat.tsx`, `client/web/src/features/table/Seat.module.css`, `client/web/src/features/turn/DeclaredCard.tsx`, `client/web/src/features/turn/ContestPick.tsx`, `client/web/src/features/table/Table.test.tsx`, `client/web/src/game/selectors.test.ts` · depende: — · verificar: `cd client/web && npm test -- Table selectors log Turn`
- [ ] T2 [Frontend] Extrair o cálculo restante e o hook do timer, derivar número e barra do mesmo `now` e ressincronizar ao voltar para a aba · arquivos: `client/web/src/ui/Timer/Timer.tsx`, `client/web/src/ui/Timer/Timer.module.css`, `client/web/src/ui/Timer/timer.ts`, `client/web/src/ui/Timer/timer.test.ts`, `client/web/src/features/turn/ContestPick.tsx`, `client/web/src/styles/motion.css` · depende: T1 · verificar: `cd client/web && npm test -- Timer Turn`
- [ ] T3 [QA] Cobrir nickname longo e a regressão do timer com relógio controlado e troca de visibilidade · arquivos: `client/web/src/ui/Timer/timer.test.ts`, `client/web/src/features/table/Table.test.tsx`, `client/web/src/game/selectors.test.ts` · depende: T2 · verificar: `cd client/web && npm test -- Timer Table selectors`

## Aceite final
- [ ] `cd client/web && npm test` verde.
- [ ] Um nick de 20 caracteres aparece inteiro no camarote de 96 px; reticências só quando passa de 2 linhas (`npm test -- Table`).
- [ ] O diário e o palco mostram o nick completo (`npm test -- log Turn`).
- [ ] Com `Date.now` avançando 3 s em uma janela de 10 s, o número mostra 7 e a barra fica em 70% (`npm test -- timer`).

## Riscos
- Nick longo empurrar cartas e moeda para fora do camarote: limitar a 2 linhas e reduzir a fonte em degraus, conferindo 390 px e 10 jogadores.
- A barra inline perder suavidade: atualizar a 250 ms é suficiente para 6 a 10 s e evita animação divergente.
- `ContestPick` e `Timer` divergirem: um único hook e uma única função pura usada pelos dois.

Esforço: M
