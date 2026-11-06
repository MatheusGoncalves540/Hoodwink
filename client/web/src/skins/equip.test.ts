import { describe, expect, it } from "vitest";
import { DEFAULT_EQUIP, PACKS, itemOf } from "./catalog";
import { equipItem, equipPack, resolveEquip } from "./equip";

describe("resolveEquip (fronteira de confiança)", () => {
  it("entrada ausente, vazia ou quebrada vira o padrão", () => {
    for (const bad of [undefined, null, "", "{nao é json", "42", "[]", {}, 7]) expect(resolveEquip(bad)).toEqual(DEFAULT_EQUIP);
  });

  it("id inexistente cai no padrão só daquela categoria; categoria extra é ignorada", () => {
    expect(resolveEquip(JSON.stringify({ cores: "noir", letras: "comic-sans", botoes: 3, extra: "x" }))).toEqual({ ...DEFAULT_EQUIP, cores: "noir" });
  });

  it("id que existe em outra categoria não vale (papelao é carta e moldura, não cor)", () => {
    expect(resolveEquip({ cores: "papelao", molduras: "papelao" })).toEqual({ ...DEFAULT_EQUIP, molduras: "papelao" });
  });

  it("entrada válida passa inteira", () => {
    const ok = { cores: "bastidor", letras: "pincel", molduras: "papelao", botoes: "placa", cartas: "papelao" };
    expect(resolveEquip(JSON.stringify(ok))).toEqual(ok);
  });
});

describe("equipPack e equipItem", () => {
  it("devolvem objeto novo sem mexer na entrada", () => {
    const pack = PACKS[1];
    const before = JSON.stringify(pack);
    const e = equipPack(pack);
    expect(e).toEqual(pack.equip);
    expect(e).not.toBe(pack.equip);
    const start = { ...DEFAULT_EQUIP };
    const next = equipItem(start, itemOf("cartas", "xadrez")!);
    expect(next).toEqual({ ...DEFAULT_EQUIP, cartas: "xadrez" });
    expect(start).toEqual(DEFAULT_EQUIP);
    expect(JSON.stringify(pack)).toBe(before);
  });
});
