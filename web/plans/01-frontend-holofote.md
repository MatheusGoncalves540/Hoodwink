# PLANO 01: frontend Holofote a partir do design-handoff
Status: pendente
Objetivo: implementar em `web/` as 23 telas do canvas Hoodwink v62 (mobile/desktop, Matinê/Estreia) seguindo `design-handoff/`, ligadas ao servidor de jogo.
Fora do escopo: mudanças no servidor (`Hoodwink-server`); arte final das cartas e do Proso; regras completas de "Como se joga" além do que está desenhado; Playwright/screenshot automatizado.
Premissas e decisões:
- Fonte: `design-handoff/prompts/00` a `10`; cada tarefa cita o prompt dono. Autoridade visual: captura > prancha > docs.
- Trabalho no branch do worktree `claude/hoodwink-frontend-design-232dc1` (não é main, então não cria `plano/…`).
- `design-handoff/` (16 MB de referência, não versionado) fica fora dos commits.
- Sem `typescript`/`tsc` no projeto e o prompt 01 proíbe dependências além do `vitest`: o gate é `npm run build` + `npm test`.
- Testes: Vitest ao lado do código, Node puro. Teste de fumaça de tela com `react-dom/server` (`renderToStaticMarkup`), sem jsdom nem Testing Library.
- Fidelidade visual: conferência por amostragem no Chrome (DevTools MCP) nas telas-chave; a comparação exaustiva das 88 capturas é do revisor humano (prompt 10).
- Itens não desenhados (fluxo de alvo, mensagens de erro, "Como se joga", confirmação de "Sair de cena") seguem `05-protocolo.md`/docs, marcados como inferidos no relatório.
Pré-requisitos: Node 26 + npm (`npm ci` em `web/`); servidor local em `D:\GitHub\Hoodwink-server\gameServer` (Go 1.26) só para T25; referência de UI: `design-handoff/referencias/capturas/*.png` e `pranchas/*.dc.html`.

## Tarefas

### Fase 1: fundação e dados (prompts 01 e 03)
- [x] T1 [Frontend] copiar estilos e assets, `styles/index.css`, `index.html` (pt-BR, preconnect) e `main.tsx` importando os estilos · arquivos: `web/src/styles/*`, `web/public/assets/*`, `web/index.html`, `web/src/main.tsx` · depende: — · verificar: `npm run build` passa e `tokens.css` só difere do handoff no caminho de `--hw-verso`
- [ ] T2 [Frontend] casca: `App.tsx` com `?cena=&fixture=`, `useLayout`, `useTheme`, `useMotionPref` e a página `?cena=amostras` · arquivos: `web/src/app/*`, `web/src/main.tsx` · depende: T1 · verificar: `npm run build`; `?cena=amostras` mostra tokens e fontes e troca de tema no navegador
- [ ] T3 [Backend] `game/protocol.ts` com schemas Zod (chaves do Go → camelCase) e testes; adiciona `vitest` e scripts `test`/`test:watch` · arquivos: `web/src/game/protocol.ts`, `web/src/game/protocol.test.ts`, `web/package.json` · depende: — · verificar: `npm test`
- [ ] T4 [Backend] `game/roles.ts`, `game/rules.classic.ts` e `game/pricing.ts` com testes de taxa −1..+2, dobras e limites · arquivos: `web/src/game/roles.ts`, `web/src/game/rules.classic.ts`, `web/src/game/pricing.ts`, `web/src/game/pricing.test.ts` · depende: T3 · verificar: `npm test`
- [ ] T5 [Backend] `game/seats.ts` (tabela mobile e parábola desktop) com testes de contagem, sobreposição e limites · arquivos: `web/src/game/seats.ts`, `web/src/game/seats.test.ts` · depende: — · verificar: `npm test`
- [ ] T6 [Backend] `game/copy.ts` e `game/log.ts` (diário) com testes a partir de `05-protocolo.md` · arquivos: `web/src/game/copy.ts`, `web/src/game/log.ts`, `web/src/game/log.test.ts` · depende: T3, T4 · verificar: `npm test`
- [ ] T7 [Backend] `game/fixtures.ts` (estados das pranchas) e `game/selectors.ts` com testes de `selectDecision` (vez, alvo, espectador) e de `Mesa6` · arquivos: `web/src/game/fixtures.ts`, `web/src/game/selectors.ts`, `web/src/game/selectors.test.ts` · depende: T5, T6 · verificar: `npm test`
- [ ] T8 [Backend] `store/gameStore.ts` e `store/uiStore.ts` com `loadFixture` · arquivos: `web/src/store/gameStore.ts`, `web/src/store/uiStore.ts` · depende: T7 · verificar: `npm test` (teste de `loadFixture('Mesa6')` + `applyRoom`)
- [ ] T9 [Backend] `net/GameSocket.tsx` (host de `envs.ts`, reconexão, Zod) e `net/actions.ts` · arquivos: `web/src/net/GameSocket.tsx`, `web/src/net/actions.ts`, `web/src/net/actions.test.ts` · depende: T8 · verificar: `npm test` (mensagens montadas iguais às do `ButtonPanel.jsx`) e `npm run build`

