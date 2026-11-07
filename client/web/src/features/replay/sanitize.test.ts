import { describe, expect, it } from "vitest";
import { MAX_REPLAY_TEXT, sanitizeText } from "./sanitize";

describe("sanitizeText (texto do .hwr)", () => {
  it("corta em 32 caracteres por code point", () => {
    expect(sanitizeText("a".repeat(40))).toBe("a".repeat(MAX_REPLAY_TEXT));
    expect(Array.from(sanitizeText("😀".repeat(40)))).toHaveLength(MAX_REPLAY_TEXT);
    expect(sanitizeText("😀".repeat(40))).toBe("😀".repeat(MAX_REPLAY_TEXT));
  });

  it("tira caracteres de controle e marcas de bidi", () => {
    expect(sanitizeText("A\u0000B\u001fC\u007fD\u009fE")).toBe("ABCDE");
    expect(sanitizeText("Ana\u202ebia")).toBe("Anabia");
    expect(sanitizeText("\u2066Eva\u2069")).toBe("Eva");
    expect(sanitizeText("li\u200bnea")).toBe("linea");
    expect(sanitizeText("quebra\u2028de linha")).toBe("quebrade linha");
  });

  it("um nome com HTML continua texto, sem virar marcação", () => {
    const raw = '<img src=x onerror=alert(1)>';
    expect(sanitizeText(raw)).toBe(raw);
    expect(sanitizeText("<script>alert(1)</script>")).toBe("<script>alert(1)</script>");
  });

  it("apara as pontas e recusa entrada que não é texto", () => {
    expect(sanitizeText("  Ana  ")).toBe("Ana");
    for (const bad of [undefined, null, 42, {}, [], true]) expect(sanitizeText(bad)).toBe("");
  });

  it("max inválido não deixa passar texto", () => {
    expect(sanitizeText("Ana", 0)).toBe("");
    expect(sanitizeText("Ana", -2)).toBe("");
    expect(sanitizeText("Ana", Number.NaN)).toBe("");
    expect(sanitizeText("Ana", 4)).toBe("Ana");
  });
});
