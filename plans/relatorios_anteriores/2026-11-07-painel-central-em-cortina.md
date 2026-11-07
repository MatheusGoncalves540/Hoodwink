# Relatório — 07-painel-central-em-cortina

Plano: `plans/07-painel-central-em-cortina.md` · branch `plano/07-painel-central-em-cortina` · status: concluído

## Concluído
- T1: `CenterCurtain` (handle, `aria-expanded`/`aria-controls`, `inert` + `visibility`, Esc devolve o foco) e `curtainState` como máquina pura.
- T2: slot `curtain` dentro de `.tableArea`, recortado abaixo do cabeçalho e acima de `--hw-youbar-h`; `ActionBoard` com fileira só de leitura das suas cartas; `SuddenDeathAlert` segue em `screenOverlay`.
- T3: `TargetBar` na cortina, Clarividente em Sim/Não `aria-pressed` (nasce em Não) e holofote em gradiente de tokens.
- T4: testes de cortina aberta/fechada, cartas no painel, HUD, aviso de tela inteira e Clarividente até o payload.

## Validação
- `cd client/web && npm test` verde (373 testes); `npm run build` ok.
- Conferência visual no `npm run dev`, `?cena=table&fixture=Acoes`: 390×844 e 1440×900; Esc fecha, handle recebe foco e o corpo fica `inert` (conferido no DOM).

## Ressalvas
- Vitest roda em Node com `renderToStaticMarkup`: nada de interação real; o que é estado foi extraído para `curtainState` e o resto é markup.
- `rail` só existe no desktop e o SSR cai em mobile; a conferência do trilho foi visual, no navegador.
- Zustand v5 no SSR usa `getInitialState`, então `setState` não aparece no SSR: o teste do Clarividente monta `Table` + `CenterCurtain` em vez de `TableScene`.
- `client/web/dist` é versionado; o `npm run build` foi restaurado para não misturar artefatos no commit.

## Dívida
- O handle fechado encosta na ribalta (BARALHO/TAXA) no mobile: não cobre valores, mas a posição definitiva depende da prancha (`client/web/src/features/turn/CenterCurtain.module.css:52`). Verificar com `?cena=table&fixture=Acoes` em 390 px.
