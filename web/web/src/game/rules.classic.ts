/**
 * Números de Hoodwink-server/gameServer/game/rules/modes/ClassicRules.yaml, copiados à mão.
 * Se o YAML mudar, atualize aqui (pricing.test.ts cobre a fórmula, não os números).
 */
import type { RoleId } from "./protocol";

export type CardRule = {
  value?: number;
  fixedValue?: number;
  affectedByTaxes: boolean;
  taxMinimum?: number;
  taxMaximum?: number;
  maxDoubled?: number;
  roundsUntilDecrease?: number;
};

export type Rules = {
  cards: Record<RoleId, CardRule>;
  general: { initialCoins: number; maxCoins: number; maxTax: number; minTax: number };
  /** segundos de cada deixa: padrão e em morte súbita */
  waitingAction: { default: number; suddenDeath: number };
  /** moedas que o Investidor aplica; não vem do YAML: é o "Investe 1" da prancha Acoes */
  investorStake: number;
};

export const CLASSIC_RULES: Rules = {
  cards: {
    ASSASSIN: { value: 3, taxMaximum: 9, taxMinimum: 2, affectedByTaxes: true },
    KAMIKAZE: { affectedByTaxes: false },
    POLITICAL: { fixedValue: 1, value: 2, affectedByTaxes: false, roundsUntilDecrease: 3, maxDoubled: 2 },
    REBEL: { fixedValue: 1, value: 2, affectedByTaxes: false, roundsUntilDecrease: 3, maxDoubled: 2 },
    TRILLIONAIRE: { value: 3, taxMinimum: 1, taxMaximum: 13, affectedByTaxes: true },
    CLAIRVOYANT: { value: 4, taxMaximum: 12, taxMinimum: 2, affectedByTaxes: true },
    GUARDIAN: { value: 5, taxMaximum: 15, taxMinimum: 3, affectedByTaxes: true },
    TRICKSTER: { value: 2, taxMinimum: 2, taxMaximum: 5, affectedByTaxes: true },
    GRAVEDIGGER: { value: 6, taxMaximum: 16, taxMinimum: 3, affectedByTaxes: true },
    CROUPIER: { value: 6, taxMaximum: 16, taxMinimum: 3, affectedByTaxes: true },
    INVESTOR: { fixedValue: 3, value: 4, taxMinimum: 2, taxMaximum: 6, affectedByTaxes: true },
    SELFISH: { fixedValue: 1, value: 6, taxMinimum: 3, taxMaximum: 9, affectedByTaxes: true },
  },
  general: { initialCoins: 3, maxCoins: 20, maxTax: 10, minTax: -3 },
  waitingAction: { default: 10, suddenDeath: 5 },
  investorStake: 1,
};
