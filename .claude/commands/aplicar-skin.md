---
description: Transforma a ficha preenchida do artista (PDF, foto ou valores) em uma skin do jogo
argument-hint: <caminho da ficha/PDF/anexo> [--slot=cores|letras|molduras|botoes|cartas] [--id=<slug>] [--commit]
---
Integre a skin do artista: $ARGUMENTS

Você é o integrador de skins. A ficha do artista é uma **descrição**, não código. Ela vem com sobras, faltas e imprecisões de propósito. Seu trabalho é normalizá-la para o contrato do jogo, dizer o que foi adaptado e nunca quebrar as travas de `client/web/src/skins/catalog.test.ts`.

Não anuncie etapas. Fale no final.

## 0. Ler a ficha e montar a tabela de entrada

> Se a ficha vier em PDF escaneado ou foto, leia como imagem/PDF. Se algum campo continuar ilegível, peça a transcrição só dele — não invente valor.

1. Identifique o slot: `cores`, `letras`, `molduras`, `botoes` ou `cartas`. Se vier mais de uma categoria, processe uma por vez; um pacote novo só se o usuário pedir.
2. Transcreva TUDO para uma tabela antes de tocar em código:

   `Campo | Veio? | Valor do artista | Destino | Observação`

   Use as fichas geradas por `tools/skins-pdf/build.mjs` como esquema canônico:

   | Slot | Campos esperados |
   |---|---|
   | cores | nome, id, intenção, 15 tokens × claro/escuro |
   | letras | nome, id, intenção, display/rotulo/cartao/rabisco + pesos + origem |
   | molduras | nome, id, lugares, contorno topo/base e laterais, espessura/raio, ornamento, cores, 3 quadros, desenho |
   | botoes | nome, id, intenção, principal e secundário (fundo, texto, borda/anel, raio, sombra, tipografia), desenho |
   | cartas | nome, id, intenção, desenho do verso, 5 tokens, medidas/observações |

3. Marque cada campo como `ok`, `faltou`, `sobra`, `conflito` ou `ilegível`.
4. Só pergunte se um campo `ilegível` ou `conflito` mudar materialmente o resultado. Máximo 3 perguntas fechadas, recomendada primeiro. O resto vira premissa registrada.
5. Normalize o id: `^[a-z][a-z0-9]*$` (minúsculo, sem espaço/hífen). Nome em português, capitalizado.
6. Confira se o id já existe **no mesmo slot** (`veludo` existe em `botoes` e em `cartas`, isso é normal). Id ocupado: sufixe (`veludo2`) ou pergunte.

## 1. Regras por categoria (o contrato do teste)

### cores — `client/web/src/skins/cores/<id>.css`

Duas regras, exatamente com estes seletores:

```css
[data-sk-cores="<id>"]{--hw-plateia:…;--hw-luz:…;--hw-tinta:…;--hw-tinta-suave:…;--hw-titulo:…;--hw-veludo:…;--hw-sobre-veludo:…;--hw-alerta:…;--hw-ouro:…;--hw-ouro-texto:…;--hw-lampada:…;--hw-filete:…;--hw-poca:…;--hw-cone:…;--hw-scrim:…}
[data-sk-cores="<id>"][data-theme="dark"]{…os mesmos 15…}
```

- **Os 15 tokens, nem mais nem menos, nos dois blocos.** O teste compara as chaves com `COLOR_TOKENS` e falha com token extra.
- Os 12 tokens sólidos (`plateia`, `luz`, `tinta`, `tinta-suave`, `titulo`, `veludo`, `sobre-veludo`, `alerta`, `ouro`, `ouro-texto`, `lampada`, `filete`) precisam ser HEX; só `poca`, `cone` e `scrim` aceitam `rgba()`. O teste de contraste lê HEX; `rgba()` em tinta/plateia quebra a conta.
- `--hw-tinta` sobre `--hw-plateia` ≥ **4,5:1**; `--hw-titulo` sobre `--hw-plateia` ≥ **3:1**, no claro e no escuro. Se o artista entregou algo ilegível, escureça/clareie o mínimo e registre em "Ajustado".
- `--hw-poca`, `--hw-cone`, `--hw-scrim` são véus translúcidos: aceite `rgba()`. Não os troque por HEX opaco sem motivo.
- Artista deu paleta sem os 15? Derive os que faltam pelo **papel** do token e pela paleta dada (ex.: `luz` = `plateia` um passo mais claro; `tinta-suave` = `tinta` com 60–70% de mistura na `plateia`; `sobre-veludo` = contraste alto contra `veludo`). Liste cada derivado.
- Artista deu cor extra (uma 16ª)? Não entra no bloco. Se ela for essencial, ela só pode virar regra visual adicional **depois** dos dois blocos, com um seletor descendente (ex.: `[data-sk-cores="<id>"] .algum-elemento`), usando os 15 tokens; senão, descarte e reporte.

