import { describe, expect, it } from "vitest";
import { basicPrice, canSwap, cardValue, guardianPrice, killCost, priceOf, swapCost } from "./pricing";
import { ROLE_IDS } from "./protocol";
import { declaredRole, ROLE_GROUPS } from "./roles";

const room = (tax: number, doubled: Record<string, number> = {}) => ({
  tax,
  doubledCardValues: Object.fromEntries(
    Object.entries(doubled).map(([k, t]) => [k, { timesValueDoubled: t, roundsUntilDecrease: 3, usedThisTurn: false }]),
  ),
});
const TAXAS = [-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5];
const porTaxa = (f: (t: number) => number) => TAXAS.map(f);

describe("cardValue (regras v2)", () => {
  it("valores com taxa +1", () => {
    const r = room(1);
    expect(Object.fromEntries(ROLE_IDS.map((id) => [id, cardValue(id, r)]))).toEqual({
      TRILLIONAIRE: 4, TRICKSTER: 3, INVESTOR: 1, ASSASSIN: 4, CLAIRVOYANT: 5, KAMIKAZE: 0,
      GUARDIAN: 6, GRAVEDIGGER: 6, CROUPIER: 7, POLITICAL: 2, REBEL: 2, SELFISH: 6,
    });
  });

  it("mínimos da tabela 3 com taxa −5 e sem teto onde o 02 não define", () => {
    const lo = room(-5);
    expect(["ASSASSIN", "GRAVEDIGGER", "CROUPIER", "CLAIRVOYANT", "TRILLIONAIRE", "GUARDIAN", "SELFISH"].map((id) => cardValue(id as never, lo)))
      .toEqual([2, 3, 4, 2, 1, 4, 3]);
    expect(cardValue("ASSASSIN", room(5))).toBe(8);
    expect(cardValue("SELFISH", room(5))).toBe(8); // teto 8
  });

  it("Kamikaze, Político e Rebelde não sentem a taxa; Político e Rebelde dobram até 8", () => {
    expect([-5, 5].map((t) => cardValue("KAMIKAZE", room(t)))).toEqual([0, 0]);
    expect(cardValue("POLITICAL", room(5, { POLITICAL: 1 }))).toBe(4);
    expect(cardValue("REBEL", room(-5, { REBEL: 2 }))).toBe(8);
  });

  it("desconto nunca deixa negativo", () => {
    expect(cardValue("KAMIKAZE", room(1), undefined, 5)).toBe(0);
  });
});

describe("priceOf", () => {
  it("separa custo, ganho, investimento e reação (taxa +1)", () => {
    const r = room(1);
    expect(priceOf("ASSASSIN", r)).toEqual({ kind: "cost", coins: 4 });
    expect(priceOf("TRILLIONAIRE", r)).toEqual({ kind: "gain", coins: 4 });
    expect(priceOf("TRICKSTER", r)).toEqual({ kind: "gain", coins: 2, steals: 3 });
    expect(priceOf("INVESTOR", r)).toEqual({ kind: "invest", coins: 1, lasts: 6, cap: 5 });
    expect(priceOf("GUARDIAN", r)).toEqual({ kind: "cost", coins: 7 });
    expect(priceOf("SELFISH", r)).toEqual({ kind: "reaction", coins: 1, investment: 1, cardChange: 6, threshold: 3 });
  });

  it("Trapaceiro tira clamp(2 + t, 2, 5) e fica com a metade maior", () => {
    expect(porTaxa((t) => (priceOf("TRICKSTER", room(t)) as { steals: number }).steals)).toEqual([2, 2, 2, 2, 2, 2, 3, 4, 5, 5, 5]);
    expect(porTaxa((t) => priceOf("TRICKSTER", room(t)).coins)).toEqual([1, 1, 1, 1, 1, 1, 2, 2, 3, 3, 3]);
  });

  it("Investidor: teto clamp(4 + t, 2, 6)", () => {
    expect(porTaxa((t) => (priceOf("INVESTOR", room(t)) as { cap: number }).cap)).toEqual([2, 2, 2, 2, 3, 4, 5, 6, 6, 6, 6]);
  });

  it("Egoísta: limiar de moedas 2 + t", () => {
    expect([-5, 0, 5].map((t) => (priceOf("SELFISH", room(t)) as { threshold: number }).threshold)).toEqual([-3, 2, 7]);
  });
});

describe("Guardiã: três preços", () => {
  it("reação 5 + t mín 4, contra Clarividente 4 + t mín 2, no seu turno 6 + t mín 5", () => {
    expect(porTaxa((t) => guardianPrice("reaction", room(t)))).toEqual([4, 4, 4, 4, 4, 5, 6, 7, 8, 9, 10]);
    expect(porTaxa((t) => guardianPrice("clairvoyant", room(t)))).toEqual([2, 2, 2, 2, 3, 4, 5, 6, 7, 8, 9]);
    expect(porTaxa((t) => guardianPrice("ownTurn", room(t)))).toEqual([5, 5, 5, 5, 5, 6, 7, 8, 9, 10, 11]);
  });
});

describe("ações do turno", () => {
  it("Matar: tabela do 02 (mínimo 3, taxa positiva pela metade)", () => {
    expect(porTaxa((t) => killCost(t))).toEqual([3, 3, 3, 4, 5, 6, 6, 7, 7, 8, 8]);
  });

  it("Pegar moeda: 1 + t, nunca negativo", () => {
    expect(porTaxa((t) => basicPrice("TAKE_COINS", room(t), 0).coins)).toEqual([0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6]);
    expect(basicPrice("TAKE_COINS", room(0), 0)).toEqual({ kind: "gain", coins: 1 });
  });

  it("Trocar: ⌈moedas/2⌉, as duas custam tudo, precisa de 6", () => {
    expect([6, 7, 20].map((c) => swapCost(c))).toEqual([3, 4, 10]);
    expect(swapCost(7, true)).toBe(7);
    expect(basicPrice("SWAP_CARDS", room(3), 9)).toEqual({ kind: "cost", coins: 5 });
    expect([0, 5, 6].map((c) => canSwap(c))).toEqual([false, false, true]);
  });
});

describe("roles", () => {
  it("cada personagem aparece em exatamente um grupo", () => {
    expect(ROLE_GROUPS.flatMap((g) => g.roles).sort()).toEqual([...ROLE_IDS].sort());
  });

  it("lê o personagem de CARD_PLAYED_*", () => {
    expect(declaredRole("CARD_PLAYED_ASSASSIN")).toBe("ASSASSIN");
    expect(declaredRole("CONTEST")).toBeNull();
    expect(declaredRole("CARD_PLAYED_NADA")).toBeNull();
  });
});
