/**
 * Cliente das rotas /bugReports (plano 06, T6).
 * O backend limita o corpo (16 KB) e mascara o contexto; aqui só enviamos.
 * Erros sobem como BugApiError com o status HTTP (0 = rede falhou) e o resultado
 * do classifyBugResult. 401 abre o login; 429 e rede não perdem o texto: quem
 * chama decide o que manter.
 */
import { ENV } from "../config/envs";
import type { BugCategory } from "../game/copy";
import { useAuthStore } from "../store/authStore";
import type { BugContext } from "./bugContext";
import { classifyBugResult, type BugResult } from "./bugResult";

export class BugApiError extends Error {
  readonly status: number;
  readonly result: BugResult;

  constructor(status: number) {
    const result = classifyBugResult(status);
    super(result);
    this.name = "BugApiError";
    this.status = status;
    this.result = result;
  }
}

const backend = () => ENV.VITE_BACKEND_URL.replace(/\/$/, "");

async function request(method: "POST" | "PATCH", path: string, token: string, body: unknown): Promise<unknown> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (token) headers.Authorization = `Bearer ${token}`;

  let res: Response;
  try {
    res = await fetch(`${backend()}${path}`, { method, headers, body: JSON.stringify(body) });
  } catch {
    throw new BugApiError(0);
  }

  if (res.status === 401) {
    useAuthStore.getState().openLogin();
    throw new BugApiError(401);
  }

  const raw: unknown = await res.json().catch(() => null);
  if (!res.ok) throw new BugApiError(res.status);
  return raw;
}

export type FlagResult = { reportId: string; turns: number[] };

function parseFlag(raw: unknown): FlagResult {
  if (!raw || typeof raw !== "object") throw new BugApiError(0);
  const { id, turns } = raw as { id?: unknown; turns?: unknown };
  const reportId = typeof id === "number" ? String(id) : typeof id === "string" ? id : "";
  const cleanTurns = Array.isArray(turns) ? turns.filter((t): t is number => typeof t === "number") : [];
  if (!reportId) throw new BugApiError(0);
  return { reportId, turns: cleanTurns };
}

/** Marca um turno na partida (login obrigatório). Devolve o id do relato e os turnos. */
export async function flagBug(token: string, input: { matchId: string; turn: number }): Promise<FlagResult> {
  const raw = await request("POST", "/bugReports/flag", token, { matchId: input.matchId, turn: input.turn });
  return parseFlag(raw);
}

export type LobbyBugInput = {
  category: BugCategory;
  description: string;
  context: BugContext | null;
};

/** Relato novo do Lobby (auth opcional: sem Bearer válido, o backend grava anônimo). */
export async function sendBugReport(token: string, payload: LobbyBugInput): Promise<void> {
  await request("POST", "/bugReports", token, payload);
}

export type PatchBugInput = {
  category: BugCategory | null;
  description: string | null;
  context: BugContext | null;
};

/** Detalha um relato de partida (só o dono). Texto nulo deixa o campo como está. */
export async function patchBugReport(token: string, id: string, payload: PatchBugInput): Promise<void> {
  await request("PATCH", `/bugReports/${encodeURIComponent(id)}`, token, payload);
}
