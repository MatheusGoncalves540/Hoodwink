# PLANO 10: Botão de login Google com a skin de botões
Status: pendente
Objetivo: Trocar o iframe branco/azul do Google por um botão principal `Ticket`, estilizado pela skin `botoes`, mantendo o login oficial do Google por baixo.
Fora do escopo: trocar o fluxo OAuth; mexer no backend de autenticação; criar peça ou categoria de skin; redesenhar o cartaz de login; login de visitante.
Premissas e decisões: Clientes: não se aplica (repo sem `task clientes`; módulos tocados: client/web) · decisão do usuário: usar o `Ticket` principal (2a) · o iframe oficial do Google fica invisível por cima do botão visual, com `opacity: 0`, para preservar o clique e o fluxo do GIS · o visual usa `.ticket`, que as quatro peças de `botoes` já estilizam · `applySkin` escreve em `document.documentElement` (`main.tsx`), então o portal do Radix herda `data-sk-botoes` · o "G" entra como SVG monocromático em `currentColor`, sem cores fixas · o botão visual fica `aria-hidden` e fora da ordem de foco; o iframe é o controle acessível · o anel de foco aparece no contêiner via `:focus-within` · o contêiner do GIS é irmão do `Ticket`, nunca filho, e leva um comentário de bloco com esse contrato · enquanto o GIS não carrega, só aparece `copy.loginModal.loading`; erro e formulário de nome continuam iguais · Revisado em 2026-11-07: 8 correções aplicadas.
Pré-requisitos: `client/web` com `npm install`; referência de UI: `client/tech-docs/travas-de-design.md`, `client/plans/03-skins.md` (categoria `botoes`) e o cartaz de login existente.

## Tarefas
- [ ] T1 [Frontend] Criar botão `Ticket` com G e contêiner do iframe irmão; ligar o `renderButton` com largura e registrar o contrato · arquivos: `client/web/src/componentsTSX/googleAuth.tsx`, `client/web/src/componentsTSX/styles/loginModal.css`, `client/web/src/game/copy.ts`, `client/web/src/componentsTSX/googleAuth.test.tsx` · depende: — · verificar: `cd client/web && npm test -- googleAuth`
- [ ] T2 [Frontend] Cobrir estados de carregando, pronto, erro e formulário de nome; sem `ready` não renderizar `Ticket` nem overlay, e o iframe ocupa 100% do contêiner · arquivos: `client/web/src/componentsTSX/googleAuth.tsx`, `client/web/src/componentsTSX/styles/loginModal.css`, `client/web/src/componentsTSX/googleAuth.test.tsx`, `client/web/src/game/copy.ts` · depende: T1 · verificar: `cd client/web && npm test -- googleAuth`
- [ ] T3 [QA] Provar a classe `ticket`, o rótulo acessível, o estado de carregando e a ausência de peça nova no catálogo · arquivos: `client/web/src/componentsTSX/googleAuth.test.tsx`, `client/web/src/skins/catalog.test.ts` · depende: T2 · verificar: `cd client/web && npm test -- googleAuth catalog`

## Aceite final
- [ ] `cd client/web && npm test` verde.
- [ ] No cartaz de login, o botão do Google usa `class="ticket"` e o catálogo não ganha peça nova (`npm test -- googleAuth catalog`).
- [ ] O logo `G` e o texto usam a cor do botão, sem fundo branco ou azul do Google.
- [ ] O botão visível não recebe `Tab`; o foco vai para o iframe oficial e desenha o anel no contêiner.
- [ ] `cd client/web && npm run build` verde.
- [ ] No `npm run dev`, trocar a peça de `botoes` no Camarim muda o botão de login e o clique nos 4 cantos abre o Google.

## Riscos
- O iframe de 40 px ficar menor ou mais estreito que o `Ticket`: passar `width` ao `renderButton`, escalar a altura (`scaleY`) e forçar `width/height: 100%` no iframe do overlay.
- O Google bloquear o botão invisível: manter `opacity: 0` e `pointer-events: auto`, nunca `display: none`.
- A skin trocar a fonte e o "G" ficar desalinhado: usar `currentColor` e centralizar por flex.
- O leitor de tela anunciar dois botões: `aria-hidden` e `tabIndex={-1}` no visual.
- A diretriz de marca do Google pede botão visível; a decisão do usuário é manter o iframe invisível sob o `Ticket` para preservar o visual da skin, e isso deve ser reavaliado se o login for revisado pelo Google.

Esforço: M
