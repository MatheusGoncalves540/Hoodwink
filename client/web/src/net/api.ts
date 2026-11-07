/**
 * Endpoints HTTP do gameServer usados antes do WebSocket (05-protocolo.md → "Conexão"):
 * POST /newRoom e POST /getTicket/{RoomId}, ambos com o JWT do backend no Authorization.
 * O host vem de config/envs.ts; a resposta é validada com Zod como a do WebSocket.
 */
import { z } from "zod";
import { ENV } from "../config/envs";

const serverUrl = () => ENV.VITE_GAMESERVER_URL.replace(/\/$/, "");

const createRoomSchema = z.object({
  RoomId: z.string().min(1),
  message: z.string().optional(),
});

const ticketSchema = z.object({
  message: z.string().optional(),
  data: z.object({ ticket: z.string().min(1) }),
});

const errorMessage = (raw: unknown, status: number) => {
  if (raw && typeof raw === "object" && "message" in raw && typeof raw.message === "string") return raw.message;
  return `O servidor respondeu ${status}.`;
};

async function postJson(path: string, token: string, body?: unknown): Promise<unknown> {
  const res = await fetch(`${serverUrl()}${path}`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const raw: unknown = await res.json().catch(() => null);
  if (!res.ok) throw new Error(errorMessage(raw, res.status));
  return raw;
}

export type NewRoomInput = {
  roomName: string;
  /** 2..10, como o validate do servidor */
  maxPlayers: number;
  password?: string;
};

/** Cria a sala e devolve o id que vai em `?sala=`. */
export async function newRoom(token: string, input: NewRoomInput): Promise<string> {
  const raw = await postJson("/newRoom", token, {
    roomName: input.roomName,
    password: input.password ?? "",
    maxPlayers: input.maxPlayers,
  });
  return createRoomSchema.parse(raw).RoomId;
}

/** Pede o ticket JWT da sala; é ele que abre o WebSocket (`?Ticket=`). */
export async function getTicket(token: string, roomId: string): Promise<string> {
  const raw = await postJson(`/getTicket/${encodeURIComponent(roomId)}`, token);
  return ticketSchema.parse(raw).data.ticket;
}

/** Resultado do refresh: ok (token novo) | unauthorized (401) | failed (rede/outro). */
export type RefreshResult =
  | { status: "ok"; token: string }
  | { status: "unauthorized" }
  | { status: "failed" };

let refreshInFlight: Promise<RefreshResult> | null = null;

/**
 * Troca o token de login por um novo no backend. Uma Promise por vez: chamadas
 * simultâneas reaproveitam a mesma requisição. Erro de rede mantém o token
 * (failed); 401 é devolvido para o chamador limpar a sessão.
 */
export function refreshToken(token: string): Promise<RefreshResult> {
  if (!refreshInFlight) {
    refreshInFlight = doRefresh(token).finally(() => {
      refreshInFlight = null;
    });
  }
  return refreshInFlight;
}

async function doRefresh(token: string): Promise<RefreshResult> {
  let res: Response;
  try {
    res = await fetch(`${ENV.VITE_BACKEND_URL}/auth/refresh`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
    });
  } catch {
    return { status: "failed" };
  }

  if (res.status === 401) return { status: "unauthorized" };
  if (!res.ok) return { status: "failed" };

  const raw: unknown = await res.json().catch(() => null);
  const newToken = (raw as { data?: { token?: unknown } } | null)?.data?.token;
  if (typeof newToken === "string" && newToken) return { status: "ok", token: newToken };
  return { status: "failed" };
}
