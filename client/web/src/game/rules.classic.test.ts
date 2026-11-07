/**
 * Contrato dos números clássicos: `rules.classic.ts` é uma cópia à mão do
 * ClassicRules.yaml (fonte da verdade no servidor). Se o YAML mudar e a cópia
 * ficar para trás, o preço mostrado não bate com o cobrado — e o servidor
 * recusa a jogada em silêncio (a tela de escolha de ação reabria no Assassino).
 * Este teste lê o YAML direto do repositório e compara campo a campo.
 */
import { readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { CLASSIC_RULES, type CardRule } from "./rules.classic";

const here = dirname(fileURLToPath(import.meta.url));
const yamlPath = resolve(here, "../../../../server/gameServer/game/rules/modes/ClassicRules.yaml");

type Yaml = { [key: string]: Yaml | number | boolean };

/** Parser mínimo para este YAML: mapas aninhados por indentação e escalares (`chave: valor`). */
function parseYaml(src: string): Yaml {
  const root: Yaml = {};
  const stack: { indent: number; node: Yaml }[] = [{ indent: -1, node: root }];
  for (const raw of src.split("\n")) {
    const line = raw.replace(/#.*$/, "").trimEnd();
    if (!line.trim()) continue;
    const match = line.trim().match(/^([\w-]+):\s*(.*)$/);
    if (!match) continue;
    const [, key, value] = match;
    const indent = line.length - line.trimStart().length;
    while (stack.length > 1 && stack[stack.length - 1].indent >= indent) stack.pop();
    const parent = stack[stack.length - 1].node;
    if (value.trim() === "") {
      const child: Yaml = {};
      parent[key] = child;
      stack.push({ indent, node: child });
      continue;
    }
    parent[key] = scalar(value);
  }
  return root;
}

function scalar(text: string): number | boolean {
  const value = text.trim();
  if (value === "true") return true;
  if (value === "false") return false;
  const number = Number(value);
  if (Number.isNaN(number)) throw new Error(`valor não numérico no YAML: "${value}"`);
  return number;
}

const yaml = parseYaml(readFileSync(yamlPath, "utf8"));

/** Lê um caminho `A.B.C` do YAML; campo ausente vira `undefined`. */
function at(path: string): unknown {
  return path
    .split(".")
    .reduce<unknown>((node, key) => (node as Yaml | undefined)?.[key], yaml);
}

const num = (path: string): number => {
  const value = at(path);
  if (typeof value !== "number") throw new Error(`YAML sem número em ${path}`);
  return value;
};

describe("ClassicRules.yaml × rules.classic.ts", () => {
  it("tem exatamente as mesmas cartas", () => {
    expect(Object.keys(CLASSIC_RULES.cards).sort()).toEqual(Object.keys(at("Cards") as Yaml).sort());
  });

  it("cada carta copia os mesmos campos do YAML", () => {
    const fields = ["value", "fixedValue", "taxMinimum", "taxMaximum", "roundsUntilDecrease", "maxDoubled", "affectedByTaxes"] as const satisfies readonly (keyof CardRule)[];
    for (const [id, rule] of Object.entries(CLASSIC_RULES.cards)) {
      // O Investidor é modelado como investimento (stake, teto e carência), não como `cards`.
      if (id === "INVESTOR") continue;
      for (const field of fields) {
        expect(rule[field], `${id}.${field}`).toBe(at(`Cards.${id}.${field}`));
      }
    }
  });

  it("Investidor: teto afetado por taxas e carência", () => {
    expect(at("Cards.INVESTOR.affectedByTaxes")).toBe(true);
    expect(CLASSIC_RULES.investorCap).toEqual({
      value: num("Cards.INVESTOR.value"),
      taxMinimum: num("Cards.INVESTOR.taxMinimum"),
      taxMaximum: num("Cards.INVESTOR.taxMaximum"),
    });
    expect(CLASSIC_RULES.investorGrace).toBe(num("Cards.INVESTOR.gracePeriod"));
  });

  it("Guardiã: os três preços derivam dos campos extras do YAML", () => {
    const value = num("Cards.GUARDIAN.value");
    expect(CLASSIC_RULES.guardian.ownTurn).toEqual({
      value: value + num("Cards.GUARDIAN.futureProtectionSurcharge"),
      taxMinimum: num("Cards.GUARDIAN.futureProtectionTaxMinimum"),
    });
    expect(CLASSIC_RULES.guardian.clairvoyant).toEqual({
      value: value - num("Cards.GUARDIAN.clairvoyantBlockDiscount"),
      taxMinimum: num("Cards.GUARDIAN.clairvoyantBlockTaxMinimum"),
    });
  });

  it("Egoísta: limiar de moedas", () => {
    expect(CLASSIC_RULES.selfishThreshold).toBe(num("Cards.SELFISH.coinDenialThreshold"));
  });

  it("geral e ações básicas", () => {
    expect(CLASSIC_RULES.general).toEqual({
      initialCoins: num("General.initialCoins"),
      maxCoins: num("General.maxCoins"),
      maxTax: num("General.maxTax"),
      minTax: num("General.minTax"),
      maxPlayers: num("General.maxPlayers"),
    });
    expect(CLASSIC_RULES.basic).toEqual({
      takeCoins: num("General.basicTakeCoins"),
      killCost: num("General.basicKillCardCost"),
      killMinimum: num("General.basicKillCardMinimum"),
      swapMinCoins: num("General.basicSwapMinCoins"),
    });
  });

  it("deixas: padrão e morte súbita", () => {
    const windows = (key: string) => ({ default: num(`Timeouts.DEFAULT.${key}`), suddenDeath: num(`Timeouts.SUDDEN_DEATH.${key}`) });
    expect(CLASSIC_RULES.waitingAction).toEqual(windows("waitingAction"));
    expect(CLASSIC_RULES.displayMessage).toEqual(windows("displayMessage"));
    expect(CLASSIC_RULES.displayImportantMessage).toEqual(windows("displayImportantMessage"));
  });
});
