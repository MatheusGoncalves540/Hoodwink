import { describe, expect, it } from "vitest";
import {
  MIN_STEP_MS,
  PLAYBACK_SPEEDS,
  clampPosition,
  formatClock,
  nextPosition,
  positionAtTime,
  previousPosition,
  stepDelayMs,
  timeAtPosition,
  timelineFromDeltas,
} from "./controls";

describe("timelineFromDeltas", () => {
  it("acumula os deltas e começa no zero", () => {
    const t = timelineFromDeltas([0, 1000, 500]);
    expect(t.steps).toBe(3);
    expect(t.deltas).toEqual([0, 1000, 500]);
    expect(t.times).toEqual([0, 0, 1000, 1500]);
    expect(t.durationMs).toBe(1500);
  });

  it("completa até steps e descarta delta inválido", () => {
    const t = timelineFromDeltas([100, -5, Number.NaN, 200], 6);
    expect(t.deltas).toEqual([100, 0, 0, 200, 0, 0]);
    expect(t.times).toEqual([0, 100, 100, 100, 300, 300, 300]);
    expect(timelineFromDeltas([1, 2], 0)).toEqual({ deltas: [], times: [0], durationMs: 0, steps: 0 });
  });
});

describe("posição ↔ tempo", () => {
  const t = timelineFromDeltas([0, 1000, 500]);

  it("timeAtPosition prende nas bordas", () => {
    expect(timeAtPosition(t, -1)).toBe(0);
    expect(timeAtPosition(t, 0)).toBe(0);
    expect(timeAtPosition(t, 2)).toBe(1000);
    expect(timeAtPosition(t, 3)).toBe(1500);
    expect(timeAtPosition(t, 99)).toBe(1500);
    expect(timeAtPosition(t, Number.NaN)).toBe(0);
  });

  it("positionAtTime devolve o último passo do instante", () => {
    expect(positionAtTime(t, -10)).toBe(1);
    expect(positionAtTime(t, 0)).toBe(1);
    expect(positionAtTime(t, 999)).toBe(1);
    expect(positionAtTime(t, 1000)).toBe(2);
    expect(positionAtTime(t, 1499)).toBe(2);
    expect(positionAtTime(t, 1500)).toBe(3);
    expect(positionAtTime(t, 99_999)).toBe(3);
    expect(positionAtTime(timelineFromDeltas([]), 1000)).toBe(0);
  });
});

describe("passos e fim", () => {
  it("próximo e anterior prendem em 0..steps", () => {
    expect(clampPosition(-4, 5)).toBe(0);
    expect(clampPosition(2.9, 5)).toBe(2);
    expect(clampPosition(9, 5)).toBe(5);
    expect(clampPosition(Number.NaN, 5)).toBe(0);
    expect(nextPosition(4, 5)).toBe(5);
    expect(nextPosition(5, 5)).toBe(5);
    expect(previousPosition(0, 5)).toBe(0);
    expect(previousPosition(1, 5)).toBe(0);
    expect(previousPosition(5, 5)).toBe(4);
  });

  it("a espera segue o delta e a velocidade, com piso", () => {
    const t = timelineFromDeltas([0, 2000, 400]);
    expect(stepDelayMs(t, 0, 1)).toBe(MIN_STEP_MS);
    expect(stepDelayMs(t, 1, 1)).toBe(2000);
    expect(stepDelayMs(t, 1, 2)).toBe(1000);
    expect(stepDelayMs(t, 2, 4)).toBe(100);
    expect(stepDelayMs(t, 3, 1)).toBe(0);
    expect(stepDelayMs(t, 99, 1)).toBe(0);
    expect(stepDelayMs(t, 1, 3)).toBe(2000);
    expect(PLAYBACK_SPEEDS).toEqual([0.5, 1, 2, 4]);
  });
});

describe("formatClock", () => {
  it("formata mm:ss e aguenta valores inválidos", () => {
    expect(formatClock(0)).toBe("00:00");
    expect(formatClock(-5000)).toBe("00:00");
    expect(formatClock(Number.NaN)).toBe("00:00");
    expect(formatClock(999)).toBe("00:00");
    expect(formatClock(59_999)).toBe("00:59");
    expect(formatClock(60_000)).toBe("01:00");
    expect(formatClock(3_599_000)).toBe("59:59");
    expect(formatClock(3_600_000)).toBe("60:00");
    expect(formatClock(6 * 60 * 60 * 1000)).toBe("360:00");
  });
});
