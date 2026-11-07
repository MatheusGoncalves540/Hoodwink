import { describe, expect, it } from "vitest";
import { spotShape, desktopSeats, mobileSeats, type SeatBox } from "./seats";

const overlap = (a: SeatBox, b: SeatBox) =>
  a.x < b.x + b.w && b.x < a.x + a.w && a.y < b.y + b.h && b.y < a.y + a.h;

describe("mobileSeats", () => {
  it("devolve k posições para k de 1 a 10, com o tamanho certo", () => {
    for (let k = 1; k <= 10; k++) {
      const l = mobileSeats(k);
      expect(l.boxes).toHaveLength(k);
      expect(l.boxes[0]).toMatchObject(k <= 5 ? { w: 96, h: 104 } : { w: 70, h: 88 });
    }
  });

  it("reproduz a tabela de Mesa6 e escala pela largura", () => {
    expect(mobileSeats(3).boxes).toEqual([
      { x: 30, y: 120, w: 96, h: 104 },
      { x: 147, y: 70, w: 96, h: 104 },
      { x: 264, y: 120, w: 96, h: 104 },
    ]);
    expect(mobileSeats(1, 780).boxes[0].x).toBe(294);
  });

  it("zero adversários: nada", () => {
    expect(mobileSeats(0).boxes).toEqual([]);
  });

  it("10 jogadores do replay cabem sem sobrepor", () => {
    const { boxes } = mobileSeats(10);
    expect(boxes).toHaveLength(10);
    boxes.forEach((b, i) => boxes.slice(i + 1).forEach((o) => expect(overlap(b, o)).toBe(false)));
  });
});

describe("desktopSeats", () => {
  it("nenhum camarote se sobrepõe e todos ficam no palco (k 1..10, largura 900..1400)", () => {
    for (let k = 1; k <= 10; k++) {
      for (let S = 900; S <= 1400; S += 50) {
        const { boxes } = desktopSeats(k, S);
        expect(boxes).toHaveLength(k);
        boxes.forEach((b, i) => {
          expect(b.x).toBeGreaterThanOrEqual(0);
          expect(b.x + b.w).toBeLessThanOrEqual(S);
          expect(b.y).toBeGreaterThanOrEqual(0);
          boxes.slice(i + 1).forEach((o) => expect(overlap(b, o)).toBe(false));
        });
      }
    }
  });

  it("no palco de referência (1020) com 9 adversários, mantém 108×120 sem encolher", () => {
    expect(desktopSeats(9, 1020)).toMatchObject({ scale: 1, big: false });
    expect(desktopSeats(9, 1020).boxes[4]).toMatchObject({ y: 92, w: 108, h: 120 });
  });

  it("ferradura simétrica: as pontas ficam mais baixas que o meio", () => {
    const { boxes } = desktopSeats(5, 1020);
    expect(boxes[0].y).toBe(boxes[4].y);
    expect(boxes[0].y).toBeGreaterThan(boxes[2].y);
  });
});

it("holofote: cone sai da origem e abraça o camarote com 6 px de folga; a poça fica centrada na base", () => {
  expect(spotShape({ x: 107, y: 60, w: 96, h: 104 }, 195)).toEqual({
    cone: "195,0 101,164 209,164",
    pool: { cx: 155, cy: 164, rx: 54, ry: 11 },
  });
  // só a ponta de cima mexe entre os quadros; base e poça ficam presas ao camarote
  const a = spotShape({ x: 10, y: 20, w: 80, h: 90 }, 200, 0, 3);
  expect(a.cone).toBe("203,0 4,110 96,110");
  expect(a.pool).toEqual(spotShape({ x: 10, y: 20, w: 80, h: 90 }, 200).pool);
});
