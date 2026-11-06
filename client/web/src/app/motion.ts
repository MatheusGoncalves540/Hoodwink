/**
 * Duas chaves de movimento (plans/04-handoff-refino-v2.md, fase 3.1). Função pura: sem DOM.
 * - motion: "reduzido" = Animações desligada; "ligado" = ligada de propósito (vence o sistema); null = segue o sistema.
 * - idle: "off" = Animações de espera desligada; null = ligada.
 * Animações desligada desliga a espera junto, sem apagar a escolha guardada da espera.
 */
export type MotionInput = { motion: string | null; idle: string | null; systemReduced: boolean };
export type MotionState = { all: boolean; idle: boolean };

export function motionState({ motion, idle, systemReduced }: MotionInput): MotionState {
  const all = motion === "reduzido" ? false : motion === "ligado" ? true : !systemReduced;
  return { all, idle: all && idle !== "off" };
}
