/**
 * Texto vindo do `.hwr` é hostil: o arquivo pode ter sido editado no disco.
 * Nada aqui vira HTML (o React já escapa o texto); a função só tira o que não
 * deve aparecer na tela: caracteres de controle, marcas de bidi e excesso de
 * tamanho. O corte é por code point para não partir emoji no meio.
 */
export const MAX_REPLAY_TEXT = 32;

/** Controle C0/C1, separadores de linha Unicode e o BOM/zero-width no-break. */
const CONTROL = /[\u0000-\u001f\u007f-\u009f\u2028\u2029\ufeff]/g;
/** Marcas de direção e invisíveis usadas para trocar a ordem do texto na tela. */
const BIDI = /[\u061c\u200b-\u200f\u202a-\u202e\u2066-\u2069]/g;

/**
 * Corta em `max` caracteres e remove controle/bidi. Entrada que não é string
 * vira string vazia; a função nunca lança.
 */
export function sanitizeText(raw: unknown, max = MAX_REPLAY_TEXT): string {
  if (typeof raw !== "string") return "";
  const limit = Number.isFinite(max) && max > 0 ? Math.floor(max) : 0;
  const clean = raw.normalize("NFC").replace(CONTROL, "").replace(BIDI, "").trim();
  return Array.from(clean).slice(0, limit).join("");
}
