import { isValidElement, type ReactElement, type ReactNode } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import { HowToBody } from "./HowTo";
import { GROUPS, sheetOf } from "./sheet";

const noop = () => {};
const body = (over: Partial<Parameters<typeof HowToBody>[0]> = {}) =>
  HowToBody({ layout: "mobile", tax: 0, role: "GUARDIAN", idle: true, onTax: noop, onRole: noop, onIdle: noop, ...over });
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
  it("renderiza as seções do canvas e a barra com a chave Movimento", () => {
    const out = html();
    for (const t of [
      "Como se joga", "Modo Clássico", "Resumo", "Como é uma rodada", "Ações do turno", "Quando alguém duvida",
      "A taxa mexe nos preços", "Os 12 personagens", "Outras regras", "Fim da partida", "Movimento: ligado", "Criar sala",
    ]) {
      expect(out).toContain(t);
    }
    expect(out).toContain("m-canhoto");
    expect(out).toContain('data-arte="/arte/papeis/GUARDIAN/"');
    expect(out).toContain("arte em breve");
  });

  it("os 12 personagens aparecem uma vez, em cinco grupos", () => {
    expect(GROUPS.flat()).toHaveLength(12);
    expect(new Set(GROUPS.flat()).size).toBe(12);
    expect(html().match(/aria-pressed=/g)).toHaveLength(12);
  });

  it("troca de personagem: a ficha e a etiqueta marcada seguem a escolha", () => {
    const out = html({ role: "TRICKSTER" });
    expect(out).toMatch(/aria-pressed="true"[^>]*>(<span[^>]*><\/span>)?<span[^>]*>Trapaceiro</);
    expect(out).toContain("Tira 2 moedas de outro jogador e fica com 1.");
    const onRole = vi.fn();
    find(body({ onRole }), (p) => p.onClick !== undefined && find(p.children as ReactNode, (q) => q.children === "Coveiro") !== null)!.props.onClick!();
    expect(onRole).toHaveBeenCalledWith("GRAVEDIGGER");
  });

  it("os preços mudam com a taxa de teste", () => {
    expect(sheetOf("ASSASSIN", 0).price).toBe("3");
    expect(sheetOf("ASSASSIN", 3).price).toBe("6");
    expect(sheetOf("ASSASSIN", -5).price).toBe("2");
    expect(sheetOf("TRILLIONAIRE", -5).price).toBe("+1");
    expect(sheetOf("GUARDIAN", 0).faixa).toBe("Reação: 5 · contra Clarividente: 4 · no seu turno: 6");
    expect(sheetOf("SELFISH", 2).efeito).toBe("Por 1, impede alguém de receber 4 ou mais moedas. Por 7, impede de receber uma carta.");
    expect(html({ tax: 3 })).toContain("taxa de agora: +3.");
    expect(html({ tax: 4 })).toMatch(/>8<\/span>/); // Matar com taxa +4
  });

  it("a chave alterna aria-checked e grava a preferência de espera", () => {
    expect(html({ idle: true })).toContain('role="switch" aria-checked="true"');
    expect(html({ idle: false })).toContain("Movimento: desligado");
    expect(html({ idle: false })).toContain("_pausa_");
    const onIdle = vi.fn();
    find(body({ idle: true, onIdle }), (p) => p.role === "switch")!.props.onClick!();
    expect(onIdle).toHaveBeenCalledWith(false);
  });
});