### Fase 2: primitivos (prompt 02)
- [ ] T10 [Frontend] `ui/` Ticket, Button, TextField, Coin, Label, Marquee, ProsoLine, Stamp, Toggle, com seção em amostras · arquivos: `web/src/ui/*` · depende: T2 · verificar: `npm run build`; amostras nos dois temas no navegador
- [ ] T11 [Frontend] `ui/` Timer e CountdownDots (barra só por CSS) · arquivos: `web/src/ui/Timer/*`, `web/src/ui/CountdownDots/*` · depende: T10 · verificar: `npm run build`; amostras
- [ ] T12 [Frontend] `ui/Card` (CardFace, CardBack, CardHidden, CardDead, CardMini) comparado a `Cartas.png` · arquivos: `web/src/ui/Card/*` · depende: T10 · verificar: `npm run build`; amostras vs `Cartas.png`/`CartasEscuro.png`
- [ ] T13 [Frontend] `ui/Sheet` com Radix Dialog (gaveta mobile e painel desktop) · arquivos: `web/src/ui/Sheet/*` · depende: T10 · verificar: `npm run build`; Esc fecha e devolve foco nas amostras

### Fase 3: entrada, sala e mesa (prompts 04 e 05)
- [ ] T14 [Frontend] Entrada (tela 1) mobile e desktop · arquivos: `web/src/features/lobby/*` · depende: T9, T13 · verificar: `npm test` (fumaça com aria-labels) e `?cena=lobby` vs `Lobby.png`/`LobbyDesk.png`
- [ ] T15 [Frontend] Sala de espera (telas 2, 17) · arquivos: `web/src/features/room/*` · depende: T9, T13 · verificar: `npm test` (fumaça `Sala`, `SalaCheia`) e `?cena=room&fixture=Sala` vs `Sala.png`
- [ ] T16 [Frontend] Mesa mobile (telas 3, 7, 13) com slots `stageOverlay`, `bottomPanel`, `sheet`, `scenery` · arquivos: `web/src/features/table/*` · depende: T9, T13 · verificar: `npm test` (fumaça `Mesa3`, `Mesa6`, `Mesa10`) e `?cena=table&fixture=Mesa6` vs `Mesa6.png`
- [ ] T17 [Frontend] Mesa desktop: grade palco + trilho 420 px, camarotes em parábola · arquivos: `web/src/features/table/*` · depende: T16 · verificar: `npm run build` e 1440×900 vs `Mesa6Desk.png`

