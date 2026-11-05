# 03 · Skins e Camarim

Status: pronto para implementar. Design: canvas "Hoodwink — Refino v2", pranchas **Camarim** e **Contornos**.
Fora do escopo agora: preço, loja, pagamento, cadeado. Tudo fica liberado. O catálogo só precisa ter onde encaixar isso depois.

## O que é

Todo o visual do jogo vira um conjunto de **peças** trocáveis. O jogador escolhe as peças no **Camarim** (tela nova).
Um **pacote** é uma combinação pronta de peças (uma por categoria). Dá para vestir o pacote inteiro e depois trocar só uma peça.

Categorias da v1:

| Categoria | O que muda | Onde mora no código |
|---|---|---|
| `cores` | todas as cores `--hw-*`, nas versões clara e escura | tokens |
| `letras` | `--hw-font-display`, `--hw-font-rotulo`, `--hw-font-cartao` (cartões do Como se joga) e `--hw-font-rabisco` (escrita no papelão) | tokens + `@font-face` da peça |
| `molduras` | contorno do cartaz do lobby, Elenco, Fim da partida, avisos | `.hw-moldura` |
| `botoes` | botão principal (`.ticket`) e secundário (`.hw-botao2`) | classes de componente |
| `cartas` | verso e tarja da carta: `--hw-verso`, `--hw-carta*` | tokens |

Depois, no mesmo formato: `mesa` (fundo da mesa, holofote, camarotes), `traco` (os desenhos do idle).

Travas que nenhuma skin muda:
- O texto corrido (regras, conversa, histórico) fica em Literata (`--hw-font-texto`). Uma peça de `letras` que tentar mudar isso falha no teste.
- Claro e escuro continuam vindo do aparelho e do toggle. Toda peça de `cores` traz as duas versões.
- A lógica do jogo não sabe que skin existe.

## Regra de performance

O jeito de aplicar uma skin é **um atributo por categoria no elemento raiz**:

```html
<html data-theme="dark" data-sk-cores="bastidor" data-sk-letras="pincel" data-sk-molduras="papelao" data-sk-botoes="placa" data-sk-cartas="papelao">
```

- Componente nenhum lê a skin. Nada de contexto, hook ou `if (skin === ...)` em componente. Eles só usam tokens e as classes de slot.
- O CSS de cada peça é escopado pelo atributo e carregado sob demanda (chunk do Vite). A skin padrão (Holofote) é a que já está em `tokens.css` e `components.css`: não carrega nada a mais.
- Trocar uma peça = carregar um CSS pequeno (uma vez) + mudar um atributo. O navegador recalcula estilo uma vez. Nenhum re-render do React.
- Fonte só baixa quando a peça de `letras` que a usa é vestida.
- Molduras animadas seguem o orçamento do idle (`IdleSistema`): no máximo 3 contornos e 24 tiras animados por tela, pausa com aba escondida, Menos movimento e `prefers-reduced-motion`.

`data-theme` e `data-sk-*` sempre ficam **no mesmo elemento**. Assim a prévia do Camarim funciona igual: o contêiner da prévia recebe os próprios atributos e veste só aquele pedaço.

## Arquivos

```
web/src/skins/
  catalog.ts          dados puros: categorias, peças, pacotes, peça padrão
  equip.ts            funções puras: resolveEquip, equipPack, equipItem
  applySkin.ts        carrega o CSS das peças e escreve os atributos (único lugar que mexe no DOM)
  catalog.test.ts     contrato das peças (ver Testes)
  equip.test.ts
  cores/bastidor.css  uma peça = um arquivo, escopado pelo atributo
  cores/noir.css
  cores/ensaio.css
  letras/pincel.css
  letras/gala.css
  molduras/canhoto.css  molduras/bambinela.css  molduras/cantoneiras.css  molduras/papelao.css
  botoes/placa.css  botoes/carimbo.css  botoes/veludo.css
  cartas/papelao.css  cartas/xadrez.css  cartas/azulejo.css
web/src/features/camarim/
  Camarim.tsx  Camarim.module.css  Camarim.test.tsx
web/src/ui/Moldura/Moldura.tsx (+ .module.css)
```