### letras — `client/web/src/skins/letras/<id>.css`

```css
@import url("https://fonts.googleapis.com/css2?family=…&display=swap");
/* Peça de letras "<id>": … */
[data-sk-letras="<id>"]{--hw-font-display:…;--hw-font-rotulo:…;--hw-font-cartao:…;--hw-font-rabisco:…}
```

- Só `display`, `rotulo`, `cartao` e `rabisco`. **Nunca** defina `--hw-font-texto` (o teste falha; o texto corrido é travado em Literata).
- Papel que o artista não definiu: **não escreva** a variável; ela herda a base (Cartaz). Não copie o valor da base para dentro da peça.
- Google Fonts: um único `@import` no topo, com os pesos realmente usados.
- Fonte fora do Google (arquivo, Adobe, licença): **não** coloque `@font-face` no arquivo da peça — o teste de escopo captura `@font-face` como seletor e reprova. Crie/estenda `client/web/src/styles/fonts.css` (importado por `index.css`), salve o arquivo em `client/web/public/assets/fonts/<id>/` e na peça só referencie a família. Registre a licença no relatório.
- Sempre inclua fallback coerente na família: `"Fonte", Georgia, serif` / `system-ui, sans-serif` / `cursive` etc.
- O teste também exige que todo seletor comece com `[data-sk-letras="<id>"]`. `@import` é removido antes da checagem; `@font-face`, não.

### molduras — `client/web/src/skins/molduras/<id>.css`

- Todo seletor começa com `[data-sk-molduras="<id>"]`.
- Toda regra que toca uma `.hw-moldura` precisa do guard de aninhamento, exatamente neste formato (o teste procura esta string):

```css
[data-sk-molduras="<id>"] .hw-moldura:not([data-sk-molduras="<id>"] [data-sk-molduras]:not([data-sk-molduras="<id>"]) .hw-moldura)
```

- O componente já põe três `<i class="hw-moldura-f m-f">` antes do conteúdo. Mantenha as três camadas. O desenho é o `::before` (base) + `.hw-moldura-f` (quadros).
- Traduza o desenho nesta ordem: `border`/`border-radius` → `outline`/`box-shadow` → `background` (gradientes) → `mask` (cantos mordidos) → SVG em asset. Não mexa em padding, largura, altura ou display do conteúdo.
- Cores: sempre tokens (`var(--hw-ouro)`, `var(--hw-veludo)`, `var(--hw-filete)`, `var(--hw-papelao-escuro)`…). Não invente HEX solto.
- 3 quadros: base + `:nth-of-type(2)` e `:nth-of-type(3)` com deslocamento/rotação/raio levemente diferentes. Se o artista desenhou parado, use o mesmo desenho nos três (base apenas já passa).
- Lobby é especial: o padrão usa `.hw-moldura[data-lugar="lobby"]::before` como letreiro. Se a peça não distingue lugar, ela vale para todos; não crie regra só para lobby sem pedido.

### botoes — `client/web/src/skins/botoes/<id>.css`

- Todo seletor começa com `[data-sk-botoes="<id>"]` e usa o guard equivalente ao de molduras:

```css
[data-sk-botoes="<id>"] .ticket:not([data-sk-botoes="<id>"] [data-sk-botoes]:not([data-sk-botoes="<id>"]) .ticket)
[data-sk-botoes="<id>"] .hw-botao2:not([data-sk-botoes="<id>"] [data-sk-botoes]:not([data-sk-botoes="<id>"]) .hw-botao2)
```

- Principal = `.ticket`; secundário = `.hw-botao2`. Se o artista só desenhou o principal, **não invente** o secundário: ele fica com o padrão.
- O `.ticket` base usa `!important` em fundo/cor/borda. Siga o padrão das peças existentes: `background:…!important;color:…!important` quando for sobrescrever.
- Não mexa em altura, largura, padding ou `font-size`: o componente manda no tamanho.
- Tipografia: use `var(--hw-font-…)`; não importe fonte aqui (isso é do slot `letras`).
- Cores por token. Sombra dura com `box-shadow`, anel com `box-shadow`/`border`, inclinação com `rotate`.

### cartas — `client/web/src/skins/cartas/<id>.css`

```css
[data-sk-cartas="<id>"]{--hw-verso:…;--hw-carta:…;--hw-carta-tinta:…;--hw-carta-rotulo:…;--hw-escondida-borda:…}
```

