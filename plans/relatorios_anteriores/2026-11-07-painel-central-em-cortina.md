# Relatório — 07-painel-central-em-cortina

Plano: `plans/07-painel-central-em-cortina.md` · branch `plano/07-painel-central-em-cortina` · status: concluído

## Concluído
- T1: `CenterCurtain` (handle, `aria-expanded`/`aria-controls`, `inert` + `visibility`, Esc devolve o foco) e `curtainState` como máquina pura.
- T2: slot `curtain` dentro de `.tableArea`; `ActionBoard` com fileira só de leitura das suas cartas; `SuddenDeathAlert` segue em `screenOverlay`.
- T3: `TargetBar` na cortina e a escolha do Clarividente.
- T4: testes de cortina aberta/fechada, cartas no painel, HUD, aviso de tela inteira e Clarividente até o payload.
- Revisão do dono: gaveta com o handle como borda de cima (sobe junto, sem gap), fim na ribalta; `ActionBoard` sem o bloco "Rodada/Taxa/Sua vez".
- Revisão do dono: Clarividente em duas etapas — "Revelar para" (Todos / Apenas você) no cone de luz e depois a mesa dos alvos com poças; `TargetBar` mostra o que foi decidido.
- Idle forte (`m-respira-forte`) nos textos da escolha e do alvo.

## Validação
- `cd client/web && npm test` verde (375 testes).
- Conferência no `npm run dev`, `?cena=table&fixture=Acoes`: 390×844 e 1440×900 — gaveta aberta/fechada, Esc com foco no handle e corpo `inert`, cone + cartão, poças nos elegíveis e fluxo até a carta.

## Ressalvas
- Vitest roda em Node com `renderToStaticMarkup`: nada de interação real; o estado foi extraído para `curtainState` e o resto é markup.
- `rail` só existe no desktop e o SSR cai em mobile; o trilho foi conferido no navegador.
- Zustand v5 no SSR usa `getInitialState`: o teste monta `Table` + `CenterCurtain` em vez de `TableScene`.
- `client/web/dist` é versionado; o `npm run build` foi restaurado para não misturar artefatos.

## Dívida
- Com a gaveta terminando na ribalta, a lista de ações mostra ~2 botões por tela no mobile (`ActionBoard.module.css`, `.list`). Revisar se a fileira de cartas deve encolher no painel.
- `vite.config.js`, `package.json`, `Lobby.*` e os arquivos de `server/` apareceram alterados no working tree por fora desta tarefa (trabalho paralelo); não entraram nos commits.
