import { describe, expect, it } from "vitest";
import { cardValue, priceOf } from "./pricing";
import { ROLE_IDS } from "./protocol";
import { declaredRole, ROLE_GROUPS } from "./roles";

const room = (tax: number, doubled: Record<string, number> = {}) => ({
  tax,
  doubledCardValues: Object.fromEntries(
    Object.entries(doubled).map(([k, t]) => [k, { timesValueDoubled: t, roundsUntilDecrease: 3, usedThisTurn: false }]),
  ),
});

describe("cardValue", () => {
  it("bate com a prancha Acoes (taxa +1)", () => {
    const r = room(1);
    expect(Object.fromEntries(ROLE_IDS.map((id) => [id, cardValue(id, r)]))).toEqual({
      TRILLIONAIRE: 4, TRICKSTER: 3, INVESTOR: 5, ASSASSIN: 4, CLAIRVOYANT: 5, KAMIKAZE: 0,
      GUARDIAN: 6, GRAVEDIGGER: 7, CROUPIER: 7, POLITICAL: 2, REBEL: 2, SELFISH: 7,
    });
  });

  it("aplica taxa −1, 0 e +2", () => {
    expect([-1, 0, 2].map((t) => cardValue("ASSASSIN", room(t)))).toEqual([2, 3, 5]);
    expect([-1, 0, 2].map((t) => cardValue("GUARDIAN", room(t)))).toEqual([4, 5, 7]);
  });

  it("prende nos limites taxMinimum e taxMaximum", () => {
    expect(cardValue("ASSASSIN", room(-3))).toBe(2); // 3 − 3 = 0 < mínimo 2
    expect(cardValue("ASSASSIN", room(10))).toBe(9); // 13 > máximo 9
    expect(cardValue("TRICKSTER", room(10))).toBe(5);
  });

  it("dobra Político e Rebelde 1 e 2 vezes, sem taxa", () => {
    expect(cardValue("POLITICAL", room(2, { POLITICAL: 1 }))).toBe(4);
    expect(cardValue("REBEL", room(-1, { REBEL: 2 }))).toBe(8);
  });

  it("desconto nunca deixa negativo", () => {
    expect(cardValue("KAMIKAZE", room(1), undefined, 5)).toBe(0);
  });
});

describe("priceOf", () => {
  it("separa custo, ganho, investimento e reação", () => {
    const r = room(1);
    expect(priceOf("ASSASSIN", r)).toEqual({ kind: "cost", coins: 4 });
    expect(priceOf("TRILLIONAIRE", r)).toEqual({ kind: "gain", coins: 4 });
    expect(priceOf("TRICKSTER", r)).toEqual({ kind: "gain", coins: 2, steals: 3 });
    expect(priceOf("INVESTOR", r)).toEqual({ kind: "invest", coins: 1, lasts: 5 });
    expect(priceOf("SELFISH", r)).toEqual({ kind: "reaction", coins: 1, investment: 2, cardChange: 7 });
  });
});

describe("roles", () => {
  it("cada papel aparece em exatamente um grupo", () => {
    expect(ROLE_GROUPS.flatMap((g) => g.roles).sort()).toEqual([...ROLE_IDS].sort());
  });

  it("lê o papel de CARD_PLAYED_*", () => {
    expect(declaredRole("CARD_PLAYED_ASSASSIN")).toBe("ASSASSIN");
    expect(declaredRole("CONTEST")).toBeNull();
    expect(declaredRole("CARD_PLAYED_NADA")).toBeNull();
  });
});
