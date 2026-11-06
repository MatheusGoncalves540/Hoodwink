import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";
import { COLOR_TOKENS, ITEMS, PACKS, SLOTS, itemOf } from "./catalog";

const DIR = __dirname;
const css = (slot: string, id: string) => readFileSync(join(DIR, slot, `${id}.css`), "utf8");
const withFile = ITEMS.filter((i) => i.load);
/** Sem comentários e sem @import url(...). */
const clean = (src: string) => src.replace(/\/\*[\s\S]*?\*\//g, "").replace(/@import\s+url\([^)]*\)[^;]*;/g, "");
/** Seletores de topo, separados por vírgula. */
const selectors = (src: string) =>
  [...clean(src).matchAll(/([^{}]+)\{[^{}]*\}/g)].flatMap((m) => m[1].split(",").map((s) => s.trim()));
/** Declarações do bloco cujo seletor é exatamente `sel`. */
function vars(src: string, sel: string): Record<string, string> {
  const m = [...clean(src).matchAll(/([^{}]+)\{([^{}]*)\}/g)].find((x) => x[1].trim() === sel);
  return Object.fromEntries((m?.[2] ?? "").split(";").filter(Boolean).map((d) => [d.slice(0, d.indexOf(":")).trim(), d.slice(d.indexOf(":") + 1).trim()]));
}
const lum = (hex: string) => {
  const [r, g, b] = [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255).map((c) => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
};
const contrast = (a: string, b: string) => {
  const [x, y] = [lum(a), lum(b)].sort((p, q) => q - p);
  return (x + 0.05) / (y + 0.05);
};

describe("catálogo de skins", () => {
  it("todo pacote tem as 5 categorias e só ids que existem", () => {
    for (const p of PACKS) {
      expect(Object.keys(p.equip).sort()).toEqual([...SLOTS].sort());
      for (const s of SLOTS) expect(itemOf(s, p.equip[s]), `${p.id}.${s}`).toBeTruthy();
    }
  });

  it("toda peça com load tem o arquivo CSS; a peça padrão não carrega nada", () => {
    for (const i of withFile) expect(existsSync(join(DIR, i.slot, `${i.id}.css`)), `${i.slot}/${i.id}`).toBe(true);
    for (const s of SLOTS) expect(itemOf(s, PACKS[0].equip[s])?.load).toBeUndefined();
  });

  it("todo seletor começa com [data-sk-<slot>=\"<id>\"] (nada vaza para fora da peça)", () => {
    for (const i of withFile) {
      const prefix = `[data-sk-${i.slot}="${i.id}"]`;
      const sels = selectors(css(i.slot, i.id));
      expect(sels.length, `${i.slot}/${i.id}`).toBeGreaterThan(0);
      for (const s of sels) expect(s.startsWith(prefix), `${i.slot}/${i.id}: ${s}`).toBe(true);
    }
  });

  it("molduras e botões valem só no contêiner de skin mais próximo (prévia aninhada não herda a peça de fora)", () => {
    for (const i of withFile.filter((x) => x.slot === "molduras" || x.slot === "botoes")) {
      const prefix = `[data-sk-${i.slot}="${i.id}"]`;
      for (const s of selectors(css(i.slot, i.id))) {
        expect(s, `${i.slot}/${i.id}`).toContain(`:not(${prefix} [data-sk-${i.slot}]:not(${prefix}) `);
      }
    }
  });

  it("peça de cores define todos os COLOR_TOKENS no claro e no escuro, com contraste legível", () => {
    for (const i of withFile.filter((x) => x.slot === "cores")) {
      const src = css("cores", i.id);
      const base = `[data-sk-cores="${i.id}"]`;
      for (const sel of [base, `${base}[data-theme="dark"]`]) {
        const v = vars(src, sel);
        expect(Object.keys(v).sort(), `${i.id} ${sel}`).toEqual([...COLOR_TOKENS].sort());
        expect(contrast(v["--hw-tinta"], v["--hw-plateia"]), `${i.id} ${sel} tinta`).toBeGreaterThanOrEqual(4.5);
        expect(contrast(v["--hw-titulo"], v["--hw-plateia"]), `${i.id} ${sel} titulo`).toBeGreaterThanOrEqual(3);
      }
    }
  });

  it("nenhuma peça redefine --hw-font-texto (texto corrido fica em Literata)", () => {
    for (const i of withFile) expect(css(i.slot, i.id), `${i.slot}/${i.id}`).not.toContain("--hw-font-texto");
  });
});