### Fase 4: gavetas, turno e fim (prompts 06, 07 e 08)
- [ ] T18 [Frontend] SeatSheet e YouSheet (telas 4, 12, 14, 16) · arquivos: `web/src/features/sheets/SeatSheet.tsx`, `web/src/features/sheets/YouSheet.tsx` · depende: T17 · verificar: `npm test` (fumaça `Camarote`, `Voce`) e vs `Camarote.png`
- [ ] T19 [Frontend] RoomMenu e ChatSheet (telas 10, 11, 15), preferências em `localStorage`, texto do chat como texto · arquivos: `web/src/features/sheets/RoomMenu.tsx`, `web/src/features/sheets/ChatSheet.tsx` · depende: T17 · verificar: `npm test` (fumaça `Conversa`, `ConversaCheia`) e vs `Menu.png`
- [ ] T20 [Frontend] SceneLog e DesktopRail, ligando `uiStore.openSheet` no mobile e no desktop · arquivos: `web/src/features/sheets/SceneLog.tsx`, `web/src/features/sheets/DesktopRail.tsx`, `web/src/features/table/Table.tsx` · depende: T18, T19 · verificar: `npm run build` e trilho vs `Mesa6Desk.png`
- [ ] T21 [Frontend] ActionBoard + ActionButton (tela 6) e fluxo de alvo inferido · arquivos: `web/src/features/turn/ActionBoard.tsx`, `web/src/features/turn/ActionButton.tsx` · depende: T17 · verificar: `npm test` (fumaça `Acoes`: preço de `pricing.ts`, Egoísta não clicável) e vs `Acoes.png`
- [ ] T22 [Frontend] ReactionPanel e ContestPick (telas 5, 18, 19) · arquivos: `web/src/features/turn/ReactionPanel.tsx`, `web/src/features/turn/ContestPick.tsx` · depende: T21 · verificar: `npm test` (fumaça `Main`, `ContestaBlefe`, `ContestaVerdade`) e vs `Main.png`
- [ ] T23 [Frontend] OutOfScene e Finale (telas 20, 23) · arquivos: `web/src/features/endgame/OutOfScene.tsx`, `web/src/features/endgame/Finale.tsx` · depende: T17 · verificar: `npm test` (fumaça `ForaDeCena`, `Fim`) e vs `Fim.png`
- [ ] T24 [Frontend] SuddenDeathAlert, SuddenDeathStage e FlameRow portados da prancha (telas 21, 22) · arquivos: `web/src/features/endgame/SuddenDeathAlert.tsx`, `web/src/features/endgame/SuddenDeathStage.tsx`, `web/src/features/endgame/FlameRow.tsx` · depende: T23 · verificar: `npm test` (fumaça `MorteSubita`, `MesaSubita`) e vs `MorteSubita.png`

### Fase 5: integração (prompt 09)
- [ ] T25 [Frontend] `App.tsx` pelo estado real: login → criar/entrar (`?sala=`) → ticket → `GameSocket`; fila de apresentação; reconexão sem duplicar diário · arquivos: `web/src/app/App.tsx`, `web/src/net/*`, `web/src/store/gameStore.ts` · depende: T14, T15, T20, T22, T24 · verificar: `npm test` (reaplicar a mesma sala não duplica o diário) e partida local no servidor
- [ ] T26 [Frontend] remover legado (`componentsJSX/ButtonPanel`, `CommandPanel`, `ChatViewer`, `roomViewer`, `inGame/main.tsx`) mantendo `Auth/` e `authStore.ts` · arquivos: `web/src/componentsJSX/*`, `web/src/inGame/main.tsx` · depende: T25 · verificar: `npm run build` e nenhuma importação restante (`grep -r componentsJSX web/src`)

### Fase 6: revisão (prompt 10)
- [ ] T27 [QA] varredura: cores fora dos tokens, `transition`/`animation` fora de `transform`/`opacity`, texto fora de `copy.ts`, preço fixo; corrigir os de 1 linha e registrar o resto em `plans/01-revisao.md` · arquivos: `web/src/**`, `plans/01-revisao.md` · depende: T26 · verificar: greps do prompt 10 sem ocorrências não justificadas

## Aceite final
- [ ] `npm run build` passa em `web/`
- [ ] `npm test` passa (game/, store/, net/ e fumaça de cada tela)
- [ ] `?cena=…&fixture=…` abre cada uma das 21 telas de jogo em 390×844 e 1440×900, nos dois temas, sem erro no console
- [ ] Movimento reduzido (`prefers-reduced-motion` e menu) desliga cortina e fogo da morte súbita
- [ ] Nenhuma importação de `componentsJSX/ButtonPanel|CommandPanel|ChatViewer|roomViewer` ou `inGame/main` em `web/src`

## Riscos
- Servidor local não sobe ou protocolo difere dos docs → T25 marca [!] e o adaptador fica em `net/`, com pedido ao servidor no relatório
- Volume de telas (88 pranchas) estoura a conferência visual → amostragem nas telas-chave, resto para o revisor humano
- Sem checagem de tipos (sem `tsc`) → testes de `game/` e fumaça cobrem os contratos
- Fluxo de alvo e "Como se joga" são inferidos → marcados no relatório para o dono do produto

Esforço: G
