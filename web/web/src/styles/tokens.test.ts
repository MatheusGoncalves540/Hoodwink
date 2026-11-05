import { readFileSync, readdirSync } from "node:fs";
import { join, relative } from "node:path";
import { describe, expect, it } from "vitest";

const SRC = join(__dirname, "..");
/** Arquivos que podem ter cor literal (ex.: arte que não muda com skin). Hoje nenhum. */
const EXCECOES = new Set<string>([]);
const COR_LITERAL = /#[0-9a-f]{3,8}\b|\b(rgba?|hsla?)\(/gi;

const modulos = (readdirSync(SRC, { recursive: true }) as string[])
  .filter((f) => f.endsWith(".module.css"))
  .map((f) => relative(SRC, join(SRC, f)).replaceAll("\\", "/"));

describe("guarda de tokens", () => {
  it("acha os módulos CSS", () => expect(modulos.length).toBeGreaterThan(10));

  it("nenhum *.module.css tem cor literal: tudo passa por --hw-*", () => {
    const achados = modulos
      .filter((f) => !EXCECOES.has(f))
      .flatMap((f) => (readFileSync(join(SRC, f), "utf8").match(COR_LITERAL) ?? []).map((c) => `${f}: ${c}`));
    expect(achados).toEqual([]);
  });

  it("texto corrido em Literata", () => {
    expect(readFileSync(join(SRC, "styles/tokens.css"), "utf8")).toMatch(/--hw-font-texto:"Literata"/);
  });
});
