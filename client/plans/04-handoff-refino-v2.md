# 04 · Handoff completo: Refino v2

Para o agente de código. Este arquivo é o índice: diz **o que fazer, em que ordem, onde e como saber que terminou**. Os detalhes moram nos documentos ligados aqui; quando houver conflito, o documento específico vence.

Fontes:
- Design: canvas "Hoodwink — Refino v2" (https://claude.ai/artifact/C2ezeZouotdf8xpNZTNU6a). Valores exatos (cores, medidas, textos) estão nas pranchas.
- `plans/02-regras-v2.md`: regras novas do jogo (front + servidor).
- `plans/03-skins.md`: skins e a tela Camarim.
- `tech-docs/travas-de-design.md`: regras de design que o código não pode quebrar. **Leia antes de começar.**
- Skill `texto-claro` (`.claude/skills/texto-claro/SKILL.md`): todo texto que o jogador lê.

Regras do repositório (AGENTS.md): Vitest (`npm test`), teste ao lado do código, regra de jogo em função pura, "pronto" só com a suíte inteira verde.

---

## Pranchas do canvas → telas do código

| Prancha | Tela / componente |
|---|---|
| Como se joga · desktop, Como se joga · mobile | tela nova `features/howto/` (fase 7) |
| Entrada · Proso em placeholder | `features/lobby/Lobby.tsx` |
| Idle · Lobby, Sala, Fim (mobile) | `Lobby.tsx`, `room/WaitingRoom.tsx` + `Cast.tsx`, `endgame/Finale.tsx` |
| Idle · Mesa, Sua vez, Menu (mobile) | `table/*`, `turn/ActionBoard.tsx` + `ActionButton.tsx`, `sheets/RoomMenu.tsx` |
| Idle · Mesa desktop | `table/*` no layout desktop + `sheets/DesktopRail.tsx` |
| Idle · sistema e componentes | catálogo: como cada componente se mexe (fase 3) |
| Contornos · 5 opções | `ui/Moldura` (fase 5) |
| Camarim | `features/camarim/` (fase 8, ver 03-skins) |
| Estudo · traço fervendo | só referência, não vira tela |

---

## Fase 0 · Preparar

1. `cd web && npm install && npm test`: tem que começar verde (hoje: 20 arquivos, 114 testes).
2. A skill `texto-claro` está em `.claude/skills/`, mas o `.gitignore` ignora `.claude` inteiro. Para ela entrar no repositório, troque a linha `.claude` por:
   ```
   .claude/*
   !.claude/skills/
   ```

## Fase 1 · Tokens e letras

Arquivos: `styles/tokens.css`, `styles/components.css`, CSS modules.

1. **Texto corrido em Literata.** Hoje `--hw-font-texto` é Bodoni Moda. Troque para `"Literata", Georgia, serif` e inclua Literata (ital, opsz, 400..700) no `@import` do Google Fonts. Títulos continuam em Bodoni/Limelight.
2. **Cores soltas viram token** (passo 1 do `03-skins.md`). Inclui `TONES` e `rgb(30, 21, 18)` do `ActionButton`. Crie o teste de guarda `tokens.test.ts`.
3. Tokens novos: `--hw-font-cartao` (Patrick Hand → Comic Sans MS → cursive), `--hw-font-rabisco` (Permanent Marker). `--hw-poca` já existe.

Pronto quando: nenhuma cor literal nos `*.module.css` fora da lista de exceções; telas iguais às de antes, exceto pela fonte do texto.

## Fase 2 · Regras v2

Tudo em `plans/02-regras-v2.md`, seção 8 (front) e seção 6 (mensagens). Resumo do que muda no front:
- `game/rules.classic.ts`: `maxPlayers: 10`, `minTax: -5`, `maxTax: 5` (hoje está 6, −3 e 10).
- `game/pricing.ts` + testes: fórmulas novas (Trapaceiro fica com a metade maior, Guardiã com 3 preços, Egoísta com limiar `2 + t`, Investidor novo).
- Ações do turno no `ActionBoard`: grupo "Ações do turno" no topo, com Pegar moeda, Matar e Trocar cartas, usando o **mesmo `ActionButton`** (com a faixa de arte, ver trava 1). Mensagens novas em `net/actions.ts`.
- Mesa com até 10 jogadores: `seats.ts` já cobre até 9 adversários; testar a 390 px.

**Depende do servidor.** O servidor já tem `TAKE_COINS`, `KILL_CARD` e `SWAP_CARDS` (nomes oficiais, ver 02 seção 6). O que falta do lado de lá está em `Hoodwink-server/plans/01-regras-v2-servidor.md`; não ligue o Matar no front antes da fórmula nova (mínimo 3) estar no servidor.

## Fase 3 · Sistema de movimento

Arquivos: `styles/motion.css`, `app/prefs.ts`, `app/useMotionPref.ts`, `features/sheets/RoomMenu.tsx`, `features/lobby/Lobby.tsx`, `game/copy.ts`.

### 3.1 Duas chaves nas configurações (obrigatório)

Duas preferências separadas, nas Preferências do Menu da sala e nas configurações do Lobby. Prancha: Idle · Mesa → Menu.

| Chave | O que desliga | Padrão |
|---|---|---|
| **Animações** | **todas**: entradas, gavetas, carimbos, viradas de carta, cortina, timer animado e também as de espera | ligada |
| **Animações de espera** | só o idle: tira (traço tremendo), contorno, letreiro, respiração do texto, 3 quadros e canhoto do Como se joga | ligada |

Regra entre as duas:
- **Animações** desligada ⇒ **Animações de espera** desliga junto. A chave de espera aparece desligada e desabilitada (`aria-disabled`), com a dica "Desligada junto com Animações".
- Religar **Animações** devolve a chave de espera ao valor que ela tinha antes (as duas preferências são guardadas separadas).
- `prefers-reduced-motion` do sistema vale como **Animações** desligada, salvo se o jogador ligar a chave de propósito.

Implementação, sem nada por componente:
- `prefs.ts`: chaves `motion` (já existe; `"reduzido"` = Animações desligada) e `idle` (nova; `"off"` = espera desligada).
- Uma função pura em `app/motion.ts`: `motionState({ motion, idle, systemReduced }) => { all: boolean, idle: boolean }`, com `idle = all && idlePref`. Testes: as 8 combinações e o caso de religar.
- Um único lugar escreve na raiz: `data-motion="reduced"` quando `all` é falso e `data-idle="off"` quando `idle` é falso. Componente nenhum lê essas preferências.
- `motion.css`: `[data-motion="reduced"] *{animation:none!important;transition:none!important}` e `[data-idle="off"] :is(.m-tira,.m-contorno,.m-letreiro,.m-respira,.m-q3,.m-canhoto) ...{animation:none}`. Na tira, o desenho 1 fica parado; no contorno, a primeira camada fica visível.
- Textos em `copy.ts`: "Animações", dica "Todas, inclusive as transições."; "Animações de espera", dica "Desenhos tremendo, texto respirando, lâmpadas.". Passe pela skill `texto-claro`. Hoje a chave se chama "Menos movimento": renomeie e atualize os testes que conferem esse texto.

### 3.2 Classes de idle em `motion.css`

Técnicas e valores: nota "HANDOFF · sistema de idle" do canvas, prancha "Idle · sistema e componentes" e travas 3, 4 e 5.
- `.m-tira`: glifo de tamanho fixo. 3 desenhos lado a lado, `translateX(-100%)` em `steps(3)`, 1,5 s (ambiente) ou 0,75 s (sua vez).
- `.m-contorno` com 3 SVG filhos: caixa que estica, opacidade em degraus com atrasos 0, −⅓ e −⅔.
- `.m-letreiro`: lâmpadas só na borda, 3 camadas acendendo em sequência (0,9 s). Só no cartaz do lobby.
- `.m-respira`: texto. Sobe 4 px e balança 0,6° em 3,6 s, `ease-in-out`, usando `translate`/`rotate` (não `transform`). Atrasos 0, −1,2 s e −2,4 s por bloco. **Fora do Como se joga.**
- `.m-q3`: só no Como se joga, e **só no papel do cartão** (fundo + sombra num pseudo-elemento). O texto do cartão fica parado.
- `.m-canhoto`: picote tracejado em contorno de 3 quadros, mais cantos mordidos (máscara parada). No Como se joga, no cartão "Usar um personagem".
- Pausa automática: aba escondida (`document.hidden` → `data-idle="off"` temporário) e, na sua vez, o ambiente para e só o timer e o botão de ação tremem (`data-deixa` na raiz da mesa).
- Orçamento por tela: até 24 tiras, 3 contornos e uns 40 blocos respirando.

Pronto quando: as duas chaves funcionam como na tabela, o teste de `motionState` passa e nenhuma animação de espera roda com a aba escondida.

## Fase 4 · Mesa e Sua vez

- `table/Spotlight.tsx`: cone e poça num elemento só, a poça centrada na base do camarote (trava 2). Suba para cima do camarote, translúcido, com `pointer-events: none`. Cone e poça tremem juntos (contorno de 3 quadros), e só a ponta de cima mexe.
- `turn/ActionButton`: a faixa de arte 230×76 fica como está, é trava (1). Comentário já aponta para o doc.
- `ui/Timer`: o "?" de "restam 7?" vira `.m-tira` rápida em `--hw-alerta`. O número não treme.
- Camarotes, moedas e versos de carta: ver o mapa da prancha de sistema (versos e moedas tremem; nome e número só respiram).
- A Mesa em si (posições, arco, tamanhos) já está certa no código. As pranchas foram ajustadas a ela.

## Fase 5 · Lobby, Sala e Fim

- `Lobby.tsx:89`: troque `<img src="/assets/proso.webp">` pelo placeholder "me troque :c" (moldura tracejada com contorno + "?" em tira). Remova `--hw-proso-filtro` e `--hw-proso-mistura` de `tokens.css`.
- `ui/Moldura/Moldura.tsx` novo (formato em `03-skins.md`). Lugar `lobby` = letreiro de lâmpadas na borda. O `.hw-letreiro` de bolinhas sai.
- Elenco (Sala) e Fim da partida: usar `Moldura` com `data-lugar="elenco"` e `"fim"`. **A moldura padrão ainda não foi escolhida** (prancha Contornos, opções B a E). Até a escolha, use B (filete fervendo).
- "Fim da partida": já feito em `copy.ts` e no teste.

## Fase 6 · Texto

- Rode a skill `texto-claro` em `game/copy.ts`. Hoje há 14 ocorrências de "ator(es)", "em cena" e "contest…". Atualize os testes que conferem texto no mesmo commit.
- O tema de teatro fica nos títulos curtos e na fala do Proso. Regras e instruções usam palavras comuns.

## Fase 7 · Tela Como se joga

Pranchas "Como se joga · desktop" e "· mobile". É uma tela nova em `features/howto/` (um componente, layout por `useLayout`), aberta pelos links `☞ Como se joga` de `Lobby.tsx:140` e `RoomMenu.tsx:40`.
- **Visual de bastidor**: fundo de papelão, cartões de papel (`.folha`) com durex, placas recortadas com títulos em Permanent Marker. Tokens de papelão só desta tela (nota "HANDOFF · papelão").
- **Letra dos cartões**: `--hw-font-cartao` (Patrick Hand). Os números grandes ficam em `--hw-font-rabisco`.
- **Rabisco**: frases curtas escritas direto no papelão, com seta desenhada em tira. Fita só para rótulo curto (trava 5).
- **Movimento**: só o papel dos cartões se mexe (`.m-q3` no pseudo-elemento). Texto, placas e rabiscos ficam parados; as setas e o "?" da moldura de arte continuam tremendo (são desenho). Nada de respiração aqui.
- **Cartão "Usar um personagem"**: estilo Canhoto (`.m-canhoto`), igual à opção C da prancha Contornos.
- **Chave "Movimento"** na barra fixa do topo (`position: sticky`), em papelão, desktop e mobile. Ela escreve a preferência **Animações de espera** (a mesma da fase 3), não uma preferência própria.
- **Conteúdo**: textos das pranchas (já revisados em linguagem simples). Os preços vêm de `pricing.ts` com a taxa do controle "Teste outra taxa" (estado local, não muda a partida).
- **Personagens**: lista em etiquetas + ficha do selecionado. A moldura de arte é 4:5, com placeholder "arte em breve!" e caminho `public/arte/papeis/<id>/` para os quadros futuros.
- Teste: renderiza as seções, troca de personagem, os preços mudam com a taxa, a chave alterna `aria-checked` e grava a preferência.

## Fase 8 · Skins e Camarim

Tudo em `plans/03-skins.md`. Pré-requisito: fase 1 (tokens) e fase 5 (`Moldura`). Nada de preço ou cadeado.

---

## Ordem e dependências

```
0 → 1 → 3 → 4 → 5 → 6 → 7 → 8
          2 (quando o servidor aceitar as mensagens novas; pode correr em paralelo a partir da 1)
```

Faça um commit por fase, com `npm test` verde em cada um.

## Decisões em aberto (perguntar, não decidir sozinho)

1. Moldura do Elenco e do Fim da partida: B, C, D ou E (prancha Contornos). Padrão provisório: B.
2. D1 a D9 das regras v2 (`02-regras-v2.md`, seção 5). A D7 (taxa aleatória) o servidor já respondeu: desligada.
3. Servidor: `maxPlayers` 6 → 10 e mínimos da Guardiã conforme o PDF.

## Não fazer

- Não animar texto com o tremido de rabisco (tira/contorno). Fora do Como se joga, texto só respira; no Como se joga, texto fica parado.
- Não ler skin nem preferência de movimento dentro de componente. Tudo passa pelos atributos na raiz.
- Não usar `transform` na respiração e nos 3 quadros (apaga a inclinação dos cartões). Use `translate`/`rotate`.
- Não usar `filter` ou `feTurbulence` animados.
- Não mudar a faixa de arte do `ActionButton` sem atualizar a trava 1.

---

## Status

Status: concluído (com premissas abaixo, à espera do design)

- [x] Fase 0 · Preparar (skill `texto-claro` copiada para `.claude/skills/`, `.gitignore` liberado)
- [x] Fase 1 · Tokens e letras
- [x] Fase 2 · Regras v2 (preços, ações do turno, histórico com os nomes reais do servidor)
- [x] Fase 3 · Sistema de movimento
- [x] Fase 4 · Mesa e Sua vez (holofote, `?` do timer, `data-deixa`)
- [x] Fase 5 · Lobby, Sala e Fim (`Moldura`, Proso em placeholder)
- [x] Fase 6 · Texto
- [x] Fase 7 · Como se joga
- [x] Fase 8 · Skins e Camarim

## Premissas (decididas sem o canvas, que não estava acessível)

- Canvas "Refino v2" não abriu para o agente. Valores das peças de skin (cores, fontes, molduras, botões, versos), textos do Como se joga e o mapa de tremor de camarotes/moedas/versos foram feitos de forma conservadora; trocar pelos da prancha.
- Fase 4: camarotes, moedas e versos ainda não tremem (falta o mapa da prancha de sistema).
- Moldura do Elenco e do Fim: opção B (filete fervendo), provisória (decisão em aberto 1).
- Pegar moeda com taxa negativa dá 0 (como o servidor); D2 continua em aberto.
- Egoísta barrando investimento: `5 + t` com desconto 5, como o servidor (o 02 não dá o preço).
- Ações do turno usam o tom `cedula` na faixa de arte.
- Lobby: as duas chaves de movimento ficam num `<details>` "Preferências" (não havia configurações no Lobby).
- Peças de `cores` não mexem nas cores da carta (`--hw-carta*` é da categoria `cartas`).
- Molduras e botões de skin valem só no contêiner `data-sk-*` mais próximo, para as prévias do Camarim não herdarem a peça vestida.
- Sem typecheck no projeto (`typescript` e `@types/react` não estão instalados); a verificação foi `npm test` + `vite build`.
