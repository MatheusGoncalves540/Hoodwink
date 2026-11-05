import { describe, expect, it } from "vitest";
import { motionState } from "./motion";

describe("motionState", () => {
  it("todas as combinações de Animações × espera × sistema", () => {
    const rows = [null, "reduzido", "ligado"].flatMap((motion) =>
      [null, "off"].flatMap((idle) => [false, true].map((systemReduced) => ({ motion, idle, systemReduced }))),
    );
    expect(rows.map((r) => [r.motion, r.idle, r.systemReduced, motionState(r)])).toEqual([
      [null, null, false, { all: true, idle: true }],
      [null, null, true, { all: false, idle: false }],
      [null, "off", false, { all: true, idle: false }],
      [null, "off", true, { all: false, idle: false }],
      ["reduzido", null, false, { all: false, idle: false }],
      ["reduzido", null, true, { all: false, idle: false }],
      ["reduzido", "off", false, { all: false, idle: false }],
      ["reduzido", "off", true, { all: false, idle: false }],
      ["ligado", null, false, { all: true, idle: true }],
      ["ligado", null, true, { all: true, idle: true }],
      ["ligado", "off", false, { all: true, idle: false }],
      ["ligado", "off", true, { all: true, idle: false }],
    ]);
  });

  it("religar Animações devolve a espera ao valor guardado", () => {
    expect(motionState({ motion: "reduzido", idle: null, systemReduced: false }).idle).toBe(false);
    expect(motionState({ motion: "ligado", idle: null, systemReduced: false }).idle).toBe(true);
    expect(motionState({ motion: "reduzido", idle: "off", systemReduced: false }).idle).toBe(false);
    expect(motionState({ motion: "ligado", idle: "off", systemReduced: false }).idle).toBe(false);
  });
});
