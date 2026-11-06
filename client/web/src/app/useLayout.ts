import { useMedia } from "./prefs";

export type Layout = "mobile" | "desktop";

/** Pelo tamanho da janela, não pelo aparelho (04-arquitetura.md → Mobile e desktop). */
export const useLayout = (): Layout => (useMedia("(min-width: 1024px)") ? "desktop" : "mobile");
