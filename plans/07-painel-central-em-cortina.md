# PLANO 07: Painel central em cortina para ações e escolha do Clarividente
Status: concluído
Objetivo: No seu turno, as ações, o alvo e a escolha "mostrar para todos" do Clarividente ficam num painel central que abre e fecha como cortina sobre a área dos jogadores, sem cobrir cabeçalho, trilho e faixa "Você".
Fora do escopo: mover reação e contestação para o painel central; mudar o layout do replay; criar peça de skin; redesenhar o `YouBar`; mudar as regras do Clarividente no servidor.
Premissas e decisões: Clientes: não se aplica (repo sem `task clientes`; módulos tocados: client/web) · decisão do usuário: o painel fica no centro, onde ficam os jogadores, e as barras laterais e de baixo são HUD · o painel abre por padrão no seu turno e guarda o estado fechado até o próximo evento · a cortina sobe de baixo para cima e deixa um handle visível quando fechada · o painel expandido mostra uma fileira só de leitura com suas cartas; o `YouSheet` continua sendo o seletor de carta própria · o Clarividente é `target: "other"`, então o `YouSheet` não abre; o Sim/Não vive no `TargetBar` da cortina e nasce em `Não` · a escolha fica visível do clique no papel até confirmar a jogada · durante o alvo a cortina fica aberta e o `YouBar` continua visível como HUD · o holofote é um gradiente CSS com `--hw-poca`/`--hw-luz`; o `<Spotlight>` da mesa e as travas do desenho não mudam · o `SuddenDeathAlert` continua em `screenOverlay` cobrindo a tela inteira; só o `ActionBoard` sai de lá · o estado abre/fecha mora em `TableScene` e reinicia quando o evento da sala muda · textos novos passam pela skill `texto-claro` · Revisado em 2026-11-07: 9 correções aplicadas.
Pré-requisitos: `client/web` com `npm install`; referência de UI: `client/tech-docs/travas-de-design.md`, `client/plans/04-handoff-refino-v2.md` (fase 4) e as pranchas "Idle · Mesa" e "Sua vez".

## Tarefas
- [x] T1 [Frontend] Criar `CenterCurtain` com abre/fecha, handle, foco, Esc, reduced motion e máquina de estado testável · arquivos: `client/web/src/features/turn/CenterCurtain.tsx`, `client/web/src/features/turn/CenterCurtain.module.css`, `client/web/src/features/turn/CenterCurtain.test.tsx` · depende: — · verificar: `cd client/web && npm test -- CenterCurtain`
- [x] T2 [Frontend] Montar o `ActionBoard` num slot `curtain` dentro de `.tableArea`, classe sem tela inteira, `SuddenDeathAlert` em `screenOverlay` e `YouBar` visível no alvo · arquivos: `client/web/src/features/turn/ActionBoard.tsx`, `client/web/src/features/turn/ActionBoard.module.css`, `client/web/src/features/table/Table.tsx`, `client/web/src/features/table/Table.module.css`, `client/web/src/features/turn/ActionBoard.test.tsx` · depende: T1 · verificar: `cd client/web && npm test -- ActionBoard Table`
- [x] T3 [Frontend] Levar o `TargetBar` à cortina e trocar o switch do Clarividente por Sim/Não `aria-pressed`, padrão `Não`, holofote CSS · arquivos: `client/web/src/features/turn/TargetBar.tsx`, `client/web/src/features/turn/TargetBar.module.css`, `client/web/src/features/table/Table.tsx`, `client/web/src/game/copy.ts`, `client/web/src/features/turn/Turn.test.tsx` · depende: T2 · verificar: `cd client/web && npm test -- Turn`
- [x] T4 [QA] Cobrir cortina fechada, cartas no painel, HUD visível, morte súbita e Clarividente até confirmar · arquivos: `client/web/src/features/turn/CenterCurtain.test.tsx`, `client/web/src/features/turn/ActionBoard.test.tsx`, `client/web/src/features/turn/Turn.test.tsx`, `client/web/src/features/table/Table.test.tsx` · depende: T3 · verificar: `cd client/web && npm test -- CenterCurtain ActionBoard Turn Table`

## Aceite final
- [x] `cd client/web && npm test` verde.
- [x] `Table.test.tsx` monta o slot `curtain` dentro de `.tableArea` e confirma `Header`, `YouBar` e `rail` fora dele (`npm test -- Table`).
- [x] Em `390 px` e no desktop, conferir à mão que a cortina não cobre cabeçalho, trilho nem `YouBar`; no alvo o `YouBar` continua visível (`npm run dev`).
- [x] O botão da cortina alterna `aria-expanded`/`aria-controls` e, fechado, deixa só o handle focável (`npm test -- CenterCurtain`).
- [x] Com Clarividente, `Não` vem selecionado e `Sim`/`Não` ficam visíveis até a jogada; o payload leva o valor escolhido (`npm test -- Turn`).
- [x] Com `footlights.suddenDeath`, o `SuddenDeathAlert` continua cobrindo a tela inteira (`npm test -- Table`).

## Riscos
- A altura do `YouBar` varia com `svh` e a cortina pode encostar: usar a mesma clamp em `--hw-youbar-h` e conferir 390 px e desktop.
- A cortina cobre camarotes baixos: o handle e o estado fechado precisam manter os alvos clicáveis, inclusive com 10 jogadores.
- Foco preso no painel fechado: usar `inert` e `visibility: hidden` no conteúdo fechado, `aria-controls` no handle e devolver o foco ao fechar com Esc.
- O `ActionBoard` sair de `screenOverlay` e levar o `SuddenDeathAlert` junto: manter o aviso separado e coberto por teste.
- A trava do holofote da mesa ser confundida com o destaque da cortina: o gradiente CSS usa só tokens e não altera `Spotlight.tsx`/`seats.ts`.

Esforço: G
