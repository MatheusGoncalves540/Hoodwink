import { motionState } from "./motion";
import { setPref, useMedia, usePref } from "./prefs";

const KEY = "hw-movimento";
const IDLE_KEY = "hw-espera";

/**
 * As duas chaves de movimento (Menu da sala, Lobby, Como se joga). Só o App escreve na raiz
 * (data-motion / data-idle); componente nenhum lê isto para decidir animação.
 */
export function useMotionPref() {
  const st = motionState({
    motion: usePref(KEY),
    idle: usePref(IDLE_KEY),
    systemReduced: useMedia("(prefers-reduced-motion: reduce)"),
  });
  return {
    ...st,
    setAll: (on: boolean) => setPref(KEY, on ? "ligado" : "reduzido"),
    setIdle: (on: boolean) => setPref(IDLE_KEY, on ? null : "off"),
  };
}
