/**
 * Preço de cada personagem e das ações do turno. Regras v2 (plans/02-regras-v2.md §2 e §3).
 *
 *   v = value << timesValueDoubled            (só Político e Rebelde têm dobra)
 *   se affectedByTaxes: v + taxa, preso em [taxMinimum, taxMaximum]
 *   v = max(0, v − desconto)
 *
 * O mesmo número significa coisas diferentes por personagem:
 * - custo pago: Assassino, Clarividente, Guardiã, Coveiro, Crupiê, Político, Rebelde, Kamikaze (0);
 * - ganho: Trilionário recebe v; Trapaceiro tira X = clamp(2 + t, 2, 5) e fica com ⌈X/2⌉;
 * - Investidor: investe `investorStake` por turno, até o teto clamp(4 + t, 2, 6); a moeda volta em 6 turnos;
 * - Egoísta (reação): barrar moedas custa fixedValue (só se o ganho for ≥ 2 + t); barrar troca de carta, v;
 *   barrar investimento, v com desconto 5 (ponytail: o 02 não dá preço para isso; segue o servidor).
 */
import type { Doubled, RoleId } from "./protocol";
import { CLASSIC_RULES, type Rules, type Taxed } from "./rules.classic";

export type PriceRoom = { tax: number; doubledCardValues: Record<string, Doubled> };

export type Price =
  | { kind: "cost"; coins: number }
  | { kind: "gain"; coins: number; steals?: number }
  | { kind: "invest"; coins: number; lasts: number; cap: number }
  | { kind: "reaction"; coins: number; investment: number; cardChange: number; threshold: number };

export const taxed = (b: Taxed, tax: number) =>
  Math.min(b.taxMaximum ?? Infinity, Math.max(b.taxMinimum ?? -Infinity, b.value + tax));

export function cardValue(role: RoleId, room: PriceRoom, rules: Rules = CLASSIC_RULES, discount = 0): number {
  const r = rules.cards[role];
  let v = (r.value ?? 0) << (room.doubledCardValues[role]?.timesValueDoubled ?? 0);
  if (r.affectedByTaxes) v = taxed({ ...r, value: v }, room.tax);
  return Math.max(0, v - discount);
}

/** Guardiã: reação a um ataque, contra o Clarividente, ou no seu turno (protege o próximo ataque). */
export function guardianPrice(when: "reaction" | "clairvoyant" | "ownTurn", room: PriceRoom, rules: Rules = CLASSIC_RULES): number {
  return when === "reaction" ? cardValue("GUARDIAN", room, rules) : taxed(rules.guardian[when], room.tax);
}

export function priceOf(role: RoleId, room: PriceRoom, rules: Rules = CLASSIC_RULES): Price {
  const v = cardValue(role, room, rules);
  switch (role) {
    case "TRILLIONAIRE":
      return { kind: "gain", coins: v };
    case "TRICKSTER":
      return { kind: "gain", coins: Math.ceil(v / 2), steals: v };
    case "INVESTOR":
      return { kind: "invest", coins: rules.investorStake, lasts: rules.investorGrace, cap: taxed(rules.investorCap, room.tax) };
    case "GUARDIAN":
      return { kind: "cost", coins: guardianPrice("ownTurn", room, rules) };
    case "SELFISH":
      return {
        kind: "reaction",
        coins: rules.cards.SELFISH.fixedValue ?? 0,
        investment: cardValue(role, room, rules, 5),
        cardChange: v,
        threshold: rules.selfishThreshold + room.tax,
      };
    default:
      return { kind: "cost", coins: v };
  }
}

export const BASIC_ACTIONS = ["TAKE_COINS", "KILL_CARD", "SWAP_CARDS"] as const;
export type BasicAction = (typeof BASIC_ACTIONS)[number];

/** Matar: taxa positiva entra pela metade (para baixo), zero ou negativa entra inteira; mínimo 3. */
export const killCost = (tax: number, rules: Rules = CLASSIC_RULES) =>
  Math.max(rules.basic.killMinimum, rules.basic.killCost + (tax > 0 ? Math.floor(tax / 2) : tax));

/** Trocar: ⌈moedas/2⌉ por uma carta; as duas custam todas as moedas. */
export const swapCost = (coins: number, both = false) => (both ? coins : Math.ceil(coins / 2));

/**
 * Preço das ações do turno. Pegar moeda: 1 + t quando t > 0; taxa negativa não reduz
 * (mínimo o valor-base, nunca 0).
 */
export function basicPrice(action: BasicAction, room: PriceRoom, coins: number, rules: Rules = CLASSIC_RULES): Price {
  switch (action) {
    case "TAKE_COINS":
      return { kind: "gain", coins: rules.basic.takeCoins + Math.max(0, room.tax) };
    case "KILL_CARD":
      return { kind: "cost", coins: killCost(room.tax, rules) };
    case "SWAP_CARDS":
      return { kind: "cost", coins: swapCost(coins) };
  }
}

/** Trocar exige pelo menos 6 moedas. */
export const canSwap = (coins: number, rules: Rules = CLASSIC_RULES) => coins >= rules.basic.swapMinCoins;

/** Moedas que saem do bolso para declarar no seu turno (0 para ganhos). */
export function costOf(price: Price): number {
  return price.kind === "gain" ? 0 : price.coins;
}
