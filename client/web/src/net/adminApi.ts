/**
 * Cliente da API de requests do super-admin (plano 04, T8).
 * A resposta é plana: `{ items, total, page, limit }`, com os campos em snake_case.
 * O token vai no Authorization; 401 e 403 sobem como AdminApiError com o status.
 */
import { z } from "zod";
import { ENV } from "../config/envs";
import { BUG_CATEGORIES } from "../game/copy";

const requestLogSchema = z.object({
  id: z.number().int(),
  at: z.string().min(1),
  service: z.enum(["backend", "gameServer"]),
  method: z.string().nullable(),
  path: z.string().nullable(),
  route: z.string().nullable(),
  status: z.number().int().nullable(),
  duration_ms: z.number().int().nullable(),
  ip: z.string().nullable(),
  user_agent: z.string().nullable(),
  user_id: z.string().nullable(),
  username: z.string().nullable(),
  request_id: z.string().nullable(),
  body: z.string().nullable(),
  error: z.string().nullable(),
});

const requestLogsSchema = z.object({
  items: z.array(requestLogSchema),
  total: z.number().int().nonnegative(),
  page: z.number().int().positive(),
  limit: z.number().int().positive(),
});

export type AdminRequestLog = z.infer<typeof requestLogSchema>;
export type AdminRequestLogsResponse = z.infer<typeof requestLogsSchema>;

export type AdminRequestLogFilters = {
  from?: string;
  to?: string;
  route?: string;
  status?: string;
  userId?: string;
  username?: string;
  service?: "" | "backend" | "gameServer";
  ip?: string;
  archived?: boolean;
  page?: number;
  limit?: number;
};

/** Erro da API com o status HTTP preservado (0 quando nem houve resposta). */
export class AdminApiError extends Error {
  readonly status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "AdminApiError";
    this.status = status;
  }
}

const put = (params: URLSearchParams, key: string, value: string | number | undefined) => {
  if (value === undefined) return;
  const text = String(value).trim();
  if (text !== "") params.set(key, text);
};

/** Serializa só o que tem valor; `archived` só aparece quando é `true`. */
export function adminRequestLogsQuery(filters: AdminRequestLogFilters): URLSearchParams {
  const params = new URLSearchParams();
  put(params, "from", filters.from);
  put(params, "to", filters.to);
  put(params, "route", filters.route);
  put(params, "status", filters.status);
  put(params, "userId", filters.userId);
  put(params, "username", filters.username);
  put(params, "service", filters.service);
  put(params, "ip", filters.ip);
  if (filters.archived) params.set("archived", "true");
  put(params, "page", filters.page);
  put(params, "limit", filters.limit);
  return params;
}

const messageOf = (raw: unknown, status: number) => {
  if (raw && typeof raw === "object" && "message" in raw && typeof raw.message === "string") return raw.message;
  return `O servidor respondeu ${status}.`;
};

/**
 * Busca uma página de requests. Erros:
 * - `AdminApiError.status === 401`: sessão inválida/expirada;
 * - `AdminApiError.status === 403`: conta sem `is_super_admin`;
 * - `AdminApiError.status === 0`: rede falhou, sem resposta.
 */
export async function listRequestLogs(
  token: string,
  filters: AdminRequestLogFilters = {},
  signal?: AbortSignal,
): Promise<AdminRequestLogsResponse> {
  const query = adminRequestLogsQuery(filters).toString();
  const base = ENV.VITE_BACKEND_URL.replace(/\/$/, "");
  const init: RequestInit = { headers: { Authorization: `Bearer ${token}` } };
  if (signal) init.signal = signal;

  let res: Response;
  try {
    res = await fetch(`${base}/admin/requestLogs${query ? `?${query}` : ""}`, init);
  } catch (error) {
    if ((error as { name?: string } | null)?.name === "AbortError") throw error;
    throw new AdminApiError("Não foi possível falar com o servidor.", 0);
  }

  const raw: unknown = await res.json().catch(() => null);

  if (res.status === 401) throw new AdminApiError("Sessão expirada. Entre de novo.", 401);
  if (res.status === 403) throw new AdminApiError("Sem acesso", 403);
  if (!res.ok) throw new AdminApiError(messageOf(raw, res.status), res.status);

  const parsed = requestLogsSchema.safeParse(raw);
  if (!parsed.success) throw new AdminApiError("Resposta inesperada do servidor.", res.status);
  return parsed.data;
}

