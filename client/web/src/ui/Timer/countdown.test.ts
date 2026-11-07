import { describe, expect, it } from "vitest";
import { countdown } from "./countdown";

describe("countdown", () => {
  it("número arredonda para cima e barra usa a fração exata do mesmo now", () => {
    const now = 1_000_000;
    expect(countdown(now + 10_000, 10, now + 3_000, 10_000)).toEqual({ left: 7, progress: 0.7 });
    expect(countdown(now + 10_000, 10, now, 10_000)).toEqual({ left: 10, progress: 1 });
    expect(countdown(now + 10_000, 10, now + 9_500, 10_000)).toEqual({ left: 1, progress: 0.05 });
  });

  it("zera no prazo, sem o segundo extra depois da barra", () => {
    const now = 1_000_000;
    expect(countdown(now, 10, now, 10_000)).toEqual({ left: 0, progress: 0 });
    expect(countdown(now - 1, 10, now, 10_000)).toEqual({ left: 0, progress: 0 });
  });

  it("sem janela (antes do primeiro evento) mostra o total cheio", () => {
    expect(countdown(undefined, 6, 1_000, 6_000)).toEqual({ left: 6, progress: 1 });
  });

  it("janela real maior que o total não congela a barra no cheio", () => {
    const now = 1_000_000;
    expect(countdown(now + 15_000, 10, now + 10_000, 15_000)).toEqual({ left: 5, progress: 1 / 3 });
  });
});
