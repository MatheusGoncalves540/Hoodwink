import { setPref, useMedia, usePref } from "./prefs";

export type Theme = "light" | "dark";
const KEY = "hw-tema";

/** Matinê (light) ou Estreia (dark): segue o sistema até o jogador escolher no menu. */
export function useTheme() {
  const saved = usePref(KEY);
  const override: Theme | null = saved === "light" || saved === "dark" ? saved : null;
  const systemDark = useMedia("(prefers-color-scheme: dark)");
  return {
    theme: override ?? (systemDark ? "dark" : "light") as Theme,
    override,
    setOverride: (t: Theme | null) => setPref(KEY, t),
  };
}
