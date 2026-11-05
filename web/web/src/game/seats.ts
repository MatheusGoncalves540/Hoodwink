/**
 * Posição dos camarotes na mesa (04-arquitetura.md → "Posições dos camarotes").
 * A ordem é a de jogo a partir de quem vem depois de você.
 *
 * mobileSeats(k, S): tabela de Mesa6.dc.html para palco de 390 px, escalada por S/390.
 *   até 5 adversários: 96×104; de 6 a 9: 70×88.
 * desktopSeats(k, S): parábola de 04-arquitetura.md, camarotes 132×140 (até 5) ou 108×120.
 *   passo = min(210, (S − 30 − W)/(k − 1)), x0 = S/2 − (passo·(k−1) + W)/2, u = i/(k−1),
 *   y = 92 + (k > 5 ? 150 : 128)·(2u − 1)².
 *   ponytail: em palcos estreitos (< ~1020 px com 9 adversários) a fórmula encostaria os camarotes;
 *   aí o tamanho encolhe por igual (`scale`) até sobrar 2 px entre eles.
 */
export type SeatBox = { x: number; y: number; w: number; h: number };
export type SeatLayout = { boxes: SeatBox[]; big: boolean; scale: number };

const MOBILE_SEATS: Record<number, [number, number][]> = {
  1: [[147, 84]],
  2: [[64, 96], [230, 96]],
  3: [[30, 120], [147, 70], [264, 120]],
  4: [[8, 168], [92, 76], [202, 76], [286, 168]],
  5: [[8, 182], [52, 76], [147, 56], [242, 76], [286, 182]],
  6: [[8, 196], [8, 82], [107, 60], [213, 60], [312, 82], [312, 196]],
  7: [[8, 196], [8, 84], [84, 64], [160, 56], [236, 64], [312, 84], [312, 196]],
  8: [[8, 206], [84, 182], [8, 84], [107, 60], [213, 60], [312, 84], [234, 182], [312, 206]],
  9: [[8, 206], [84, 182], [8, 84], [84, 64], [160, 56], [236, 64], [312, 84], [234, 182], [312, 206]],
};

export const MAX_OPPONENTS = 9;

export function mobileSeats(k: number, stageWidth = 390): SeatLayout {
  const big = k <= 5;
  const [w, h] = big ? [96, 104] : [70, 88];
  const f = stageWidth / 390;
  const boxes = (MOBILE_SEATS[k] ?? []).map(([x, y]) => ({ x: x * f, y, w, h }));
  return { boxes, big, scale: 1 };
}

const GAP = 2;

export function desktopSeats(k: number, stageWidth: number): SeatLayout {
  if (k < 1) return { boxes: [], big: true, scale: 1 };
  const big = k <= 5;
  const [W0, H0] = big ? [132, 140] : [108, 120];
  const avail = stageWidth - 30;
  const scale = Math.min(1, avail / ((k - 1) * (W0 + GAP) + W0));
  const w = W0 * scale;
  const h = H0 * scale;
  if (k === 1) return { boxes: [{ x: stageWidth / 2 - w / 2, y: 92, w, h }], big, scale };
  const step = Math.min(210, (avail - w) / (k - 1));
  const x0 = stageWidth / 2 - (step * (k - 1) + w) / 2;
  const boxes = Array.from({ length: k }, (_, i) => {
    const u = i / (k - 1);
    return { x: x0 + step * i, y: 92 + (big ? 128 : 150) * (2 * u - 1) ** 2, w, h };
  });
  return { boxes, big, scale };
}

/** clip-path do cone de luz: do ponto (ox, oy) até a base do camarote, com 6 px de folga de cada lado. */
export function conePolygon(b: SeatBox, ox: number, oy = 0): string {
  return `polygon(${ox}px ${oy}px, ${b.x - 6}px ${b.y + b.h}px, ${b.x + b.w + 6}px ${b.y + b.h}px)`;
}
