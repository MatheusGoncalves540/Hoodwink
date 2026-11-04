/**
 * Sessão de entrada (04-arquitetura.md → sem roteador): criar sala ou entrar pelo código.
 * Guarda só o que existe antes da sala chegar — código, ticket e o erro da última tentativa.
 * O ticket abre o GameSocket (T25); a cena em si continua saindo do gameStore.
 */
import { create } from "zustand";
import { getTicket, newRoom, type NewRoomInput } from "../net/api";

/** ClassicRules.yaml → General.maxPlayers e CreateRoomRequest (2..10). */
export const DEFAULT_MAX_PLAYERS = 6;

/** Nome da sala cabe nos 30 caracteres do servidor (CreateRoomRequest.RoomName). */
export const createInput = (username: string): NewRoomInput => ({
  roomName: `Sala de ${username || "Você"}`.slice(0, 30),
  maxPlayers: DEFAULT_MAX_PLAYERS,
});

/** Código do convite na URL (`?sala=K7Q2`), no formato que o servidor emite: hexa minúsculo de 4 a 12. */
export function roomCodeFromSearch(search: string): string | null {
  const code = new URLSearchParams(search).get("sala")?.trim().toLowerCase() ?? "";
  return /^[0-9a-f]{4,12}$/.test(code) ? code : null;
}

/** Payload do JWT sem validar a assinatura (quem valida é o servidor). */
function tokenClaims(token: string): Record<string, unknown> {
  try {
    const payload = token.split(".")[1];
    return JSON.parse(atob(payload.replace(/-/g, "+").replace(/_/g, "/"))) as Record<string, unknown>;
  } catch {
    return {};
  }
}

/** O JWT do backend tem { id, username } (backend/structures/jwt.go). Sem token, sem nome. */
export function usernameFromToken(token: string): string {
  const username = tokenClaims(token).username;
  return typeof username === "string" ? username : "";
}

/** Id do jogador no jogo: o ticket tem { playerId, roomId } (ClaimsHoodwink); o JWT do backend tem { id }. */
export function playerIdFromToken(token: string): string {
  const { playerId, id } = tokenClaims(token);
  const value = playerId ?? id;
  return typeof value === "string" ? value : "";
}

type SessionState = {
  roomCode: string | null;
  ticket: string | null;
  busy: boolean;
  error: string | null;
  /** Cria a sala e pede o ticket. Devolve o código, ou null (o motivo fica em `error`). */
  create(token: string, input: NewRoomInput): Promise<string | null>;
  /** Entra numa sala existente. Devolve o código, ou null. */
  join(token: string, roomCode: string): Promise<string | null>;
  reset(): void;
};

const reason = (e: unknown) => (e instanceof Error ? e.message : String(e));

export const useSessionStore = create<SessionState>((set, get) => ({
  roomCode: null,
  ticket: null,
  busy: false,
  error: null,
  create: async (token, input) => {
    if (get().busy) return null;
    set({ busy: true, error: null });
    try {
      const roomCode = await newRoom(token, input);
      const ticket = await getTicket(token, roomCode);
      set({ roomCode, ticket, busy: false });
      return roomCode;
    } catch (e) {
      set({ busy: false, error: reason(e) });
      return null;
    }
  },
  join: async (token, roomCode) => {
    if (get().busy) return null;
    set({ busy: true, error: null });
    try {
      const ticket = await getTicket(token, roomCode);
      set({ roomCode, ticket, busy: false });
      return roomCode;
    } catch (e) {
      set({ busy: false, error: reason(e) });
      return null;
    }
  },
  reset: () => set({ roomCode: null, ticket: null, busy: false, error: null }),
}));