// --- Relatos de bug (plano 06, T7) ---

export const BUG_STATUSES = ["novo", "visto", "resolvido"] as const;
export type BugReportStatus = (typeof BUG_STATUSES)[number];

export const BUG_SOURCES = ["partida", "lobby"] as const;
export type BugReportSource = (typeof BUG_SOURCES)[number];

const bugReportSchema = z.object({
  id: z.number().int().positive(),
  created_at: z.string().min(1),
  updated_at: z.string().min(1),
  user_id: z.string().nullable(),
  username: z.string().nullable(),
  source: z.enum(BUG_SOURCES),
  match_id: z.string().nullable(),
  turns: z.array(z.number().int()),
  category: z.enum(BUG_CATEGORIES).nullable(),
  description: z.string().nullable(),
  context: z.unknown(),
  ip: z.string().nullable(),
  user_agent: z.string().nullable(),
  status: z.enum(BUG_STATUSES),
});

const bugReportsSchema = z.object({
  items: z.array(bugReportSchema),
  total: z.number().int().nonnegative(),
  novos: z.number().int().nonnegative(),
  page: z.number().int().positive(),
  limit: z.number().int().positive(),
});

export type AdminBugReport = z.infer<typeof bugReportSchema>;
export type AdminBugReportsResponse = z.infer<typeof bugReportsSchema>;

export type AdminBugReportFilters = {
  status?: "" | BugReportStatus;
  source?: "" | BugReportSource;
  matchId?: string;
  page?: number;
  limit?: number;
};

/** Serializa só o que tem valor, no mesmo formato dos filtros de requests. */
export function adminBugReportsQuery(filters: AdminBugReportFilters): URLSearchParams {
  const params = new URLSearchParams();
  put(params, "status", filters.status);
  put(params, "source", filters.source);
  put(params, "matchId", filters.matchId);
  put(params, "page", filters.page);
  put(params, "limit", filters.limit);
  return params;
}

/** GET/PATCH do admin com o erro padrão: 401 sessão, 403 "Sem acesso", 0 rede. */
async function readAdmin<T>(
  path: string,
  token: string,
  schema: z.ZodType<T>,
  init: RequestInit = {},
  signal?: AbortSignal,
): Promise<T> {
  const headers: Record<string, string> = { Authorization: `Bearer ${token}` };
  if (init.body !== undefined) headers["Content-Type"] = "application/json";

  const request: RequestInit = { ...init, headers };
  if (signal) request.signal = signal;

  let res: Response;
  try {
    res = await fetch(`${ENV.VITE_BACKEND_URL.replace(/\/$/, "")}${path}`, request);
  } catch (error) {
    if ((error as { name?: string } | null)?.name === "AbortError") throw error;
    throw new AdminApiError("Não foi possível falar com o servidor.", 0);
  }

  const raw: unknown = await res.json().catch(() => null);

  if (res.status === 401) throw new AdminApiError("Sessão expirada. Entre de novo.", 401);
  if (res.status === 403) throw new AdminApiError("Sem acesso", 403);
  if (!res.ok) throw new AdminApiError(messageOf(raw, res.status), res.status);

  const parsed = schema.safeParse(raw);
  if (!parsed.success) throw new AdminApiError("Resposta inesperada do servidor.", res.status);
  return parsed.data;
}

/** Busca uma página de relatos, do mais novo para o mais antigo. */
export async function listBugReports(
  token: string,
  filters: AdminBugReportFilters = {},
  signal?: AbortSignal,
): Promise<AdminBugReportsResponse> {
  const query = adminBugReportsQuery(filters).toString();
  return readAdmin(`/admin/bugReports${query ? `?${query}` : ""}`, token, bugReportsSchema, {}, signal);
}

/** Troca o status do relato e devolve o relato atualizado. */
export async function updateBugReportStatus(
  token: string,
  id: number,
  status: BugReportStatus,
  signal?: AbortSignal,
): Promise<AdminBugReport> {
  return readAdmin(
    `/admin/bugReports/${id}`,
    token,
    bugReportSchema,
    { method: "PATCH", body: JSON.stringify({ status }) },
    signal,
  );
}
