/**
 * Pranchas de skins em PDF para o artista (e ficha em branco para devolver).
 *
 * Lê o CSS real de client/web/src, monta HTMLs de prévia e manda o Chrome headless
 * imprimir em A4 na raiz do repositório. Sem dependências.
 *
 * Uso: node tools/skins-pdf/build.mjs
 */
import { execFileSync } from "node:child_process";
import { copyFileSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const HERE = dirname(fileURLToPath(import.meta.url));
const REPO = resolve(HERE, "../..");
const WEB = join(REPO, "client/web");
const SRC = join(WEB, "src");
const BUILD = mkdtempSync(join(tmpdir(), "hoodwink-skins-"));
const read = (p) => readFileSync(join(SRC, p), "utf8");

const CHROME = [
  "C:/Program Files/Google/Chrome/Application/chrome.exe",
  "C:/Program Files (x86)/Google/Chrome/Application/chrome.exe",
  "C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe",
  "C:/Program Files/Microsoft/Edge/Application/msedge.exe",
].find(existsSync);

if (!CHROME) throw new Error("Chrome/Edge não encontrado. Ajuste o caminho em CHROME.");

/* ────────────────────────────── catálogo ────────────────────────────── */

const SLOTS = ["cores", "letras", "molduras", "botoes", "cartas"];
const catalog = read("skins/catalog.ts");

const ITEMS = [...catalog.matchAll(/\{ id: "([^"]+)", slot: "(cores|letras|molduras|botoes|cartas)", name: "([^"]+)"([^}]*)\}/g)].map(
  (m) => ({ id: m[1], slot: m[2], name: m[3], custom: /load/.test(m[4]) }),
);

const DEFAULT_EQUIP = Object.fromEntries(
  SLOTS.map((slot) => [slot, ITEMS.find((i) => i.slot === slot && !i.custom).id]),
);

const PACKS = [...catalog.matchAll(/\{ id: "([^"]+)", name: "([^"]+)", equip: (DEFAULT_EQUIP|\{([^}]*)\}) \}/g)].map((m) => ({
  id: m[1],
  name: m[2],
  equip: m[4]
    ? Object.fromEntries(m[4].split(",").map((p) => p.split(":").map((x) => x.trim().replace(/^"|"$/g, ""))))
    : { ...DEFAULT_EQUIP },
}));

const itemsOf = (slot) => ITEMS.filter((i) => i.slot === slot);

const COLOR_META = [
  ["plateia", "Fundo geral das telas"],
  ["luz", "Superfícies elevadas (painéis, cartões)"],
  ["tinta", "Texto principal e botão primário"],
  ["tinta-suave", "Texto secundário"],
  ["titulo", "Títulos"],
  ["veludo", "Destaque escuro (faixas, letreiro, moldura)"],
  ["sobre-veludo", "Texto sobre o veludo"],
  ["alerta", "Erro, perigo e carta morta"],
  ["ouro", "Filetes e detalhes nobres"],
  ["ouro-texto", "Texto dourado / rótulos"],
  ["lampada", "Lâmpadas acesas e brilho"],
  ["filete", "Linhas e bordas suaves"],
  ["poca", "Poça de luz (véu suave)"],
  ["cone", "Cone de luz (véu)"],
  ["scrim", "Véu escuro sobre o fundo"],
];
const COLOR_TOKENS = COLOR_META.map(([t]) => `--hw-${t}`);

/* ─────────────────────────────── CSS ─────────────────────────────── */

const strip = (s) =>
  s
    .replace(/\/\*[\s\S]*?\*\//g, "")
    .replace(/@import\s+url\([^)]*\)[^;]*;/g, "")
    .replace(/@import\s+[^;]+;/g, "");
const blockList = (src) => [...strip(src).matchAll(/([^{}]+)\{([^{}]*)\}/g)].map((m) => ({ sel: m[1].trim(), body: m[2] }));
const decls = (body) =>
  Object.fromEntries(
    body
      .split(";")
      .filter((d) => d.includes(":"))
      .map((d) => {
        const i = d.indexOf(":");
        return [d.slice(0, i).trim(), d.slice(i + 1).trim()];
      }),
  );
const pickBlock = (src, test) => blockList(src).find((b) => test(b.sel, b.body));

const tokensCss = read("styles/tokens.css");
const baseLight = decls(
  pickBlock(tokensCss, (sel, body) => sel.includes('[data-theme="light"]') && body.includes("--hw-plateia:")).body,
);
const baseDark = decls(
  pickBlock(tokensCss, (sel, body) => sel.includes('[data-theme="dark"]') && body.includes("--hw-plateia:")).body,
);
const baseFonts = decls(pickBlock(tokensCss, (sel) => sel.includes('data-sk-letras="cartaz"')).body);

function coresOf(item) {
  if (!item.custom) return { light: baseLight, dark: baseDark };
  const src = read(`skins/cores/${item.id}.css`);
  const light = pickBlock(src, (sel) => sel === `[data-sk-cores="${item.id}"]`);
  const dark = pickBlock(src, (sel) => sel === `[data-sk-cores="${item.id}"][data-theme="dark"]`);
  return { light: decls(light.body), dark: decls(dark.body) };
}

function fontsOf(item) {
  if (!item.custom) return { base: baseFonts, over: {} };
  const src = read(`skins/letras/${item.id}.css`);
  const over = pickBlock(src, (sel) => sel === `[data-sk-letras="${item.id}"]`);
  return { base: baseFonts, over: over ? decls(over.body) : {} };
}

function cartasOf(item) {
  if (!item.custom) return { light: baseLight, dark: baseDark };
  const src = read(`skins/cartas/${item.id}.css`);
  const block = pickBlock(src, (sel) => sel === `[data-sk-cartas="${item.id}"]`);
  const own = decls(block.body);
  return { light: { ...baseLight, ...own }, dark: { ...baseDark, ...own } };
}

const skinFiles = SLOTS.flatMap((slot) =>
  itemsOf(slot)
    .filter((i) => i.custom)
    .map((i) => read(`skins/${slot}/${i.id}.css`).replaceAll("url(/assets/", "url(./assets/")),
);

/* ───────────────────────────── contraste ───────────────────────────── */

