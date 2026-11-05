import { isValidElement, type ReactElement, type ReactNode } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import { DEFAULT_EQUIP, PACKS } from "../../skins/catalog";
import { CamarimBody, preloadOf, wear, type Tab } from "./Camarim";

const noop = () => {};
const body = (tab: Tab, over: Partial<Parameters<typeof CamarimBody>[0]> = {}) =>
  CamarimBody({ layout: "mobile", theme: "light", equip: DEFAULT_EQUIP, tab, onTab: noop, onEquip: noop, ...over });
const html = (tab: Tab, over = {}) => renderToStaticMarkup(body(tab, over));

function findAll(node: ReactNode, ok: (p: Record<string, unknown>) => boolean, out: ReactElement<Record<string, unknown>>[] = []) {
  if (Array.isArray(node)) node.forEach((n) => findAll(n, ok, out));
  else if (isValidElement<Record<string, unknown>>(node)) {
    if (ok(node.props)) out.push(node);
    findAll(node.props.children as ReactNode, ok, out);
  }
  return out;
}

describe("Camarim", () => {
  it("renderiza as seis abas, a frase da aba e a prévia vestida com a escolha atual", () => {
    const out = html("pacotes");
    for (const t of ["Pacotes", "Cores", "Letras", "Molduras", "Botões", "Cartas"]) expect(out).toContain(`>${t}</button>`);
    expect(out).toContain('role="tab" aria-selected="true"');
    expect(out).toContain("Um visual pronto, com uma peça de cada tipo.");
    expect(out).toContain('aria-label="Prévia" class="_preview_');
    expect(out).toContain('data-sk-cores="holofote" data-sk-letras="cartaz"');
    expect(out).not.toMatch(/preço|comprar|cadeado/i);
  });

  it("marca a peça vestida com ✓ vestido e aria-pressed", () => {
    const out = html("cartas", { equip: { ...DEFAULT_EQUIP, cartas: "xadrez" } });
    expect(out.match(/aria-pressed="true"/g)).toHaveLength(1);
    expect(out).toMatch(/aria-pressed="true"[^>]*><span[^>]*data-sk-cartas="xadrez"[\s\S]*?Xadrez<\/span><span[^>]*>✓ vestido/);
  });

  it("tocar numa peça entrega o Equip novo; vestir chama applySkin com ele e salva", async () => {
    const onEquip = vi.fn();
    const tiles = findAll(body("cores", { onEquip }), (p) => "aria-pressed" in p);
    expect(tiles).toHaveLength(4);
    (tiles[2].props.onClick as () => void)();
    expect(onEquip).toHaveBeenCalledWith({ ...DEFAULT_EQUIP, cores: "noir" });

    const apply = vi.fn(async () => {});
    const el = {} as HTMLElement;
    await wear(PACKS[2].equip, el, apply);
    expect(apply).toHaveBeenCalledWith(el, PACKS[2].equip);
  });

  it("pacote veste as 5 peças de uma vez; abrir a aba pré-carrega as peças dela", () => {
    const onEquip = vi.fn();
    (findAll(body("pacotes", { onEquip }), (p) => "aria-pressed" in p)[1].props.onClick as () => void)();
    expect(onEquip).toHaveBeenCalledWith(PACKS[1].equip);
    expect(preloadOf("letras").map((i) => i.id)).toEqual(["cartaz", "pincel", "gala"]);
  });
});
