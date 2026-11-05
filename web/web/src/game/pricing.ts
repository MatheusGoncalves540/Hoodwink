/**
 * Preço de cada papel. Fórmula do servidor (game/structs/rooms/roomCards.go → GetCardValue):
 *
 *   v = value << timesValueDoubled            (só Político e Rebelde têm dobra)
 *   se affectedByTaxes: v + taxa, preso em [taxMinimum, taxMaximum]
 *                       (passou do limite: vira o próprio limite)
 *   v = max(0, v − desconto)
 *
 * O mesmo número significa coisas diferentes por papel (protocols/*.go):
 * - custo pago: Assassino, Clarividente, Guardiã, Coveiro, Crupiê, Político, Rebelde, Kamikaze (0);
 * - ganho: Trilionário recebe v; Trapaceiro rouba v e fica com v − 1 (prancha Acoes: "Rouba 3 e fica com 2";
 *   ponytail: o servidor não mostra onde some a moeda, confirmar);
 * - Investidor: paga `investorStake` e o investimento dura v (gracePeriod);
 * - Egoísta (reação, selfish.go): barrar moedas custa fixedValue; barrar investimento, v com desconto 5;
 *   barrar troca de carta, v.
 */
import type { Doubled, RoleId } from "./protocol";
import { CLASSIC_RULES, type Rules } from "./rules.classic";

export type PriceRoom = { tax: number; doubledCardValues: Record<string, Doubled> };

export type Price =
  | { kind: "cost"; coins: number }
  | { kind: "gain"; coins: number; steals?: number }
  | { kind: "invest"; coins: number; lasts: number }
  | { kind: "reaction"; coins: number; investment: number; cardChange: number };

export function cardValue(role: RoleId, room: PriceRoom, rules: Rules = CLASSIC_RULES, discount = 0): number {
  const r = rules.cards[role];
  let v = (r.value ?? 0) << (room.doubledCardValues[role]?.timesValueDoubled ?? 0);
  if (r.affectedByTaxes) {
    const max = r.taxMaximum ?? Infinity;
    const min = r.taxMinimum ?? -Infinity;
    v = v + room.tax > max ? max : v + room.tax < min ? min : v + room.tax;
  }
  return Math.max(0, v - discount);
}

export function priceOf(role: RoleId, room: PriceRoom, rules: Rules = CLASSIC_RULES): Price {
  const v = cardValue(role, room, rules);
  switch (role) {
    case "TRILLIONAIRE":
      return { kind: "gain", coins: v };
    case "TRICKSTER":
      return { kind: "gain", coins: Math.max(0, v - 1), steals: v };
    case "INVESTOR":
      return { kind: "invest", coins: rules.investorStake, lasts: v };
    case "SELFISH":
      return {
        kind: "reaction",
        coins: rules.cards.SELFISH.fixedValue ?? 0,
        investment: cardValue(role, room, rules, 5),
        cardChange: v,
      };
    default:
      return { kind: "cost", coins: v };
  }
}

/** Moedas que saem do bolso para declarar no seu turno (0 para ganhos). */
export function costOf(price: Price): number {
  return price.kind === "gain" ? 0 : price.coins;
}
