import { isValidElement, type ReactElement, type ReactNode } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import { HowToBody } from "./HowTo";

const noop = () => {};
const body = (over: Partial<Parameters<typeof HowToBody>[0]> = {}) =>
  HowToBody({ layout: "mobile", tax: 0, role: "ASSASSIN", idle: true, onTax: noop, onRole: noop, onIdle: noop, ...over });
const html = (over: Partial<Parameters<typeof HowToBody>[0]> = {}) => renderToStaticMarkup(body(over));

/** Acha o primeiro elemento cujas props passam no teste (sem DOM). */
function find(node: ReactNode, ok: (p: Record<string, unknown>) => boolean): ReactElement<Record<string, unknown>> | null {
  if (Array.isArray(node)) {
    for (const n of node) {
      const r = find(n, ok);
      if (r) return r;
    }
    return null;
  }
  if (!isValidElement<Record<string, unknown>>(node)) return null;
  if (ok(node.props)) return node;
  return find(node.props.children as ReactNode, ok);
}

describe("Como se joga", () => {
  it("renderiza as seções e a barra com a chave Movimento", () => {
    const out = html();
    for (const t of ["Como se joga", "O objetivo", "Na sua vez", "Usar um personagem", "Duvidar", "A taxa", "Personagens", "Movimento", "Modo Clássico"]) {
      expect(out).toContain(t);
    }
    expect(out).toContain("m-canhoto");
    expect(out).toContain('data-arte="/arte/papeis/ASSASSIN/"');
    expect(out).toContain("arte em breve!");
  });

  it("troca de personagem: a ficha e a etiqueta marcada seguem a escolha", () => {
    const out = html({ role: "TRICKSTER" });
    expect(out).toContain('aria-pressed="true" class="_etiqueta_');
    expect(out).toMatch(/aria-pressed="true"[^>]*>Trapaceiro</);
    expect(out).toContain("Tira 2 moedas de outro jogador e fica com 1.");
    const onRole = vi.fn();
    find(body({ onRole }), (p) => p.children === "Coveiro")!.props.onClick!();
    expect(onRole).toHaveBeenCalledWith("GRAVEDIGGER");
  });

  it("os preços mudam com a taxa de teste", () => {
    expect(html({ tax: 0 })).toContain("Paga 3 e mata uma carta de outro jogador.");
    expect(html({ tax: 3 })).toContain("Paga 6 e mata uma carta de outro jogador.");
    expect(html({ tax: 0 })).toContain("Você paga 6 e mata");
    expect(html({ tax: 5 })).toContain("Você paga 8 e mata");
    expect(html({ tax: -5 })).toContain("Você recebe 0 moedas.");
    expect(html({ tax: 3 })).toContain("Taxa +3");
  });

  it("a chave alterna aria-checked e grava a preferência de espera", () => {
    expect(html({ idle: true })).toContain('role="switch" aria-checked="true"');
    expect(html({ idle: false })).toContain('role="switch" aria-checked="false"');
    expect(html({ idle: false })).toContain("_pausa_");
    const onIdle = vi.fn();
    find(body({ idle: true, onIdle }), (p) => p.role === "switch")!.props.onClick!();
    expect(onIdle).toHaveBeenCalledWith(false);
  });
});