- Um bloco só: a peça de cartas usa as mesmas cores no claro e no escuro. Se o artista entregou duas versões, una com um padrão neutro ou crie SVGs separados e use a versão clara como `--hw-verso` (o verso não muda com o tema nas peças customizadas).
- `--hw-verso` aceita camadas CSS (`repeating-linear-gradient`, `conic-gradient`, `radial-gradient`) terminando em cor de fallback, ou `url(/assets/cartas/verso-<id>.svg) center/100% 100% no-repeat,#cor`.
- Se veio SVG: salve em `client/web/public/assets/cartas/verso-<id>.svg` (ou `verso-<id>-claro.svg`/`-escuro.svg`), proporção 3:4, e referencie com caminho absoluto `/assets/...`.
- Garanta legibilidade de `--hw-carta-tinta` sobre `--hw-carta` (≥ 4,5:1) e de `--hw-carta-rotulo` sobre `--hw-carta` (≥ 3:1).
- Token que faltou: herde o valor do padrão Holofote (`veludo`), mas registre.
- Não altere medidas da carta no CSS da peça (78×110, 132×186 etc. são do componente).

## 2. Sobras e faltas (o coração do comando)

**Faltou**
- `cores`: derive pelo papel do token e checagem de contraste (acima).
- `letras`: herde a base; não declare a variável.
- `molduras`: parte não descrita mantém o desenho base (ex.: só cantoneiras → sem contorno contínuo).
- `botoes`: botão não descrito mantém a base.
- `cartas`: token ausente herda o padrão Holofote.
- 3 quadros ausentes: desenho estático (só a base). Não invente animação.

**Sobrou**
- Elemento que já existe no jogo (ex.: carta morta desenhada de novo): ignore e diga que é estado do componente.
- Efeito visual representável com os tokens: implemente como regra adicional escopada, **depois** dos blocos canônicos, sem novos tokens obrigatórios.
- Efeito que exige layout novo, estado/aba nova, categoria nova ou preço/cadeado: **não force**. Pare e reporte como pedido de feature (aí é `/planejar`).
- Cor/token extra em `cores`: não entra no bloco dos 15 (o teste falha). Descarte ou vire regra adicional com tokens existentes.

**Conflito**
- Viola contraste, trava de fonte, escopo de seletor ou medidas: ajuste o mínimo para passar, mantenha a intenção e registre em "Ajustado".
- Ficha ilegível num ponto que muda o resultado: pergunte (máx. 3).
- Nunca edite `catalog.test.ts` para acomodar a ficha. O teste é o contrato.

## 3. Escrever o código

1. Crie `client/web/src/skins/<slot>/<id>.css` com o cabeçalho no estilo das peças existentes:

   ```css
   /* Peça de <slot> "<id>": <uma linha do desenho>. */
   ```

2. Registre em `client/web/src/skins/catalog.ts` no grupo do slot:

   ```ts
   { id: "<id>", slot: "<slot>", name: "<Nome>", load: () => import("./<slot>/<id>.css") },
   ```

   Sem `load` é a peça padrão; peça nova sempre tem `load`. Não mexa em `DEFAULT_EQUIP`, `PACKS` ou `isUnlocked` sem o usuário pedir.
3. Se o usuário pediu pacote novo, ele só pode usar ids existentes nas 5 categorias e precisa de `equip` completo; senão, não crie.
4. Se veio asset, salve o arquivo em `client/web/public/assets/...` e referencie no CSS pelo caminho absoluto `/assets/...`; confirme que `npm run build` copia.

## 4. Provar

Rode, nesta ordem:

```sh
cd client/web
npx vitest run src/skins
npm test
npm run build
```

- `npx vitest run src/skins` cobre: catálogo, 15 tokens e contraste, escopo dos seletores, guard de aninhamento, arquivo existente e trava de fonte.
- `npm test` é a suíte inteira (não pode regredir).
- `npm run build` prova que o `import()` dinâmico resolve.
- Se falhar, corrija a skin (até 3 tentativas), nunca o teste.
- Se o ambiente tiver navegador, abra `/#camarim`, escolha a peça e confira o claro e o escuro; se não der, diga que a conferência visual ficou pendente.

## 5. Saída

**Skin:** <Nome> (`<slot>/<id>`)
**Arquivo:** `client/web/src/skins/<slot>/<id>.css`
**Catálogo:** linha adicionada em `catalog.ts`
**Ficha → código:** 1 linha por campo relevante (o que virou o quê)
**Faltou (premissas):** N itens, 1 linha cada, com o valor herdado/derivado
**Sobrou (não usado):** N itens, 1 linha cada, e por quê
**Ajustado (travas):** N itens (contraste, fonte, escopo, medida)
**Validação:** comandos rodados e resultado em 1 linha
**Decidir:** perguntas fechadas, se houver (recomendada primeiro)

Regras finais:
- Não commite por padrão. Com `--commit`, um único commit `feat(skins): adiciona <nome>`, sem push.
- Não apague nem altere peças existentes para "abrir espaço".
- Menor diff possível. Sem refatoração, sem dependência nova, sem mudar o componente para caber a ficha.
- Se a ficha pedir algo que o sistema não modela (categoria, estado, preço, layout), pare e explique em 3 linhas; o caminho é `/planejar`.
