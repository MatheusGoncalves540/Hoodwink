import { setPref, usePref } from "./prefs";

const KEY = "hw-movimento";

/**
 * "Menos movimento" do menu. O prefers-reduced-motion do sistema já é tratado no CSS;
 * isto só cobre a escolha explícita do jogador (vira data-motion="reduced" na raiz).
 */
export function useMotionPref() {
  const reduced = usePref(KEY) === "reduzido";
  return { reduced, setReduced: (on: boolean) => setPref(KEY, on ? "reduzido" : null) };
}
