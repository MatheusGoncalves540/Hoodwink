/**
 * Preferências do aparelho (localStorage) e media queries como hooks.
 * Sem window (teste de fumaça em Node), tudo cai no padrão: sem preferência, sem match.
 */
import { useSyncExternalStore } from "react";

const listeners = new Set<() => void>();
const subscribe = (cb: () => void) => {
  listeners.add(cb);
  return () => void listeners.delete(cb);
};

function read(key: string): string | null {
  try {
    return localStorage.getItem(key);
  } catch {
    return null; // sem localStorage (Node, modo privado bloqueado)
  }
}

export function setPref(key: string, value: string | null) {
  try {
    if (value === null) localStorage.removeItem(key);
    else localStorage.setItem(key, value);
  } catch {
    /* preferência só desta sessão */
  }
  listeners.forEach((l) => l());
}

export const usePref = (key: string) => useSyncExternalStore(subscribe, () => read(key), () => null);

export function useMedia(query: string): boolean {
  return useSyncExternalStore(
    (cb) => {
      const m = window.matchMedia(query);
      m.addEventListener("change", cb);
      return () => m.removeEventListener("change", cb);
    },
    () => window.matchMedia(query).matches,
    () => false,
  );
}
