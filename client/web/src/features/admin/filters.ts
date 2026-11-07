/**
 * Regras puras dos filtros da tela de requests (plano 04, T8).
 * O backend valida de novo: aqui só evitamos mandar lixo e explicamos o erro antes.
 */

export const DAY_MS = 86_400_000;
export const DEFAULT_LIMIT = 50;
export const MAX_LIMIT = 200;
export const MAX_PAGE = 1000;
export const MAX_RANGE_DAYS = 31;
export const MAX_ARCHIVED_RANGE_DAYS = 183;
export const PAGE_SIZE_OPTIONS = [50, 100, 200] as const;

export type Period = "24h" | "7d" | "30d" | "custom";

export const PERIODS: { value: Period; label: string }[] = [
  { value: "24h", label: "24 horas" },
  { value: "7d", label: "7 dias" },
  { value: "30d", label: "30 dias" },
  { value: "custom", label: "Personalizado" },
];

export type DateRange = { from: string; to: string };

/** Últimas 24h, 7 dias ou 30 dias, em ISO (UTC). `now` entra como parâmetro para o teste. */
export function presetRange(period: Exclude<Period, "custom">, now = new Date()): DateRange {
  const days = period === "24h" ? 1 : period === "7d" ? 7 : 30;
  const to = now.getTime();
  return { from: new Date(to - days * DAY_MS).toISOString(), to: new Date(to).toISOString() };
}

export type RangeResult = { range?: DateRange; error?: string };

/**
 * Valida o período personalizado (valores de `datetime-local` ou ISO) e devolve
 * ISO para a API. O limite é 31 dias, ou 183 quando a consulta é nos arquivados.
 */
export function customRange(fromInput: string, toInput: string, archived = false): RangeResult {
  const from = fromInput.trim();
  const to = toInput.trim();
  if (!from || !to) return { error: "Informe o início e o fim do período." };

  const fromMs = Date.parse(from);
  const toMs = Date.parse(to);
  if (!Number.isFinite(fromMs) || !Number.isFinite(toMs)) return { error: "Data inválida." };
  if (fromMs > toMs) return { error: "O início precisa vir antes do fim." };

  const maxDays = archived ? MAX_ARCHIVED_RANGE_DAYS : MAX_RANGE_DAYS;
  if (toMs - fromMs > maxDays * DAY_MS) return { error: `O período máximo é de ${maxDays} dias.` };

  return { range: { from: new Date(fromMs).toISOString(), to: new Date(toMs).toISOString() } };
}

export type StatusResult = { value?: string; error?: string };

const STATUS_RANGE = /^[1-5]xx$/i;

/** Aceita vazio, `4xx` (normaliza para minúsculas) ou um inteiro de 100 a 599. */
export function parseStatusInput(input: string): StatusResult {
  const value = input.trim().toLowerCase();
  if (!value) return { value: "" };

  if (STATUS_RANGE.test(value)) return { value };

  const code = Number(value);
  if (Number.isInteger(code) && code >= 100 && code <= 599) return { value: String(code) };

  return { error: "Use 4xx ou um número de 100 a 599." };
}

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

/** Formato canônico de UUID, como o id aparece no banco. */
export const isUuid = (value: string) => UUID.test(value.trim());

/** Mantém a página no intervalo que o backend aceita (1 a 1000). */
export function clampPage(value: number): number {
  if (!Number.isFinite(value)) return 1;
  return Math.min(MAX_PAGE, Math.max(1, Math.trunc(value)));
}

/** Mantém o limite entre 1 e 200; valor inválido cai no padrão. */
export function clampLimit(value: number): number {
  if (!Number.isFinite(value)) return DEFAULT_LIMIT;
  return Math.min(MAX_LIMIT, Math.max(1, Math.trunc(value)));
}

/** Total de páginas, nunca menor que 1 e nunca acima do teto do backend. */
export function pageCount(total: number, limit: number): number {
  const size = clampLimit(limit);
  return Math.min(MAX_PAGE, Math.max(1, Math.ceil(Math.max(0, total) / size)));
}

const pad = (n: number) => String(n).padStart(2, "0");

/** ISO → valor de `datetime-local` no fuso do navegador. */
export function toLocalInput(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}
