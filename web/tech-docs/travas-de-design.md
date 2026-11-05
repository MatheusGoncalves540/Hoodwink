# Travas de design

Regras que o código não pode quebrar sem combinar com o design antes. Cada trava diz onde mora e o que conferir.
Design: canvas "Hoodwink — Refino v2" (pranchas Idle · Mesa, Contornos, Camarim).

## 1. Espaço da arte dos personagens nos botões de ação

Onde: `web/src/features/turn/ActionButton.tsx` e `ActionButton.module.css` (tela "Sua vez", `ActionBoard`).

Cada botão de personagem tem uma **faixa de arte** reservada para a arte final, que vai ser animada quadro a quadro. Hoje ali estão os olhos provisórios (`RoleEyes`).

- Botão: 84 px de altura, borda de 2 px. No mobile de 390 px, o miolo tem cerca de 350 px de largura.
- Faixa de arte: **230 × 76 px**, presa à direita (`right: 44px; top: 1px`). Proporção ~3:1.
- O degradê `.fade` cobre a esquerda para o texto ler bem: opaco até 40% da largura, 35% em 62%, transparente a partir de 80%.
- **Zona segura:** o rosto e o gesto principal ficam nos **~90 px da direita da faixa** (de 62% da largura do botão em diante). O resto da faixa é fundo, que pode ficar sob o degradê.
- O texto ocupa no máximo 64% da largura (`.text`). A moeda de preço (46 px) fica à direita, por cima da faixa: deixe esse canto sem detalhe importante.
- Botão desabilitado: a arte fica em cinza (`grayscale(1) brightness(.55)`). A arte tem que continuar legível assim.

Arquivos de arte: `public/arte/papeis/<ROLE_ID>/acao.webp`, uma tira horizontal com N quadros de 460 × 152 (2x). Toca com a técnica TIRA do idle: `translateX` em `steps(N)`, só no compositor. Com Menos movimento ou `prefers-reduced-motion`, mostra o quadro 1.

Não mudar sem atualizar a arte: altura do botão, largura máxima do texto, posição e tamanho da faixa, o degradê.
As ações básicas (Pegar moeda, Matar, Trocar cartas, das regras v2) usam o mesmo `ActionButton`, com a mesma faixa.

## 2. Holofote e poça de luz são um desenho só

Onde: `web/src/features/table/Spotlight.tsx`, `game/seats.ts` (`conePolygon`).

- O cone sai do alto do palco (centro) e termina na **base do camarote de quem está jogando**, com 6 px de folga de cada lado.
- A poça de luz é uma elipse **centrada nessa mesma base**, com a largura da base do cone. Cone e poça ficam no mesmo elemento, então nunca se separam quando o camarote muda de lugar ou de tamanho.
- Cores: `--hw-cone` e `--hw-poca` (os tokens já existem).
- O cone fica por cima do fundo do camarote, translúcido e sem receber clique (`pointer-events: none`). Hoje ele está em `z-index: 0`, por baixo, e some atrás do camarote: subir.
- Idle: os 3 quadros do conjunto cone + poça trocam juntos (técnica CONTORNO). Nunca anime só um dos dois.

## 3. Letreiro de lâmpadas só no cartaz do lobby

Onde: `.hw-letreiro` em `styles/components.css`. Vai virar a moldura padrão do lugar `lobby` (ver `plans/03-skins.md`).

- Lâmpadas **só na borda**, nunca espalhadas pelo fundo.
- Acendem em sequência, em 3 tempos (3 camadas que alternam a opacidade), o que dá a sensação de luz correndo.
- Elenco, Fim da partida e avisos usam outra moldura (prancha Contornos, opções B a E).

## 4. O texto respira, mas não treme

- Todo texto tem uma respiração leve: sobe 4 px e balança 0,6° para cada lado, em 3,6 s (`ease-in-out`). Blocos vizinhos têm atrasos diferentes (0, −1,2 s, −2,4 s), para não andarem juntos.
- Use as propriedades `translate` e `rotate`, não `transform`. Assim elas somam com o `rotate` de títulos e cartas sem apagar.
- Aplique por bloco de texto (classe `.m-respira` em título, parágrafo, item de lista, botão), nunca por palavra ou `span`. Um bloco dentro de outro não respira de novo.
- Para com Menos movimento, `prefers-reduced-motion` e aba escondida.
- O texto nunca recebe o tremido de rabisco (tira ou contorno). Na moeda, o contorno treme e o número só respira. No timer, o "?" treme e o dígito só respira.
- Orçamento por tela: até 24 tiras e 3 contornos animados. A respiração não entra nessa conta, mas não passe de ~40 blocos animados por tela.

## 5. Como se joga (bastidor)

Módulos da tela, cada um trocável sem mexer no resto:
- `--hw-font-cartao` (padrão: Patrick Hand, com Comic Sans de reserva): letra de tudo que está dentro de um cartão (`.folha`). Os números grandes em `.marcador` continuam em Permanent Marker.
- `--hw-font-rabisco` + `.rabisco`: frase escrita direto no papelão do fundo, sem fita, com uma seta desenhada em 3 quadros. Serve para explicação curta que não cabe em cartão.
- `.q3` (`--q3-ciclo`, padrão 1,5 s): **só o papel do cartão** troca de posição em 3 quadros (`steps`), igual ao traço fervendo. O fundo e a sombra do cartão moram num `::before`, que é o que se mexe. **Texto, placas e rabiscos ficam parados.** Nesta tela não há respiração de texto. Usa `translate`/`rotate`, que somam com a inclinação de cada cartão.
- `.canhoto`: cartão de ingresso (cantos mordidos por máscara parada + picote tracejado em contorno de 3 quadros), igual à opção C da prancha Contornos. Usado no cartão "Usar um personagem".
- A fita (`.fita`) fica só para rótulo curto (ex.: "Modo Clássico", nome do grupo), nunca para frase.
- Deslocamento do `.q3`: no máximo 0,5 px e 0,15°. É para sentir, não para chamar atenção.
- **Chave "Movimento"**: toggle de papelão (`.chave`, `role="switch"`) na barra fixa do topo (`.barra`, `position: sticky`), no desktop e no mobile. Desligada, põe `.pausa` na raiz da tela e corta toda animação dela. No código, grava a preferência **Animações de espera** (ver `plans/04-handoff-refino-v2.md`, fase 3.1), para valer no app todo e não haver duas fontes de verdade.

## 6. Texto para o jogador

Todo texto que o jogador lê passa pela skill `texto-claro` (`.claude/skills/texto-claro/SKILL.md`): linguagem simples, sem jargão do jogo sem explicação, sem cara de texto de IA.