function hexRgb(hex) {
  const h = String(hex).trim();
  const m = /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.exec(h);
  if (!m) return null;
  const v = m[1].length === 3 ? m[1].split("").map((c) => c + c).join("") : m[1];
  return [0, 2, 4].map((i) => parseInt(v.slice(i, i + 2), 16));
}
function luminance(hex) {
  const rgb = hexRgb(hex);
  if (!rgb) return null;
  const [r, g, b] = rgb
    .map((c) => c / 255)
    .map((c) => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}
function contrast(a, b) {
  const [x, y] = [luminance(a), luminance(b)];
  if (x == null || y == null) return null;
  const [hi, lo] = x > y ? [x, y] : [y, x];
  return (hi + 0.05) / (lo + 0.05);
}
const ratio = (n) => (n == null ? "—" : n.toFixed(2).replace(".", ",") + ":1");
const badge = (n, min) => (n == null ? "" : n >= min ? "ok" : "revisar");

/* ─────────────────────────────── HTML ─────────────────────────────── */

const esc = (s) =>
  String(s).replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" })[c]);

const CSS_BASE = `
@page{size:A4;margin:13mm 12mm 15mm}
*{box-sizing:border-box;-webkit-print-color-adjust:exact;print-color-adjust:exact}
html,body{margin:0;padding:0;background:#FFFDF7;color:#1E1512;font-family:var(--hw-font-texto);font-size:11px;line-height:1.5}
body{padding-bottom:4mm}
h1,h2,h3,h4,p,figure,table{margin:0}
ul,ol{margin:0;padding:0;list-style:none}
.p-run{display:flex;justify-content:space-between;margin-top:14px;padding-top:4px;border-top:1px solid #E6DAC2;font-family:var(--hw-font-rotulo);font-size:7.5px;letter-spacing:.16em;text-transform:uppercase;color:#8A7358}
.p-cover{padding:4px 0 12px;border-bottom:2px solid #1E1512;margin-bottom:14px}
.p-kicker{font-family:var(--hw-font-rotulo);font-weight:700;font-size:8.5px;letter-spacing:.22em;text-transform:uppercase;color:#A3201A}
.p-title{font-family:var(--hw-font-display);font-weight:400;font-size:40px;line-height:1;color:#5A1820;margin-top:5px}
.p-sub{font-size:12px;max-width:160mm;margin-top:6px;color:#4E3B2C}
.p-lede{font-size:10px;color:#5A4636;margin-top:5px;max-width:165mm}
.p-h2{font-family:var(--hw-font-display);font-weight:400;font-size:21px;color:#5A1820;margin:0 0 6px;break-after:avoid}
.p-h3{font-family:var(--hw-font-rotulo);font-weight:700;font-size:8px;letter-spacing:.18em;text-transform:uppercase;color:#7A5A1E;margin:8px 0 3px;break-after:avoid}
.p-break{break-before:page}
.p-piece{border:1px solid #E0D3B8;border-radius:8px;padding:9px 11px 11px;margin-bottom:10px;break-inside:avoid}
.p-piece-head{display:flex;align-items:baseline;gap:8px;margin-bottom:7px;border-bottom:1px dashed #DCCDB0;padding-bottom:5px}
.p-piece-head h3{font-family:var(--hw-font-display);font-size:18px;font-weight:400;color:#1E1512}
.p-id{font-family:var(--hw-font-rotulo);font-size:7.5px;letter-spacing:.14em;text-transform:uppercase;color:#8A7358;margin-left:auto}
.p-badge{font-family:var(--hw-font-rotulo);font-weight:700;font-size:7px;letter-spacing:.12em;text-transform:uppercase;color:#7A5A1E;border:1px solid #C9B48C;border-radius:999px;padding:1px 6px}
.p-grid2{display:grid;grid-template-columns:1fr 1fr;gap:9px}
.p-fig{margin:0;min-width:0}
.p-fig figcaption{font-family:var(--hw-font-rotulo);font-size:7.5px;letter-spacing:.14em;text-transform:uppercase;color:#8A7358;margin-top:4px;text-align:center}
.p-table{width:100%;border-collapse:collapse;font-size:9px;break-inside:avoid}
.p-table tr{break-inside:avoid}
.p-table th{font-family:var(--hw-font-rotulo);font-size:7.5px;letter-spacing:.12em;text-transform:uppercase;text-align:left;color:#7A5A1E;border-bottom:1px solid #DCCDB0;padding:3px 5px}
.p-table td{padding:2.5px 5px;border-bottom:1px solid #F0E7D5;vertical-align:middle}
.p-table tr:last-child td{border-bottom:0}
.p-swatch{width:16px;height:16px;border-radius:4px;box-shadow:inset 0 0 0 1px rgba(0,0,0,.2);display:inline-block;vertical-align:middle}
.p-code{font-family:ui-monospace,Consolas,monospace;font-size:8px}
.p-muted{color:#8A7358}
.p-ok{font-family:var(--hw-font-rotulo);font-size:7px;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#3F6B3A}
.p-warn{font-family:var(--hw-font-rotulo);font-size:7px;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#A3201A}
.p-note{font-size:9px;color:#4E3B2C;margin-top:5px}
.p-cols{display:grid;grid-template-columns:1fr 1fr;gap:12px}
.p-cols3{display:grid;grid-template-columns:repeat(3,1fr);gap:9px}
.p-tip{background:#FAF6EC;border:1px solid #E6DAC2;border-radius:6px;padding:7px 9px;font-size:9px}
.p-tip b{display:block;font-family:var(--hw-font-rotulo);font-size:7.5px;letter-spacing:.13em;text-transform:uppercase;color:#A3201A;margin-bottom:2px}
.p-card{background:#FAF6EC;border:1px solid #E6DAC2;border-radius:6px;padding:8px 10px}
.p-list{display:grid;gap:3px;font-size:9.5px}
.p-list li{display:grid;grid-template-columns:4mm 1fr;gap:5px}
.p-list li b{font-family:var(--hw-font-rotulo);font-size:8px;color:#A3201A}
.p-recipe{display:grid;gap:3px}
.p-recipe li{display:grid;grid-template-columns:31mm 1fr;gap:6px;font-size:9px;border-bottom:1px dotted #EADFC9;padding-bottom:2px}
.p-recipe b{font-family:var(--hw-font-rotulo);font-size:7.5px;letter-spacing:.1em;text-transform:uppercase;color:#7A5A1E;padding-top:1px}
.p-strip{display:flex;height:5px;border-radius:3px;overflow:hidden;margin-top:8px}
.p-strip i{flex:1}
.p-form{border:1.5px solid #1E1512;border-radius:8px;padding:11px 13px;break-inside:avoid}
.p-fields{display:grid;grid-template-columns:1fr 1fr;gap:7px 14px;margin:7px 0}
.p-field{min-width:0}
.p-field.wide{grid-column:1/-1}
.p-label{display:block;font-family:var(--hw-font-rotulo);font-size:7.5px;font-weight:700;letter-spacing:.13em;text-transform:uppercase;color:#7A5A1E;margin-bottom:2px}
.p-line{display:block;height:17px;border-bottom:1px solid #C9B48C}
.p-blank{display:block;height:16px;border:1px dashed #C9B48C;border-radius:3px}
.p-canvas{height:38mm;border:1px dashed #C9B48C;border-radius:4px;background-color:#FDF9EF;background-image:linear-gradient(90deg,rgba(201,180,140,.28) 1px,transparent 1px),linear-gradient(rgba(201,180,140,.28) 1px,transparent 1px);background-size:8mm 8mm}
.p-check{display:inline-flex;align-items:center;gap:4px;font-size:9px;margin-right:12px}
.p-check i{width:9px;height:9px;border:1px solid #8A7358;border-radius:2px;display:inline-block}
`;

const CSS_PREVIEW = `
.p-stage{border-radius:8px;padding:11px;display:flex;align-items:center;justify-content:center;gap:10px;min-height:88px;position:relative;overflow:hidden;background:var(--hw-plateia);color:var(--hw-tinta)}
.p-mock{width:100%;max-width:76mm;padding:8px 9px;border-radius:7px;background:var(--hw-plateia);color:var(--hw-tinta);box-shadow:inset 0 0 0 1px var(--hw-filete);display:grid;gap:4px;position:relative;overflow:hidden}
.p-mock-top{display:flex;justify-content:space-between;font-family:var(--hw-font-rotulo);font-size:6.5px;letter-spacing:.13em;text-transform:uppercase;color:var(--hw-ouro-texto)}
.p-mock-banner{margin:0 -9px;padding:3px 9px;background:var(--hw-veludo);color:var(--hw-sobre-veludo);font-family:var(--hw-font-rotulo);font-size:6.5px;font-weight:700;letter-spacing:.16em;text-transform:uppercase}
.p-mock-title{font-family:var(--hw-font-display);font-size:17px;line-height:1;color:var(--hw-titulo)}
.p-mock-text{font-size:8px;color:var(--hw-tinta-suave)}
.p-mock-alert{font-family:var(--hw-font-rotulo);font-size:7px;font-weight:700;letter-spacing:.1em;text-transform:uppercase;color:var(--hw-alerta)}
.p-mock-actions{display:flex;gap:6px;align-items:center;margin-top:2px}
.p-mock-ticket{font-family:var(--hw-font-display);font-size:8px;padding:4px 10px;border-radius:999px;background:var(--hw-tinta);color:var(--hw-plateia);box-shadow:0 0 0 2px var(--hw-plateia),0 0 0 3.5px var(--hw-ouro);white-space:nowrap}
.p-mock-secondary{font-family:var(--hw-font-rotulo);font-size:6.5px;letter-spacing:.1em;text-transform:uppercase;border:1px solid var(--hw-tinta);padding:3.5px 7px;border-radius:2px;color:var(--hw-tinta)}
.p-mock-lamp{position:absolute;inset:0;pointer-events:none;background:radial-gradient(circle at 10% 8%,var(--hw-lampada) 0 2px,transparent 3px),radial-gradient(circle at 90% 8%,var(--hw-lampada) 0 2px,transparent 3px);opacity:.95}
.p-specimen{width:100%;max-width:78mm;padding:8px 9px;border-radius:7px;background:var(--hw-plateia);box-shadow:inset 0 0 0 1px var(--hw-filete);display:grid;gap:2px}
.p-sp-display{font-family:var(--hw-font-display);font-size:22px;line-height:1;color:var(--hw-titulo)}
.p-sp-rotulo{font-family:var(--hw-font-rotulo);font-weight:700;font-size:7.5px;letter-spacing:.22em;text-transform:uppercase;color:var(--hw-ouro-texto)}
.p-sp-cartao{font-family:var(--hw-font-cartao);font-size:13px;color:var(--hw-tinta)}
.p-sp-rabisco{font-family:var(--hw-font-rabisco);font-size:11px;color:var(--hw-tinta-suave)}
.p-sp-texto{font-family:var(--hw-font-texto);font-size:8px;color:var(--hw-tinta)}
.p-frame-stage{width:100%;min-height:104px;padding:16px;border-radius:8px;display:flex;align-items:center;justify-content:center;background:var(--hw-plateia);color:var(--hw-tinta)}
.p-frame{position:relative;padding:16px 22px;min-width:47mm;text-align:center;background:var(--hw-luz);color:var(--hw-tinta)}
.p-frame-label{display:block;font-family:var(--hw-font-rotulo);font-size:6.5px;font-weight:700;letter-spacing:.24em;text-transform:uppercase;color:var(--hw-ouro-texto)}
.p-frame-title{display:block;font-family:var(--hw-font-display);font-size:19px;font-weight:400;line-height:1.1;color:var(--hw-titulo)}
.p-btn-stage{display:flex;align-items:center;justify-content:center;gap:12px;flex-wrap:wrap}
.p-ticket{width:auto;height:40px;padding:0 18px;font-size:13px}
.p-secondary{height:36px;padding:0 15px;display:inline-flex;align-items:center;background:transparent;color:var(--hw-tinta);font-family:var(--hw-font-rotulo);font-weight:700;font-size:8px;letter-spacing:.12em;text-transform:uppercase;cursor:default}
.p-cards-stage{display:flex;align-items:center;justify-content:center;gap:13px;padding:12px 8px;border-radius:8px;background:var(--hw-plateia);color:var(--hw-tinta)}
.p-skin{padding:0}
.p-skin-letras .p-stage{min-height:140px;padding:18px}
.p-skin-letras .p-specimen{max-width:98mm;padding:12px}
.p-skin-letras .p-sp-display{font-size:30px}
.p-skin-letras .p-sp-rotulo{font-size:9px}
.p-skin-letras .p-sp-cartao{font-size:17px}
.p-skin-letras .p-sp-rabisco{font-size:15px}
.p-skin-letras .p-sp-texto{font-size:10px}
.p-skin-molduras .p-frame-stage{min-height:150px;padding:26px}
.p-skin-molduras .p-frame{min-width:64mm;padding:24px 32px}
.p-skin-molduras .p-frame-title{font-size:26px}
.p-skin-botoes .p-stage{min-height:130px;padding:20px}
.p-skin-botoes .p-ticket{height:50px;padding:0 26px;font-size:17px}
.p-skin-botoes .p-secondary{height:44px;padding:0 20px;font-size:10px}
`;

const styleTags = [
  read("styles/tokens.css").replaceAll("url(/assets/", "url(./assets/"),
  read("styles/components.css"),
  read("styles/motion.css"),
  read("ui/Card/Card.module.css"),
  read("ui/Ticket/Ticket.module.css"),
  ...skinFiles,
]
  .map((css) => `<style>${css}</style>`)
  .join("\n");

function cover(kicker, title, lead) {
  return `<header class="p-cover">
  <p class="p-kicker">${esc(kicker)}</p>
  <h1 class="p-title">${esc(title)}</h1>
  ${lead ? `<p class="p-sub">${lead}</p>` : ""}
</header>`;
}

function htmlDoc({ title, kicker, lead, body, footer, cover: withCover = true }) {
  return `<!doctype html>
<html lang="pt-BR" data-motion="reduced" data-idle="off">
<head>
<meta charset="utf-8">
<title>${esc(title)}</title>
${styleTags}
<style>${CSS_BASE}</style>
<style>${CSS_PREVIEW}</style>
</head>
<body>
${withCover ? cover(kicker, title, lead) : ""}
${body}
<footer class="p-run"><span>Hoodwink · Camarim · prancha de arte</span><span>${esc(footer)}</span></footer>
</body>
</html>`;
}

/* ───────────────────────────── prévias ───────────────────────────── */

function coresStage(item, theme, label) {
  return `<figure class="p-fig">
  <div class="p-stage" data-theme="${theme}" data-sk-cores="${item.id}">
    <div class="p-mock">
      <i class="p-mock-lamp" aria-hidden="true"></i>
      <div class="p-mock-top"><span>Sala nº K7Q2</span><span>4/10</span></div>
      <div class="p-mock-banner">Elenco</div>
      <h3 class="p-mock-title">Hoodwink</h3>
      <p class="p-mock-text">Uma pista falsa, um papel escondido.</p>
      <p class="p-mock-alert">Carta protegida</p>
      <div class="p-mock-actions">
        <span class="p-mock-ticket">Criar sala</span>
        <span class="p-mock-secondary">Voltar</span>
      </div>
    </div>
  </div>
  <figcaption>${esc(label)}</figcaption>
</figure>`;
}

function fontSpecimen(item, theme, label) {
  return `<figure class="p-fig">
  <div class="p-stage" data-theme="${theme}" data-sk-letras="${item.id}">
    <div class="p-specimen">
      <p class="p-sp-display">Elenco</p>
      <p class="p-sp-rotulo">Criar sala</p>
      <p class="p-sp-cartao">Trilionário</p>
      <p class="p-sp-rabisco">☞ anote aqui</p>
      <p class="p-sp-texto">O texto corrido continua em Literata, sempre.</p>
    </div>
  </div>
  <figcaption>${esc(label)}</figcaption>
</figure>`;
}

function molduraFrame(lugar, theme, label, sk) {
  const f = `<i aria-hidden="true" class="hw-moldura-f m-f"></i>`;
  return `<figure class="p-fig">
  <div class="p-frame-stage" data-theme="${theme}" ${sk ? `data-sk-molduras="${sk}"` : ""}>
    <div class="hw-moldura m-contorno p-frame" data-lugar="${lugar}">
      ${f}${f}${f}
      <span class="p-frame-label">Sala nº K7Q2</span>
      <b class="p-frame-title">Elenco</b>
    </div>
  </div>
  <figcaption>${esc(label)}</figcaption>
</figure>`;
}

function botoesStage(item, theme, label) {
  return `<figure class="p-fig">
  <div class="p-stage" data-theme="${theme}" data-sk-botoes="${item.id}">
    <div class="p-btn-stage">
      <span class="ticket p-ticket">Criar sala</span>
      <span class="hw-botao2 p-secondary">Cancelar</span>
    </div>
  </div>
  <figcaption>${esc(label)}</figcaption>
</figure>`;
}

function cardsStage(item, theme) {
  return `<figure class="p-fig">
  <div class="p-cards-stage" data-theme="${theme}" data-sk-cartas="${item.id}">
    <div class="hw-carta face s">
      <div class="bottom withCost"><span class="hw-num cost">3</span>Protegida</div>
      <div class="top"><span class="name">Assassino</span></div>
    </div>
    <div class="card back s"></div>
    <div class="card hidden s"><span class="hiddenLabel">escondida</span><span class="ask">?</span></div>
    <div class="card dead s"><span class="deadName">Coveiro</span></div>
  </div>
  <figcaption>${theme === "light" ? "Matinê (claro) · face, verso, escondida e morta" : "Estreia (escuro) · face, verso, escondida e morta"}</figcaption>
</figure>`;
}

/* ───────────────────────────── fichas ───────────────────────────── */

function fichaHead(title, note) {
  return `<h2 class="p-h2">${esc(title)}</h2><p class="p-note">${note}</p>`;
}

function fichaCores(first = false) {
  const rows = COLOR_META.map(
    ([token, use]) => `<tr>
      <td><b class="p-code">--hw-${token}</b><div class="p-muted" style="font-size:8px">${use}</div></td>
      <td><span class="p-blank"></span></td>
      <td><span class="p-blank"></span></td>
    </tr>`,
  ).join("");
  return `<section class="p-form${first ? "" : " p-break"}">
  ${fichaHead("Ficha · nova peça de CORES", "Preencha os 15 tokens no claro e no escuro. Aceito HEX (#RRGGBB) ou RGBA. O contraste é conferido depois: tinta sobre plateia ≥ 4,5:1 e título sobre plateia ≥ 3:1.")}
  <div class="p-fields">
    <div class="p-field"><span class="p-label">Nome da peça</span><span class="p-line"></span></div>
    <div class="p-field"><span class="p-label">Id (slug, sem espaços)</span><span class="p-line"></span></div>
    <div class="p-field wide"><span class="p-label">Intenção em uma frase</span><span class="p-line"></span></div>
  </div>
  <table class="p-table">
    <thead><tr><th style="width:52%">Token e uso</th><th>Matinê (claro)</th><th>Estreia (escuro)</th></tr></thead>
    <tbody>${rows}</tbody>
  </table>
  <div class="p-fields">
    <div class="p-field wide"><span class="p-label">Observações, texturas ou cores extras</span><div class="p-canvas" style="height:20mm"></div></div>
  </div>
</section>`;
}

function fichaLetras(first = false) {
  const rows = [
    ["display", "Títulos e nomes nas cartas"],
    ["rotulo", "Rótulos em caixa alta"],
    ["cartao", "Rótulos dentro das cartas"],
    ["rabisco", "Rabiscos e anotações"],
  ]
    .map(
      ([role, use]) => `<tr>
      <td><b class="p-code">--hw-font-${role}</b><div class="p-muted" style="font-size:8px">${use}</div></td>
      <td><span class="p-blank"></span></td>
      <td><span class="p-blank"></span></td>
      <td><span class="p-blank"></span></td>
    </tr>`,
    )
    .join("");
  return `<section class="p-form${first ? "" : " p-break"}">
  ${fichaHead("Ficha · nova peça de LETRAS", "Informe o nome exato da fonte, os pesos usados e a origem (Google Fonts, licença ou arquivo). O texto corrido é travado em Literata e não entra na ficha.")}
  <div class="p-fields">
    <div class="p-field"><span class="p-label">Nome da peça</span><span class="p-line"></span></div>
    <div class="p-field"><span class="p-label">Id (slug, sem espaços)</span><span class="p-line"></span></div>
    <div class="p-field wide"><span class="p-label">Intenção em uma frase</span><span class="p-line"></span></div>
  </div>
  <table class="p-table">
    <thead><tr><th style="width:34%">Papel</th><th>Fonte</th><th>Pesos</th><th>Origem / licença</th></tr></thead>
    <tbody>${rows}</tbody>
  </table>
  <p class="p-note"><b>Não preencher:</b> o texto corrido (<span class="p-code">--hw-font-texto</span>) fica sempre em Literata. Se a fonte tiver fallback obrigatório, anote abaixo.</p>
  <div class="p-fields">
    <div class="p-field wide"><span class="p-label">Fallbacks, caracteres especiais e observações</span><div class="p-canvas" style="height:18mm"></div></div>
  </div>
</section>`;
}

function fichaMolduras(first = false) {
  return `<section class="p-form${first ? "" : " p-break"}">
  ${fichaHead("Ficha · nova peça de MOLDURAS", "A moldura é desenhada em CSS e vale no lobby, no elenco, no fim e nos avisos. Descreva o contorno e desenhe no quadro. Use os tokens de cor (ouro, veludo, filete, tinta…), não HEX solto.")}
  <div class="p-fields">
    <div class="p-field"><span class="p-label">Nome da peça</span><span class="p-line"></span></div>
    <div class="p-field"><span class="p-label">Id (slug, sem espaços)</span><span class="p-line"></span></div>
    <div class="p-field wide"><span class="p-label">Lugares</span>
      <span><span class="p-check"><i></i>Lobby (letreiro)</span><span class="p-check"><i></i>Elenco</span><span class="p-check"><i></i>Fim</span><span class="p-check"><i></i>Aviso</span></span>
    </div>
  </div>
  <div class="p-fields">
    <div class="p-field"><span class="p-label">Contorno (topo / base)</span><span class="p-line"></span></div>
    <div class="p-field"><span class="p-label">Contorno (laterais)</span><span class="p-line"></span></div>
    <div class="p-field"><span class="p-label">Espessura e raio dos cantos</span><span class="p-line"></span></div>
    <div class="p-field"><span class="p-label">Ornamento (cantoneira, sanefa, picote…)</span><span class="p-line"></span></div>
    <div class="p-field"><span class="p-label">Cores / tokens</span><span class="p-line"></span></div>
    <div class="p-field"><span class="p-label">Animação em 3 quadros (o que muda)</span><span class="p-line"></span></div>
  </div>
  <span class="p-label">Desenho da moldura</span>
  <div class="p-canvas" style="height:40mm"></div>
  <div class="p-fields">
    <div class="p-field wide"><span class="p-label">Observações</span><div class="p-canvas" style="height:16mm"></div></div>
  </div>
</section>`;
}

function fichaBotoes(first = false) {
  return `<section class="p-form${first ? "" : " p-break"}">
  ${fichaHead("Ficha · nova peça de BOTÕES", "São dois botões: o principal (pílula de ação) e o secundário (contorno). Descreva forma, cor, borda, sombra e tipografia de cada um.")}
  <div class="p-fields">
    <div class="p-field"><span class="p-label">Nome da peça</span><span class="p-line"></span></div>
    <div class="p-field"><span class="p-label">Id (slug, sem espaços)</span><span class="p-line"></span></div>
    <div class="p-field wide"><span class="p-label">Intenção em uma frase</span><span class="p-line"></span></div>
  </div>
  <div class="p-cols">
    <div class="p-card">
      <span class="p-label">Botão principal</span>
      <div class="p-fields" style="grid-template-columns:1fr">
        <div class="p-field"><span class="p-label">Fundo</span><span class="p-line"></span></div>
        <div class="p-field"><span class="p-label">Texto</span><span class="p-line"></span></div>
        <div class="p-field"><span class="p-label">Borda / anel</span><span class="p-line"></span></div>
        <div class="p-field"><span class="p-label">Raio e sombra</span><span class="p-line"></span></div>
        <div class="p-field"><span class="p-label">Tipografia</span><span class="p-line"></span></div>
      </div>
    </div>
    <div class="p-card">
      <span class="p-label">Botão secundário</span>
      <div class="p-fields" style="grid-template-columns:1fr">
        <div class="p-field"><span class="p-label">Fundo</span><span class="p-line"></span></div>
        <div class="p-field"><span class="p-label">Texto</span><span class="p-line"></span></div>
        <div class="p-field"><span class="p-label">Borda / anel</span><span class="p-line"></span></div>
        <div class="p-field"><span class="p-label">Raio e sombra</span><span class="p-line"></span></div>
        <div class="p-field"><span class="p-label">Tipografia</span><span class="p-line"></span></div>
      </div>
    </div>
  </div>
  <span class="p-label" style="margin-top:7px">Desenho dos dois botões</span>
  <div class="p-canvas" style="height:30mm"></div>
  <div class="p-fields">
    <div class="p-field wide"><span class="p-label">Observações (estados desabilitado, foco…)</span><div class="p-canvas" style="height:14mm"></div></div>
  </div>
</section>`;
}

function fichaCartas(first = false) {
  const rows = [
    ["--hw-verso", "Padrão do verso (CSS ou SVG)"],
    ["--hw-carta", "Cor da faixa da face"],
    ["--hw-carta-tinta", "Texto da faixa"],
    ["--hw-carta-rotulo", "Rótulos e notas da face"],
    ["--hw-escondida-borda", "Borda da carta escondida"],
  ]
    .map(
      ([token, use]) => `<tr>
      <td><b class="p-code">${token}</b><div class="p-muted" style="font-size:8px">${use}</div></td>
      <td><span class="p-blank"></span></td>
    </tr>`,
    )
    .join("");
  return `<section class="p-form${first ? "" : " p-break"}">
  ${fichaHead("Ficha · nova peça de CARTAS", "A face mostra nome, custo e estado; o verso é o padrão do baralho. Descreva ou desenhe o verso e informe as 5 cores. Para verso ilustrado, pode anexar um SVG (proporção 3:4) que eu transformo em asset.")}
  <div class="p-fields">
    <div class="p-field"><span class="p-label">Nome da peça</span><span class="p-line"></span></div>
    <div class="p-field"><span class="p-label">Id (slug, sem espaços)</span><span class="p-line"></span></div>
    <div class="p-field wide"><span class="p-label">Intenção em uma frase</span><span class="p-line"></span></div>
  </div>
  <div class="p-cols">
    <div>
      <span class="p-label">Verso (desenho / padrão)</span>
      <div class="p-canvas" style="height:42mm"></div>
    </div>
    <div>
      <span class="p-label">Cores</span>
      <table class="p-table"><tbody>${rows}</tbody></table>
      <p class="p-note">Medidas de referência: face <b>78×110</b> (mão), <b>132×186</b> (catálogo), <b>104×144</b> (mesa desktop).</p>
    </div>
  </div>
  <div class="p-fields">
    <div class="p-field wide"><span class="p-label">Observações (brilho, borda, textura)</span><div class="p-canvas" style="height:14mm"></div></div>
  </div>
</section>`;
}

/* ─────────────────────────── blocos de peça ─────────────────────────── */

const DESC = {
  holofote: "O visual padrão: papel quente, vinho de veludo e ouro de bilheteria. Ponto de partida de tudo.",
  bastidor: "Papelão e madeira de backstage: quente, terroso, com vinho queimado. Mais informal que o Holofote.",
  noir: "Preto, branco e cinza com um único vermelho de alerta. Contraste alto, clima de cinema mudo.",
  ensaio: "Azul de ensaio e prata de refletor, com ouro discreto. Mais frio e técnico que os outros pacotes.",
  cartaz: "Limelight nos títulos, Josefin nos rótulos, Patrick Hand nas cartas e Permanent Marker nos rabiscos.",
  pincel: "Caveat Brush nos títulos e rabiscos, Patrick Hand nos rótulos e cartas. Traço de guache.",
  gala: "Playfair Display SC nos títulos e Cormorant SC nos rótulos. Versalete de programa de gala.",
  canhoto: "Ingresso de cinema com cantos mordidos e picote dourado, como um canhoto destacável.",
  cantoneiras: "Cantoneiras douradas discretas nos quatro cantos, como um porta-retrato de gala.",
  papelao: "Recorte de papelão grosso, torto e preso com fita. O visual do pacote Bastidor.",
  bambinela: "Sanefa de cortina de veludo no topo e filete nas laterais. Clima de palco.",
  ingresso: "O botão padrão: pílula de tinta com anel de ouro, como um ingresso de teatro.",
  placa: "Placa de madeira com sombra dura. O visual do pacote Bastidor.",
  carimbo: "Carimbo de tinta vermelha, torto e com borda dupla. O visual do pacote Noir.",
  veludo: "Almofada de veludo com filete de ouro. O visual do pacote Ensaio.",
  xadrez: "Xadrez preto e branco de tabuleiro, faixas pretas e texto claro.",
  azulejo: "Azulejo azul e branco, faixas azuis e rótulos claros.",
};

function pieceHead(item) {
  const def = !item.custom;
  const desc = DESC[item.id];
  return `<div class="p-piece-head">
    <h3>${esc(item.name)}</h3>
    ${def ? `<span class="p-badge">padrão</span>` : ""}
    <span class="p-id">${item.slot} / ${item.id}</span>
  </div>
  ${desc ? `<p class="p-note" style="margin:-3px 0 7px">${esc(desc)}</p>` : ""}`;
}

function coresPiece(item, withHead = true) {
  const { light, dark } = coresOf(item);
  const rows = COLOR_META.map(([token, use]) => {
    const l = light[`--hw-${token}`] ?? "—";
    const d = dark[`--hw-${token}`] ?? "—";
    return `<tr>
      <td><b class="p-code">--hw-${token}</b><div class="p-muted" style="font-size:7.5px">${use}</div></td>
      <td><span class="p-swatch" style="background:${esc(l)}"></span> <span class="p-code">${esc(l)}</span></td>
      <td><span class="p-swatch" style="background:${esc(d)}"></span> <span class="p-code">${esc(d)}</span></td>
    </tr>`;
  }).join("");
  const cLight = contrast(light["--hw-tinta"], light["--hw-plateia"]);
  const tLight = contrast(light["--hw-titulo"], light["--hw-plateia"]);
  const cDark = contrast(dark["--hw-tinta"], dark["--hw-plateia"]);
  const tDark = contrast(dark["--hw-titulo"], dark["--hw-plateia"]);
  const check = (n, min, label) =>
    `<span class="${n != null && n >= min ? "p-ok" : "p-warn"}">${label} ${ratio(n)} ${badge(n, min)}</span>`;
  return `<section class="p-piece">
  ${withHead ? pieceHead(item) : ""}
  <div class="p-grid2">
    ${coresStage(item, "light", "Matinê (claro)")}
    ${coresStage(item, "dark", "Estreia (escuro)")}
  </div>
  <h4 class="p-h3">Tokens da peça</h4>
  <table class="p-table">
    <thead><tr><th style="width:48%">Token</th><th>Matinê (claro)</th><th>Estreia (escuro)</th></tr></thead>
    <tbody>${rows}</tbody>
  </table>
  <p class="p-note">${check(cLight, 4.5, "tinta/plateia")} · ${check(tLight, 3, "título/plateia")} · ${check(
    cDark,
    4.5,
    "escuro tinta/plateia",
  )} · ${check(tDark, 3, "escuro título/plateia")}</p>
</section>`;
}

function letrasPiece(item, withHead = true) {
  const { base, over } = fontsOf(item);
  const roles = [
    ["display", "Títulos, nomes nas cartas"],
    ["rotulo", "Rótulos em caixa alta"],
    ["cartao", "Rótulos dentro das cartas"],
    ["rabisco", "Rabiscos e anotações"],
    ["texto", "Texto corrido (travado em Literata)"],
  ];
  const rows = roles
    .map(([role, use]) => {
      const key = `--hw-font-${role}`;
      const val = over[key] ?? base[key] ?? "—";
      const herd = !over[key];
      return `<tr>
        <td><b class="p-code">${key}</b><div class="p-muted" style="font-size:7.5px">${use}</div></td>
        <td class="p-code">${esc(val)}</td>
        <td>${herd ? `<span class="p-muted" style="font-size:8px">herdado da base</span>` : `<span class="p-ok">definido</span>`}</td>
      </tr>`;
    })
    .join("");
  return `<section class="p-piece">
  ${withHead ? pieceHead(item) : ""}
  <div class="p-grid2">
    ${fontSpecimen(item, "light", "Matinê (claro)")}
    ${fontSpecimen(item, "dark", "Estreia (escuro)")}
  </div>
  <h4 class="p-h3">Fontes da peça</h4>
  <table class="p-table">
    <thead><tr><th style="width:38%">Papel</th><th>Família</th><th style="width:18%">Nesta peça</th></tr></thead>
    <tbody>${rows}</tbody>
  </table>
  <p class="p-note">Só display, rótulo, carta e rabisco mudam. O texto corrido fica em Literata.</p>
</section>`;
}

const MOLDURA_RECIPES = {
  holofote: [
    ["Lobby", "Letreiro: borda de veludo com lâmpadas apagadas e três camadas que acendem um terço das lâmpadas por vez."],
    ["Elenco / fim / aviso", "Filete fervendo: contorno de ouro 1,5px + filete interno a 7px, cantos levemente irregulares; três quadros tremidos."],
    ["Cores", "Usa --hw-ouro, --hw-veludo e --hw-lampada (acompanha a peça de cores)."],
  ],
  canhoto: [
    ["Contorno", "Picote tracejado de ouro 1,5px, recuado 5px da borda."],
    ["Cantos", "Mordida circular de 12px nos quatro cantos (máscara)."],
    ["Quadros", "Três contornos empilhados, deslocados em meio pixel."],
  ],
  cantoneiras: [
    ["Contorno", "Sem contorno contínuo; só cantoneiras."],
    ["Cantos", "Cantoneira de ouro em L, 22×2px, nos quatro cantos."],
    ["Quadros", "Três camadas com recuos diferentes, formando o tremor."],
  ],
  papelao: [
    ["Contorno", "Borda grossa de 4px em --hw-papelao-escuro (#8E6A45)."],
    ["Cantos", "Cantos irregulares (raios 3–9px) e levemente torcidos."],
    ["Quadros", "Três contornos com raios e rotações diferentes."],
  ],
  bambinela: [
    ["Contorno", "Filete de veludo 2px nas laterais e na base."],
    ["Topo", "Sanefa de cortina: semicírculos de veludo de 9px a cada 20px."],
    ["Quadros", "Três camadas com a sanefa deslocada horizontalmente."],
  ],
};

function molduraPiece(item, withHead = true) {
  const recipes = MOLDURA_RECIPES[item.id] ?? [];
  const list = recipes.map(([k, v]) => `<li><b>${esc(k)}</b><span>${esc(v)}</span></li>`).join("");
  const isDefault = !item.custom;
  return `<section class="p-piece">
  ${withHead ? pieceHead(item) : ""}
  <div class="p-grid2">
    ${molduraFrame("elenco", "light", "Matinê (claro)", item.custom ? item.id : "")}
    ${molduraFrame("elenco", "dark", "Estreia (escuro)", item.custom ? item.id : "")}
  </div>
  ${isDefault ? `<div style="margin-top:9px">${molduraFrame("lobby", "light", "Lobby · letreiro (claro)", "")}</div>` : ""}
  <h4 class="p-h3">Receita do desenho</h4>
  <ul class="p-recipe">${list}</ul>
</section>`;
}

const BOTAO_RECIPES = {
  ingresso: [
    ["Principal", "Pílula de tinta (--hw-tinta) com texto em --hw-plateia, anel duplo: 3px de plateia + 2px de ouro. Raio 999px."],
    ["Secundário", "Contorno de 1,5px em --hw-tinta, sem raio."],
  ],
  placa: [
    ["Principal", "Placa de madeira: fundo --hw-ouro, texto --hw-tinta, raio 4px, sombra dura de 4px."],
    ["Secundário", "Contorno 2px em --hw-tinta, raio 4px, sombra dura de 3px."],
  ],
  carimbo: [
    ["Principal", "Carimbo: borda dupla de 3px em --hw-alerta, fundo transparente, texto em --hw-alerta, raio 6px, inclinado -1°."],
    ["Secundário", "Tracejado de 1,5px em --hw-tinta, raio 6px."],
  ],
  veludo: [
    ["Principal", "Almofada de veludo: fundo --hw-veludo, texto --hw-sobre-veludo, anel interno de ouro 2px, raio 14px, sombra suave."],
    ["Secundário", "Contorno 1,5px em --hw-veludo, texto veludo, raio 14px."],
  ],
};

function botoesPiece(item, withHead = true) {
  const recipes = BOTAO_RECIPES[item.id] ?? [];
  const list = recipes.map(([k, v]) => `<li><b>${esc(k)}</b><span>${esc(v)}</span></li>`).join("");
  return `<section class="p-piece">
  ${withHead ? pieceHead(item) : ""}
  <div class="p-grid2">
    ${botoesStage(item, "light", "Matinê (claro)")}
    ${botoesStage(item, "dark", "Estreia (escuro)")}
  </div>
  <h4 class="p-h3">Receita do desenho</h4>
  <ul class="p-recipe">${list}</ul>
</section>`;
}

function cartasPiece(item, withHead = true) {
  const { light } = cartasOf(item);
  const rows = [
    ["--hw-carta", "Cor da faixa da face", light["--hw-carta"]],
    ["--hw-carta-tinta", "Texto da faixa", light["--hw-carta-tinta"]],
    ["--hw-carta-rotulo", "Rótulos e notas", light["--hw-carta-rotulo"]],
    ["--hw-escondida-borda", "Borda da carta escondida", light["--hw-escondida-borda"]],
    ["--hw-verso", "Padrão do verso", light["--hw-verso"]],
  ]
    .map(([token, use, val]) => {
      const isHex = /^#[0-9a-f]{3,6}$/i.test(val ?? "");
      return `<tr>
        <td><b class="p-code">${token}</b><div class="p-muted" style="font-size:7.5px">${use}</div></td>
        <td>${isHex ? `<span class="p-swatch" style="background:${esc(val)}"></span> ` : ""}<span class="p-code">${esc(val ?? "—")}</span></td>
      </tr>`;
    })
    .join("");
  return `<section class="p-piece">
  ${withHead ? pieceHead(item) : ""}
  ${cardsStage(item, "light")}
  ${cardsStage(item, "dark")}
  <h4 class="p-h3">Tokens da peça</h4>
  <table class="p-table"><tbody>${rows}</tbody></table>
  <p class="p-note">A peça de cartas usa as mesmas cores no claro e no escuro; só o entorno muda.</p>
</section>`;
}

const PIECE = { cores: coresPiece, letras: letrasPiece, molduras: molduraPiece, botoes: botoesPiece, cartas: cartasPiece };
const FICHA = { cores: fichaCores, letras: fichaLetras, molduras: fichaMolduras, botoes: fichaBotoes, cartas: fichaCartas };

const CATEGORY = {
  cores: {
    n: "01",
    title: "Cores",
    lead:
      "Muda as cores de todas as telas, no claro (Matinê) e no escuro (Estreia). A peça entrega 15 tokens; o resto do jogo se pinta sozinho.",
    where: "Todas as telas",
    deliver: "HEX/RGBA dos 15 tokens, no claro e no escuro",
  },
  letras: {
    n: "02",
    title: "Letras",
    lead: "Muda a tipografia de títulos, rótulos, cartas e rabiscos. O texto corrido fica sempre em Literata.",
    where: "Títulos, rótulos e cartas",
    deliver: "Nome da fonte, pesos e origem/licença",
  },
  molduras: {
    n: "03",
    title: "Molduras",
    lead: "Muda o contorno das caixas do lobby, do elenco, do fim de partida e dos avisos. O conteúdo não se mexe.",
    where: "Lobby, elenco, fim, aviso",
    deliver: "Desenho, medidas, cores e os 3 quadros da animação",
  },
  botoes: {
    n: "04",
    title: "Botões",
    lead: "Muda o botão principal (pílula de ação) e o botão secundário (contorno) de todas as telas.",
    where: "Ações e cancelamentos",
    deliver: "Forma, cores, borda, sombra de cada botão",
  },
  cartas: {
    n: "05",
    title: "Cartas",
    lead: "Muda o verso do baralho, as faixas da face e os estados escondida e morta.",
    where: "Mão, mesa e contestação",
    deliver: "Padrão do verso + 5 cores da carta",
  },
};

function categoryDoc(slot) {
  const cat = CATEGORY[slot];
  const items = itemsOf(slot);
  const defaultId = DEFAULT_EQUIP[slot];

  const overview = items
    .map((i) => `<tr><td><b>${esc(i.name)}</b></td><td class="p-code">${i.id}</td><td>${i.id === defaultId ? "padrão" : "peça"}</td></tr>`)
    .join("");

  const body = `
<p class="p-lede">${esc(cat.lead)}</p>
<div class="p-cols" style="margin:10px 0 12px">
  <div class="p-card">
    <span class="p-label">O que a peça entrega</span>
    <p style="font-size:9.5px">${esc(cat.deliver)}</p>
    <span class="p-label" style="margin-top:6px">Onde aparece</span>
    <p style="font-size:9.5px">${esc(cat.where)}</p>
  </div>
  <div class="p-card">
    <span class="p-label">Peças desta categoria</span>
    <table class="p-table"><thead><tr><th>Nome</th><th>Id</th><th>Tipo</th></tr></thead><tbody>${overview}</tbody></table>
  </div>
</div>
<h2 class="p-h2" style="margin-top:14px">Como ler esta prancha</h2>
<div class="p-cols3">
  <div class="p-tip"><b>1 · Prévia real</b>O desenho que você vê é o próprio jogo, no claro e no escuro. Nada de mockup aproximado.</div>
  <div class="p-tip"><b>2 · Receita</b>Medidas, cores e o que muda em cada um dos três quadros da animação.</div>
  <div class="p-tip"><b>3 · Ficha</b>No fim da prancha, a folha em branco para você preencher e devolver.</div>
</div>
<h2 class="p-h2" style="margin-top:14px">Antes de desenhar</h2>
<ul class="p-list">
  <li><b>✓</b><span>Use os <b>tokens de cor</b> do jogo. Só a peça de cores define HEX; as outras categorias herdam.</span></li>
  <li><b>✓</b><span>Confira o <b>contraste</b>: o texto precisa continuar legível no claro e no escuro.</span></li>
  <li><b>✓</b><span>Não mexa em <b>layout, espaçamento ou texto</b>: a peça muda só o acabamento.</span></li>
  <li><b>✓</b><span>Molduras, botões e cartas têm <b>três quadros</b> de animação. Descreva o que muda entre eles.</span></li>
  <li><b>✓</b><span>Se anexar arte, mande <b>SVG</b> (ou PNG grande) junto da ficha. Eu cuido da produção da peça.</span></li>
</ul>
<h2 class="p-h2" style="margin-top:14px">Combinações que já existem</h2>
<table class="p-table">
  <thead><tr><th style="width:38%">Pacote</th><th>Peça de ${esc(cat.title.toLowerCase())}</th></tr></thead>
  <tbody>${PACKS.map((p) => `<tr><td><b>${esc(p.name)}</b></td><td>${esc(ITEMS.find((i) => i.slot === slot && i.id === p.equip[slot])?.name ?? p.equip[slot])}</td></tr>`).join("")}</tbody>
</table>
<div class="p-break"></div>
${items.map((i) => PIECE[slot](i)).join("\n")}
${FICHA[slot]()}
`;

  return htmlDoc({
    title: `${cat.title}`,
    kicker: `Hoodwink · Camarim · prancha ${cat.n} · ${cat.title.toLowerCase()}`,
    lead: `Base para o artista criar novas peças de <b>${esc(cat.title.toLowerCase())}</b> e devolver preenchido. Ao final, a ficha em branco.`,
    body,
    footer: `Prancha ${cat.n} · ${cat.title}`,
  });
}

/* ────────────────────────────── guia ────────────────────────────── */

function guideBody() {
  const packs = PACKS.map(
    (p) => `<tr>
      <td><b>${esc(p.name)}</b></td>
      ${SLOTS.map((s) => `<td>${esc(ITEMS.find((i) => i.slot === s && i.id === p.equip[s])?.name ?? p.equip[s])}</td>`).join("")}
    </tr>`,
  ).join("");

  const map = SLOTS.map(
    (s) => `<tr>
      <td><b>${esc(CATEGORY[s].title)}</b></td>
      <td class="p-code">${itemsOf(s).map((i) => i.id).join(", ")}</td>
      <td>${esc(CATEGORY[s].deliver)}</td>
    </tr>`,
  ).join("");

  const body = `
<p class="p-lede">Este caderno reúne o guia e as <b>20 skins atuais</b>, uma por página. Cada categoria termina com a ficha em branco para a nova proposta. A ideia é simples: você desenha e preenche a ficha; o resto é comigo.</p>
<p class="p-note">Ordem: Cores → Letras → Molduras → Botões → Cartas.</p>

<h2 class="p-h2" style="margin-top:15px">As cinco categorias</h2>
<table class="p-table">
  <thead><tr><th style="width:15%">Categoria</th><th>O que muda</th><th style="width:24%">Onde aparece</th><th style="width:26%">O que você entrega</th></tr></thead>
  <tbody>
    ${SLOTS.map(
      (s) => `<tr>
        <td><b>${esc(CATEGORY[s].title)}</b></td>
        <td>${esc(CATEGORY[s].lead)}</td>
        <td>${esc(CATEGORY[s].where)}</td>
        <td>${esc(CATEGORY[s].deliver)}</td>
      </tr>`,
    ).join("")}
  </tbody>
</table>

<h2 class="p-h2" style="margin-top:15px">Como preencher</h2>
<ol class="p-list">
  <li><b>1</b><span>Dê um <b>nome</b> e um <b>id</b> curto, em minúsculas e sem espaços (ex.: <span class="p-code">veludo</span>, <span class="p-code">papelao</span>). O id é o nome curto que identifica a peça.</span></li>
  <li><b>2</b><span>Preencha os valores pedidos. Em cores, use HEX (<span class="p-code">#1E1512</span>) ou RGBA quando houver transparência.</span></li>
  <li><b>3</b><span>Em molduras, botões e cartas, desenhe no quadro e escreva a receita (medidas, bordas, sombras, o que muda nos 3 quadros).</span></li>
  <li><b>4</b><span>Se a arte for complexa, anexe um SVG/PNG de referência junto com a ficha. Eu cuido da produção da peça.</span></li>
  <li><b>5</b><span>Devolva as fichas preenchidas (PDF escaneado, foto legível ou os valores digitados).</span></li>
</ol>

<h2 class="p-h2" style="margin-top:15px">As travas do jogo</h2>
<div class="p-cols">
  <div class="p-card">
    <span class="p-label">Contraste</span>
    <p style="font-size:9.5px">Texto principal sobre o fundo ≥ <b>4,5:1</b>; título sobre o fundo ≥ <b>3:1</b>, no claro e no escuro. Eu confiro depois e aviso se algo precisar ajustar.</p>
    <span class="p-label" style="margin-top:6px">Texto corrido</span>
    <p style="font-size:9.5px">Fica sempre em Literata. Peças de letras mudam só display, rótulo, carta e rabisco.</p>
  </div>
  <div class="p-card">
    <span class="p-label">Cores por token</span>
    <p style="font-size:9.5px">Nada de cor solta fora dos tokens: a peça de cores define os 15 tokens e o resto do jogo acompanha.</p>
    <span class="p-label" style="margin-top:6px">Sem preço nem cadeado</span>
    <p style="font-size:9.5px">Toda peça é liberada. Não há estado bloqueado para desenhar.</p>
  </div>
</div>

<h2 class="p-h2" style="margin-top:15px">Pacotes que já existem</h2>
<p class="p-note" style="margin-bottom:4px">Um pacote é só uma combinação de peças. Serve de referência de coerência entre as categorias.</p>
<table class="p-table">
  <thead><tr><th>Pacote</th>${SLOTS.map((s) => `<th>${esc(CATEGORY[s].title)}</th>`).join("")}</tr></thead>
  <tbody>${packs}</tbody>
</table>

<h2 class="p-h2" style="margin-top:15px">Mapa das peças e dos ids</h2>
<table class="p-table">
  <thead><tr><th style="width:15%">Categoria</th><th>Ids existentes</th><th style="width:38%">Entrega</th></tr></thead>
  <tbody>${map}</tbody>
</table>

<h2 class="p-h2" style="margin-top:15px">Medidas de referência</h2>
<table class="p-table">
  <tbody>
    <tr><td><b>Carta na mão</b></td><td>78 × 110 px</td><td><b>Carta no catálogo</b></td><td>132 × 186 px</td></tr>
    <tr><td><b>Carta na mesa (desktop)</b></td><td>104 × 144 px</td><td><b>Botão principal</b></td><td>58–60 px de altura</td></tr>
    <tr><td><b>Raio pequeno</b></td><td>6 px</td><td><b>Raio médio</b></td><td>10 px</td></tr>
    <tr><td><b>Pílula</b></td><td>999 px</td><td><b>Moldura</b></td><td>contorno recuado ~5 px</td></tr>
  </tbody>
</table>
`;
  return body;
}

function guideDoc() {
  return htmlDoc({
    title: "Skins: guia do artista",
    kicker: "Hoodwink · Camarim · guia geral",
    lead: "O que é uma skin, o que muda em cada categoria, as travas do jogo e como preencher a ficha.",
    body: guideBody(),
    footer: "Guia · leia antes das pranchas",
  });
}

/* ────────────────────────────── ficha geral ────────────────────────────── */

function fichaDoc() {
  const body = `
<p class="p-lede">Ficha em branco para criar uma peça nova em qualquer categoria. Preencha só a categoria desejada (ou todas, se for um pacote completo). Uma página por categoria.</p>
${fichaCores(true)}
${fichaLetras()}
${fichaMolduras()}
${fichaBotoes()}
${fichaCartas()}
`;
  return htmlDoc({
    title: "Ficha em branco",
    kicker: "Hoodwink · Camarim · ficha do artista",
    lead: "Imprima, desenhe e preencha. Devolva legível para a peça ser produzida.",
    body,
    footer: "Ficha geral · todas as categorias",
  });
}

/* ─────────────────────────────── por skin ─────────────────────────────── */

const CHECKLIST = {
  cores: [
    "Os 15 tokens, no claro e no escuro. Os 12 sólidos em HEX; <span class=\"p-code\">poca</span>, <span class=\"p-code\">cone</span> e <span class=\"p-code\">scrim</span> podem ser RGBA.",
    "Texto principal sobre o fundo ≥ 4,5:1; título sobre o fundo ≥ 3:1, nos dois temas.",
    "Nada de cor solta fora dos tokens: a peça de cores manda no resto do jogo.",
  ],
  letras: [
    "Só display, rótulo, carta e rabisco. O texto corrido fica em Literata.",
    "Informe o nome exato da fonte, os pesos e a origem/licença.",
    "Google Fonts: um import com os pesos usados. Fonte de arquivo: avise que eu preparo o @font-face fora da peça.",
  ],
  molduras: [
    "Use os tokens de cor; não invente HEX solto.",
    "Mantenha as três camadas (<span class=\"p-code\">hw-moldura-f</span>): são os 3 quadros da animação.",
    "Não mexa em padding, largura, altura ou display do conteúdo.",
    "Medida de referência: contorno recuado ~5px; raio 6–10px.",
  ],
  botoes: [
    "Principal = ticket (pílula); secundário = contorno. Se não desenhar o secundário, ele fica o padrão.",
    "Use os tokens de cor e a tipografia do jogo.",
    "Não mexa em altura (58–60px), largura, padding ou font-size.",
  ],
  cartas: [
    "O verso, a faixa, o texto da faixa, os rótulos e a borda da escondida: 5 tokens.",
    "Mesmas cores no claro e no escuro; só o entorno muda.",
    "Para verso ilustrado, anexe SVG na proporção 3:4.",
    "Garanta o texto legível sobre a faixa (≥ 4,5:1).",
    "Medidas: face 78×110 (mão), 132×186 (catálogo), 104×144 (mesa).",
  ],
};

/** Categorias com página 1 curta ganham uma área de anotações. */
const NOTES = new Set(["letras", "molduras", "botoes"]);

function skinCover(item) {
  const cat = CATEGORY[item.slot];
  const packs = PACKS.filter((p) => p.equip[item.slot] === item.id).map((p) => p.name);
  const sisters = itemsOf(item.slot).filter((i) => i.id !== item.id).map((i) => i.name);
  const sub = [
    `Peça <span class="p-code">${item.slot}/${item.id}</span>`,
    item.custom ? "" : "<b>padrão</b>",
    packs.length ? `Pacotes: ${packs.join(", ")}` : "",
  ]
    .filter(Boolean)
    .join(" · ");
  const lead = `${sub}<br>${esc(DESC[item.id] ?? "")}${sisters.length ? `<br><span class="p-muted">Irmãs: ${esc(sisters.join(", "))}</span>` : ""}`;
  return cover(`Hoodwink · Camarim · ${cat.title.toLowerCase()}`, item.name, lead);
}

function skinBody(item) {
  const cat = CATEGORY[item.slot];
  return `
<p class="p-lede">${esc(cat.lead)}</p>
<div class="p-skin p-skin-${item.slot}">
${PIECE[item.slot](item, false)}
</div>
${NOTES.has(item.slot) ? `<div class="p-notes"><span class="p-label">Anotações sobre esta peça</span><div class="p-canvas" style="height:55mm"></div></div>` : ""}`;
}

const skinReference = (item) => `${skinCover(item)}${skinBody(item)}`;

function skinDoc(item) {
  const cat = CATEGORY[item.slot];
  return htmlDoc({
    title: item.name,
    cover: false,
    body: `${skinReference(item)}
<div class="p-break"></div>
<h2 class="p-h2">Propor uma peça nova de ${esc(cat.title.toLowerCase())}</h2>
<p class="p-note">A página anterior mostra a peça atual no jogo. Use a ficha abaixo para a sua proposta: pode manter, trocar ou acrescentar, desde que respeite as travas.</p>
<h3 class="p-h3">Travas desta categoria</h3>
<ul class="p-list">${CHECKLIST[item.slot].map((t) => `<li><b>✓</b><span>${t}</span></li>`).join("")}</ul>
<h3 class="p-h3">Ficha da proposta</h3>
${FICHA[item.slot](true)}
`,
    footer: `Skin ${item.slot}/${item.id}`,
  });
}

function singleDoc() {
  const sections = SLOTS.map((slot) => {
    const refs = itemsOf(slot)
      .map((item) => `<div class="p-break"></div>${skinReference(item)}`)
      .join("");
    return `${refs}${FICHA[slot]()}`;
  }).join("");
  return htmlDoc({
    title: "Skins do artista",
    cover: false,
    body: `${cover(
      "Hoodwink · Camarim · guia geral",
      "Skins: guia do artista",
      "O que é uma skin, o que muda em cada categoria, as travas do jogo e como preencher a ficha.",
    )}${guideBody()}${sections}`,
    footer: "Caderno de skins · Hoodwink",
  });
}

/* ────────────────────────────── build ────────────────────────────── */

mkdirSync(join(BUILD, "assets/cartas"), { recursive: true });
for (const f of ["verso-claro.svg", "verso-escuro.svg"]) {
  copyFileSync(join(WEB, "public/assets/cartas", f), join(BUILD, "assets/cartas", f));
}

function print(html, name) {
  const file = join(BUILD, `${name}.html`);
  writeFileSync(file, html);
  const pdf = join(REPO, `${name}.pdf`);
  execFileSync(
    CHROME,
    [
      "--headless=new",
      "--disable-gpu",
      "--no-sandbox",
      "--hide-scrollbars",
      "--allow-file-access-from-files",
      "--run-all-compositor-stages-before-draw",
      "--virtual-time-budget=12000",
      "--no-pdf-header-footer",
      `--user-data-dir=${join(BUILD, "profile")}`,
      `--print-to-pdf=${pdf}`,
      `file:///${file.replaceAll("\\", "/")}`,
    ],
    { stdio: "pipe" },
  );
  console.log(`✓ ${name}.pdf`);
}

const OLD = [
  "skins-00-guia",
  "skins-01-cores",
  "skins-02-letras",
  "skins-03-molduras",
  "skins-04-botoes",
  "skins-05-cartas",
  "skins-06-ficha-em-branco",
];

const STALE = [
  ...OLD,
  ...ITEMS.map((item) => `skin-${item.slot}-${item.id}`),
];

print(singleDoc(), "skins-do-artista");

// Remove os formatos antigos depois que o caderno novo passou.
for (const name of STALE) {
  const old = join(REPO, `${name}.pdf`);
  if (existsSync(old)) rmSync(old);
}

console.log(`\nPDF na raiz: ${join(REPO, "skins-do-artista.pdf")}`);