### catalog.ts

```ts
export const SLOTS = ["cores", "letras", "molduras", "botoes", "cartas"] as const;
export type Slot = (typeof SLOTS)[number];
export type Equip = Record<Slot, string>;

export type Item = {
  id: string;            // "bastidor"
  slot: Slot;
  name: string;          // "Bastidor"
  load?: () => Promise<unknown>; // import("./cores/bastidor.css"); ausente na peça padrão
};
export type Pack = { id: string; name: string; equip: Equip };

export const DEFAULT_EQUIP: Equip = { cores: "holofote", letras: "cartaz", molduras: "holofote", botoes: "ingresso", cartas: "veludo" };
export const ITEMS: Item[] = [ /* uma linha por peça */ ];
export const PACKS: Pack[] = [
  { id: "holofote", name: "Holofote", equip: DEFAULT_EQUIP },
  { id: "bastidor", name: "Bastidor", equip: { cores: "bastidor", letras: "pincel", molduras: "papelao", botoes: "placa", cartas: "papelao" } },
  { id: "noir", name: "Noir", equip: { cores: "noir", letras: "gala", molduras: "cantoneiras", botoes: "carimbo", cartas: "xadrez" } },
  { id: "ensaio", name: "Ensaio", equip: { cores: "ensaio", letras: "cartaz", molduras: "canhoto", botoes: "veludo", cartas: "azulejo" } },
];
/** Hoje tudo é liberado. Quando houver posse, só esta função muda. */
export const isUnlocked = (_item: Item) => true;
```

Os valores de cada peça (cores claras e escuras, fontes, estilos) estão na prancha Camarim, no `renderVals` (`COR`, `LET`, `MOL`, `BOT`, `CAR`).

### equip.ts (puro, sem DOM)

- `resolveEquip(saved: unknown): Equip`: lê o que veio do `localStorage`. JSON quebrado, categoria faltando ou id que não existe no catálogo viram o padrão daquela categoria. É a fronteira de confiança: nunca confie no que está salvo.
- `equipPack(pack): Equip`, `equipItem(equip, item): Equip`: devolvem um objeto novo.

### applySkin.ts

```ts
export async function applySkin(el: HTMLElement, equip: Equip) {
  await Promise.all(SLOTS.map((s) => itemOf(s, equip[s])?.load?.())); // carrega antes de trocar: sem piscar
  for (const s of SLOTS) el.dataset[`sk${cap(s)}`] = equip[s];
}
```

- Boot (`main.tsx`): `await applySkin(document.documentElement, resolveEquip(readPref("skin")))` antes do primeiro render. Os chunks são pequenos; se falhar, segue com o padrão.
- Persistência: `prefs.ts` ganha a chave `skin` (JSON de `Equip`), igual a `theme` e `motion`. Depois, quando houver conta, o servidor devolve o mesmo JSON.

### Formato de uma peça (CSS)

```css
/* cores/bastidor.css */
[data-sk-cores="bastidor"] { --hw-plateia:#E9D6B8; --hw-luz:#C8A27A; --hw-tinta:#1E1512; /* ...todas as cores... */ }
[data-sk-cores="bastidor"][data-theme="dark"] { --hw-plateia:#2C1F16; /* ... */ }
```

```css
/* molduras/canhoto.css */
[data-sk-molduras="canhoto"] .hw-moldura { /* máscara dos cantos, fundo */ }
[data-sk-molduras="canhoto"] .hw-moldura > .hw-moldura-f { /* picote tracejado em 3 quadros (contorno) */ }
[data-sk-molduras="canhoto"] .hw-moldura[data-lugar="lobby"] { /* variação por lugar, se quiser */ }
```

Uma peça só pode: redefinir variáveis `--hw-*`, estilizar as classes de slot (`.hw-moldura*`, `.ticket`, `.hw-botao2`, `.hw-carta`) e apontar para arquivos em `public/skins/<slot>/<id>/`. Sem JS e sem mudar HTML.

