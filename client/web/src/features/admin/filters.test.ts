import { describe, expect, it } from "vitest";
import {
  DAY_MS,
  MAX_ARCHIVED_RANGE_DAYS,
  MAX_LIMIT,
  MAX_PAGE,
  MAX_RANGE_DAYS,
  clampLimit,
  clampPage,
  customRange,
  isUuid,
  pageCount,
  parseStatusInput,
  presetRange,
  toLocalInput,
} from "./filters";

describe("presetRange", () => {
  const now = new Date("2026-11-06T12:00:00.000Z");

  it.each([
    ["24h", 1],
    ["7d", 7],
    ["30d", 30],
  ] as const)("volta %s dias a partir de agora", (period, days) => {
    const range = presetRange(period, now);
    expect(range.to).toBe(now.toISOString());
    expect(Date.parse(range.to) - Date.parse(range.from)).toBe(days * DAY_MS);
  });
});

describe("customRange", () => {
  const day = (n: number) => new Date(Date.UTC(2026, 0, n)).toISOString();

  it("devolve ISO quando o período é válido", () => {
    const result = customRange(day(1), day(2));
    expect(result).toEqual({ range: { from: day(1), to: day(2) } });
  });

  it("aceita exatamente o intervalo máximo de 31 dias", () => {
    const result = customRange(day(1), day(1 + MAX_RANGE_DAYS));
    expect(result.error).toBeUndefined();
    expect(result.range).toBeDefined();
  });

  it("recusa 31 dias e 1 milissegundo", () => {
    const to = new Date(Date.parse(day(1)) + MAX_RANGE_DAYS * DAY_MS + 1).toISOString();
    expect(customRange(day(1), to)).toEqual({ error: "O período máximo é de 31 dias." });
  });

  it("aceita até 183 dias nos arquivados", () => {
    const result = customRange(day(1), day(1 + MAX_ARCHIVED_RANGE_DAYS), true);
    expect(result.error).toBeUndefined();
    expect(result.range).toBeDefined();
  });

  it("recusa 184 dias nos arquivados", () => {
    const to = new Date(Date.parse(day(1)) + MAX_ARCHIVED_RANGE_DAYS * DAY_MS + 1).toISOString();
    expect(customRange(day(1), to, true)).toEqual({ error: "O período máximo é de 183 dias." });
  });

  it("recusa início depois do fim", () => {
    expect(customRange(day(3), day(2))).toEqual({ error: "O início precisa vir antes do fim." });
  });

  it("recusa campo vazio e data inválida", () => {
    expect(customRange("", day(2))).toEqual({ error: "Informe o início e o fim do período." });
    expect(customRange(day(1), "ontem")).toEqual({ error: "Data inválida." });
  });
});

describe("parseStatusInput", () => {
  it.each(["1xx", "4xx", "5xx", "200", "599"])("aceita %s", (input) => {
    expect(parseStatusInput(input).error).toBeUndefined();
  });

  it("aceita vazio como sem filtro", () => {
    expect(parseStatusInput("  ")).toEqual({ value: "" });
  });

  it("normaliza a faixa para minúsculas e o inteiro sem zeros à esquerda", () => {
    expect(parseStatusInput(" 4XX ")).toEqual({ value: "4xx" });
    expect(parseStatusInput("0400")).toEqual({ value: "400" });
  });

  it.each(["abc", "4x", "99", "600", "200.5", "6xx"])("recusa %s", (input) => {
    expect(parseStatusInput(input)).toEqual({ error: "Use 4xx ou um número de 100 a 599." });
  });
});

describe("isUuid", () => {
  it("aceita UUID com hífens, em qualquer caixa", () => {
    expect(isUuid("6f1c1d3e-2a4b-4c5d-8e9f-0123456789ab")).toBe(true);
    expect(isUuid("6F1C1D3E-2A4B-4C5D-8E9F-0123456789AB")).toBe(true);
  });

  it.each(["", "abc", "6f1c1d3e2a4b4c5d8e9f0123456789ab", "6f1c1d3e-2a4b-4c5d-8e9f-0123456789a"])(
    "recusa %s",
    (input) => {
      expect(isUuid(input)).toBe(false);
    },
  );
});

describe("clampPage", () => {
  it("mantém entre 1 e 1000", () => {
    expect(clampPage(0)).toBe(1);
    expect(clampPage(-3)).toBe(1);
    expect(clampPage(1.9)).toBe(1);
    expect(clampPage(MAX_PAGE)).toBe(MAX_PAGE);
    expect(clampPage(MAX_PAGE + 1)).toBe(MAX_PAGE);
  });

  it("valor não finito cai na página 1", () => {
    expect(clampPage(Number.NaN)).toBe(1);
    expect(clampPage(Number.POSITIVE_INFINITY)).toBe(1);
  });
});

describe("clampLimit", () => {
  it("mantém entre 1 e 200", () => {
    expect(clampLimit(0)).toBe(1);
    expect(clampLimit(-10)).toBe(1);
    expect(clampLimit(100.9)).toBe(100);
    expect(clampLimit(MAX_LIMIT)).toBe(MAX_LIMIT);
    expect(clampLimit(MAX_LIMIT + 1)).toBe(MAX_LIMIT);
  });

  it("valor não finito cai no padrão", () => {
    expect(clampLimit(Number.NaN)).toBe(50);
    expect(clampLimit(Number.POSITIVE_INFINITY)).toBe(50);
  });
});

describe("pageCount", () => {
  it("nunca é menor que 1", () => {
    expect(pageCount(0, 50)).toBe(1);
    expect(pageCount(1, 50)).toBe(1);
  });

  it("arredonda para cima", () => {
    expect(pageCount(50, 50)).toBe(1);
    expect(pageCount(51, 50)).toBe(2);
    expect(pageCount(101, 50)).toBe(3);
  });

  it("respeita o teto de páginas do backend", () => {
    expect(pageCount(10_000_000, 1)).toBe(MAX_PAGE);
  });
});

describe("toLocalInput", () => {
  it("formata no fuso local sem segundos", () => {
    const local = new Date(2026, 0, 2, 3, 4, 5);
    expect(toLocalInput(local.toISOString())).toBe("2026-01-02T03:04");
  });

  it("data inválida vira string vazia", () => {
    expect(toLocalInput("não é data")).toBe("");
  });
});
