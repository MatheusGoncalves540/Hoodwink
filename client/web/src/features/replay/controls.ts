/**
 * Matemática da barra do replay (plano 05, T14). Funções puras: nada de timer,
 * DOM ou sessão. A posição é o número de entradas já aplicadas (0 = antes do
 * início, `steps` = fim); o tempo vem da soma dos deltas das entradas.
 */
export const PLAYBACK_SPEEDS = [0.5, 1, 2, 4] as const;
export type PlaybackSpeed = (typeof PLAYBACK_SPEEDS)[number];

/** Piso do intervalo de reprodução: sem isso, entradas de 1 ms passam num piscar. */
export const MIN_STEP_MS = 120;

export type ReplayTimeline = {
  /** Delta de cada entrada, em ms, já saneado (inteiro ≥ 0). */
  deltas: readonly number[];
  /** Tempo acumulado depois de `i` entradas; sempre começa em 0 e tem `steps + 1` itens. */
  times: readonly number[];
  durationMs: number;
  steps: number;
};

/** Monta a linha do tempo; deltas faltando/ inválidos viram 0 e a lista é completada até `steps`. */
export function timelineFromDeltas(deltas: readonly number[], steps = deltas.length): ReplayTimeline {
  const total = Number.isFinite(steps) && steps > 0 ? Math.floor(steps) : 0;
  const clean: number[] = [];
  for (let i = 0; i < total; i++) {
    const delta = deltas[i];
    clean.push(Number.isFinite(delta) && delta > 0 ? Math.floor(delta) : 0);
  }

  const times = [0];
  for (const delta of clean) times.push(times[times.length - 1] + delta);
  return { deltas: clean, times, durationMs: times[times.length - 1], steps: total };
}

/** Prende a posição em 0..steps; fração e NaN não escapam. */
export function clampPosition(position: number, steps: number): number {
  const max = Number.isFinite(steps) && steps > 0 ? Math.floor(steps) : 0;
  if (!Number.isFinite(position)) return 0;
  return Math.min(Math.max(Math.trunc(position), 0), max);
}

export function nextPosition(position: number, steps: number): number {
  return clampPosition(clampPosition(position, steps) + 1, steps);
}

export function previousPosition(position: number, steps: number): number {
  return clampPosition(clampPosition(position, steps) - 1, steps);
}

/** Tempo da posição, em ms. */
export function timeAtPosition(timeline: ReplayTimeline, position: number): number {
  return timeline.times[clampPosition(position, timeline.steps)] ?? 0;
}

/**
 * Posição cujo tempo acumulado é o maior ≤ `ms` (scrub). Com deltas iguais a
 * zero, devolve o último passo daquele instante.
 */
export function positionAtTime(timeline: ReplayTimeline, ms: number): number {
  if (timeline.steps === 0) return 0;
  const target = Number.isFinite(ms) ? Math.min(Math.max(ms, 0), timeline.durationMs) : 0;
  let lo = 0;
  let hi = timeline.steps;
  while (lo < hi) {
    const mid = Math.ceil((lo + hi) / 2);
    if ((timeline.times[mid] ?? 0) <= target) lo = mid;
    else hi = mid - 1;
  }
  return lo;
}

/** Relógio curto mm:ss (minutos podem passar de 59 numa partida longa). */
export function formatClock(ms: number): string {
  const seconds = Number.isFinite(ms) && ms > 0 ? Math.floor(ms / 1000) : 0;
  const minutes = Math.floor(seconds / 60);
  return `${String(minutes).padStart(2, "0")}:${String(seconds % 60).padStart(2, "0")}`;
}

/** Espera até a próxima entrada, já dividida pela velocidade. */
export function stepDelayMs(timeline: ReplayTimeline, position: number, speed: number): number {
  const at = clampPosition(position, timeline.steps);
  if (at >= timeline.steps) return 0;
  const delta = timeline.deltas[at] ?? 0;
  const rate = PLAYBACK_SPEEDS.includes(speed as PlaybackSpeed) ? speed : 1;
  return Math.max(MIN_STEP_MS, delta) / rate;
}
