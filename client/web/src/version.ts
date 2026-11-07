/**
 * Versão do jogo (semver), independente da versão do motor de replay.
 * Fonte da verdade: arquivo VERSION na raiz do repositório. O Vite injeta o
 * valor em VITE_GAME_VERSION no dev e no build; sem ele (ex.: teste unitário),
 * fica "dev".
 */
const injected = import.meta.env.VITE_GAME_VERSION;

export const GAME_VERSION: string =
  typeof injected === "string" && injected.trim() ? injected.trim() : "dev";
