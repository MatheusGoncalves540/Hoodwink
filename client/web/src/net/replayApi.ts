/**
 * Cliente das rotas de replay do backend (plano 05, T15).
 *
 * `GET /me/matches` devolve o histórico do usuário do token (30 dias, no máximo
 * 100 itens). `GET /replays/{matchId}` devolve o `.hwr` binário. Nos dois casos
 * o JWT vai só no Authorization, nunca na URL, e a resposta não pode ser cacheada.
 *
 * Erros sobem como `ReplayApiError` com o status HTTP (0 = rede falhou), para a
 * tela decidir se abre o login (401), avisa que não há acesso (403), que o replay
 * sumiu (404) ou que basta tentar de novo (rede).
 */
import { z } from "zod";
import { ENV } from "../config/envs";
import { copy } from "../game/copy";

export const MATCHES_PAGE_SIZE = 30;
export const MATCHES_MAX_LIMIT = 100;

const matchSchema = z.object({
  match_id: z.string().min(1),
  room_id: z.string().nullish(),
  engine_version: z.number().int(),
  started_at: z.string().min(1),
  ended_at: z.string().min(1),
  winner_id: z.string().nullish(),
  player_ids: z
    .array(z.string())
    .nullish()
    .transform((ids) => ids ?? []),
});

const matchesResponseSchema = z.object({
  data: z.array(matchSchema),
});

export type MyMatch = z.infer<typeof matchSchema>;

export type ReplayDownload = {
  blob: Blob;
  filename: string;
};

/** Erro da API com o status HTTP preservado (0 quando nem houve resposta). */
export class ReplayApiError extends Error {
  readonly status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ReplayApiError";
    this.status = status;
  }
}

export const isAbortError = (error: unknown) =>
  (error as { name?: string } | null)?.name === "AbortError";

const backend = () => ENV.VITE_BACKEND_URL.replace(/\/$/, "");

async function request(
  path: string,
  token: string,
  init: RequestInit = {},
  signal?: AbortSignal,
): Promise<Response> {
  if (!token) throw new ReplayApiError(copy.perfil.needLogin, 401);

  const headers: Record<string, string> = {
    ...(init.headers as Record<string, string> | undefined),
    Authorization: `Bearer ${token}`,
  };
  const requestInit: RequestInit = { ...init, headers, cache: "no-store" };
  if (signal) requestInit.signal = signal;

  let res: Response;
  try {
    res = await fetch(`${backend()}${path}`, requestInit);
  } catch (error) {
    if (isAbortError(error)) throw error;
    throw new ReplayApiError(copy.perfil.network, 0);
  }

  if (res.status === 401) throw new ReplayApiError(copy.perfil.needLogin, 401);
  if (res.status === 403) throw new ReplayApiError(copy.perfil.forbidden, 403);
  if (res.status === 404) throw new ReplayApiError(copy.perfil.notFound, 404);
  if (!res.ok) throw new ReplayApiError(copy.perfil.generic, res.status);
  return res;
}

/** Limita o pedido ao que o backend aceita (1..100); inválido vira o padrão 30. */
export function clampMatchesLimit(limit: number): number {
  if (!Number.isFinite(limit)) return MATCHES_PAGE_SIZE;
  return Math.min(MATCHES_MAX_LIMIT, Math.max(1, Math.trunc(limit)));
}

/** Histórico de partidas do usuário do token, da mais recente para a mais antiga. */
export async function listMyMatches(
  token: string,
  limit = MATCHES_PAGE_SIZE,
  signal?: AbortSignal,
): Promise<MyMatch[]> {
  const bounded = clampMatchesLimit(limit);
  const res = await request(`/me/matches?limit=${bounded}`, token, {}, signal);
  const raw: unknown = await res.json().catch(() => null);
  const parsed = matchesResponseSchema.safeParse(raw);
  if (!parsed.success) throw new ReplayApiError(copy.perfil.badResponse, res.status);
  return parsed.data.data;
}

/**
 * Baixa o `.hwr` de uma partida. O nome vem do Content-Disposition montado no
 * servidor; sem ele, o fallback é `hoodwink-<matchId8>.hwr`.
 */
export async function downloadReplay(
  token: string,
  matchId: string,
  signal?: AbortSignal,
): Promise<ReplayDownload> {
  const id = matchId.trim();
  if (!id) throw new ReplayApiError(copy.perfil.notFound, 404);

  const res = await request(`/replays/${encodeURIComponent(id)}`, token, {}, signal);
  let blob: Blob;
  try {
    blob = await res.blob();
  } catch {
    throw new ReplayApiError(copy.perfil.network, 0);
  }
  return { blob, filename: replayFileName(res.headers.get("Content-Disposition"), id) };
}

/** Nome seguro do arquivo: prioriza filename*=UTF-8'', aceita filename="..." e nunca vira caminho. */
export function replayFileName(disposition: string | null, matchId: string): string {
  const id8 = matchId.replace(/[^0-9a-zA-Z]/g, "").slice(0, 8);
  const fallback = `hoodwink-${id8 || "replay"}.hwr`;
  if (!disposition) return fallback;

  const extended = /filename\*\s*=\s*(?:UTF-8'')?([^;]+)/i.exec(disposition)?.[1]?.trim();
  const plain = /filename\s*=\s*(?:"([^"]*)"|([^;]+))/i.exec(disposition);
  const raw = extended ?? plain?.[1] ?? plain?.[2]?.trim();
  if (!raw) return fallback;

  let decoded = raw.replace(/^"|"$/g, "");
  try {
    decoded = decodeURIComponent(decoded);
  } catch {
    /* mantém o valor como veio */
  }

  const clean = decoded.replace(/[\\/:*?"<>|\u0000-\u001f\u007f]/g, "").trim().slice(0, 120);
  return clean || fallback;
}

/** Dispara o download do Blob no navegador; sem DOM (testes em Node) não faz nada. */
export function saveReplayFile(file: ReplayDownload): void {
  if (typeof document === "undefined" || typeof URL === "undefined") return;
  const url = URL.createObjectURL(file.blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = file.filename;
  link.rel = "noopener";
  document.body.appendChild(link);
  link.click();
  link.remove();
  setTimeout(() => URL.revokeObjectURL(url), 0);
}