## Componentes que mudam

1. **Cores soltas viram token.** Hoje há cor fixa fora dos tokens (ex.: `TONES` e `rgb(30, 21, 18)` em `ActionButton`, `#5A1820` em vários CSS). Cada uma vira um token (`--hw-tom-ouro`, `--hw-tom-vinho`, `--hw-tom-cedula`, `--hw-faixa-fundo`, ...). Sem isso a skin de cores fica pela metade.
2. **`ui/Moldura/Moldura.tsx`** (novo): `<div class="hw-moldura" data-lugar={lugar}><i class="hw-moldura-f"/><i class="hw-moldura-f"/><i class="hw-moldura-f"/>{children}</div>`. Os três `<i>` são os quadros do contorno (técnica CONTORNO do idle). Usar em: cartaz do Lobby (`lobby`), Sala/Elenco (`elenco`), Finale (`fim`), avisos (`aviso`). O `.hw-letreiro` de bolinhas sai; o letreiro vira a moldura padrão do lugar `lobby` (lâmpadas só na borda, correndo em 3 tempos, ver prancha Contornos, opção A).
3. **Botões:** o principal já é `.ticket`. Criar `.hw-botao2` para o secundário e usar nos lugares que hoje estilizam na mão.
4. **Cartas:** `--hw-verso` já existe. Garantir que `Card` e `Seat` só usem `--hw-verso` e `--hw-carta*`.

## Tela Camarim

Prancha **Camarim** (mobile e desktop, claro e escuro).
- Entrada: perfil no Lobby (`☞ Camarim`) e uma linha no Menu da sala, abaixo de Preferências.
- Topo: título, uma frase, prévia (moldura com "Elenco", verso da carta, botão principal) vestida com a escolha atual.
- Abas: Pacotes, Cores, Letras, Molduras, Botões, Cartas. Cada aba tem uma frase dizendo o que muda.
- Grade de peças (2 colunas no mobile, 4 no desktop). Tocar veste na hora e salva. A peça vestida mostra "✓ vestido" e borda de ouro (`aria-pressed`).
- Ao abrir uma aba, pré-carregar o CSS das peças dela (`load()`), para a prévia trocar sem espera.
- Sem preço, sem cadeado, sem "comprar".

## Testes (Vitest, Node puro)

- `equip.test.ts`: `resolveEquip` com `undefined`, `""`, JSON quebrado, objeto vazio, id inexistente, categoria extra e entrada válida (compare o objeto inteiro). `equipPack` e `equipItem` não mutam a entrada.
- `catalog.test.ts`:
  - todo pacote tem as 5 categorias e só ids que existem;
  - todo item com `load` tem o arquivo CSS correspondente;
  - todo seletor do arquivo começa com `[data-sk-<slot>="<id>"]` (nenhuma regra vaza para fora da peça);
  - peça de `cores` define todas as variáveis da lista `COLOR_TOKENS`, no claro e no escuro;
  - contraste: `--hw-tinta` sobre `--hw-plateia` ≥ 4.5:1 e `--hw-titulo` sobre `--hw-plateia` ≥ 3:1, nos dois temas;
  - nenhuma peça redefine `--hw-font-texto`.
- `tokens.test.ts` (guarda): nenhum `*.module.css` tem cor literal (`#xxx`, `rgb(`) fora da lista de exceções. É isso que garante que trocar a skin troca tudo.
- `Camarim.test.tsx`: renderiza as abas, marca a peça vestida, tocar numa peça chama `applySkin` com o `Equip` novo.

## Ordem

1. Tokens: tirar cor solta dos componentes (com o teste de guarda).
2. `catalog.ts` + `equip.ts` + testes.
3. `applySkin` + chave `skin` em `prefs.ts` + boot.
4. `Moldura` e troca do letreiro; `.hw-botao2`.
5. Peças da v1 (CSS), uma por arquivo, com o teste de contrato.
6. Tela Camarim e as duas entradas.
7. `npm test` inteiro verde.
