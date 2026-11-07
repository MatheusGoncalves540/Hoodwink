/**
 * Números de server/gameServer/game/rules/modes/ClassicRules.yaml, copiados à mão (regras v2, plans/02-regras-v2.md).
 * Se o YAML mudar, atualize aqui: `rules.classic.test.ts` compara cópia e YAML campo a campo.
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

/** Valor-base que soma a taxa e fica preso entre mínimo e máximo. */
export type Taxed = { value: number; taxMinimum?: number; taxMaximum?: number };

export type Rules = {
  cards: Record<RoleId, CardRule>;
  general: { initialCoins: number; maxCoins: number; maxTax: number; minTax: number; maxPlayers: number };
  /** Ações do turno (02 §2). */
  basic: { takeCoins: number; killCost: number; killMinimum: number; swapMinCoins: number };
  /** Guardiã: a reação usa cards.GUARDIAN; estes são os outros dois preços (02 §3). */
  guardian: { clairvoyant: Taxed; ownTurn: Taxed };
  /** Egoísta barra moedas quando o ganho é pelo menos `selfishThreshold + t`. */
  selfishThreshold: number;
  /** Investidor: teto do total investido e turnos até a moeda voltar. */
  investorCap: Taxed;
  investorGrace: number;
  /** segundos de cada deixa: padrão e em morte súbita */
  waitingAction: { default: number; suddenDeath: number };
  displayMessage: { default: number; suddenDeath: number };
  displayImportantMessage: { default: number; suddenDeath: number };
  /** moedas que o Investidor aplica por turno */
  investorStake: number;
};

export const CLASSIC_RULES: Rules = {
  cards: {
    ASSASSIN: { value: 4, taxMinimum: 3, affectedByTaxes: true },
    KAMIKAZE: { affectedByTaxes: false },
    POLITICAL: { fixedValue: 1, value: 2, affectedByTaxes: false, roundsUntilDecrease: 3, maxDoubled: 2 },
    REBEL: { fixedValue: 1, value: 2, affectedByTaxes: false, roundsUntilDecrease: 3, maxDoubled: 2 },
    TRILLIONAIRE: { value: 3, taxMinimum: 1, affectedByTaxes: true },
    CLAIRVOYANT: { value: 4, taxMinimum: 2, affectedByTaxes: true },
    GUARDIAN: { value: 5, taxMinimum: 3, affectedByTaxes: true },
    TRICKSTER: { value: 2, taxMinimum: 2, taxMaximum: 5, affectedByTaxes: true },
    GRAVEDIGGER: { value: 5, taxMinimum: 3, affectedByTaxes: true },
    CROUPIER: { value: 6, taxMinimum: 4, affectedByTaxes: true },
    INVESTOR: { value: 1, affectedByTaxes: false },
    SELFISH: { fixedValue: 1, value: 5, taxMinimum: 3, taxMaximum: 8, affectedByTaxes: true },
  },
  general: { initialCoins: 3, maxCoins: 20, maxTax: 5, minTax: -5, maxPlayers: 10 },
  basic: { takeCoins: 1, killCost: 6, killMinimum: 3, swapMinCoins: 6 },
  guardian: { clairvoyant: { value: 4, taxMinimum: 2 }, ownTurn: { value: 6, taxMinimum: 5 } },
  selfishThreshold: 2,
  investorCap: { value: 4, taxMinimum: 2, taxMaximum: 6 },
  investorGrace: 6,
  waitingAction: { default: 30, suddenDeath: 15 },
  displayMessage: { default: 6, suddenDeath: 3 },
  displayImportantMessage: { default: 8, suddenDeath: 4 },
  investorStake: 1,
};
