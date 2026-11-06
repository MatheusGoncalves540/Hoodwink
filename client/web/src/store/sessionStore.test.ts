import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createInput,
  DEFAULT_MAX_PLAYERS,
  playerIdFromToken,
  roomCodeFromSearch,
  useSessionStore,
  usernameFromToken,
} from "./sessionStore";

const json = (body: unknown, status = 200) => new Response(JSON.stringify(body), { status });
const s = () => useSessionStore.getState();

beforeEach(() => s().reset());
afterEach(() => vi.unstubAllGlobals());

describe("sessionStore", () => {
  it("cria a sala, pede o ticket e guarda os dois", async () => {
    const fetch = vi.fn(async (url: string) =>
      String(url).endsWith("/newRoom") ? json({ RoomId: "k7q2" }, 201) : json({ data: { ticket: "ticket.jwt" } }),
    );
    vi.stubGlobal("fetch", fetch);

    await expect(s().create("jwt", createInput("Matheus"))).resolves.toBe("k7q2");
    expect({ roomCode: s().roomCode, ticket: s().ticket, busy: s().busy, error: s().error }).toEqual({
      roomCode: "k7q2",
      ticket: "ticket.jwt",
      busy: false,
      error: null,
    });
    expect(fetch.mock.calls.map(([url]) => String(url).split("4444")[1])).toEqual(["/newRoom", "/getTicket/k7q2"]);
  });

  it("entrar com sala inexistente guarda a mensagem do servidor e não abre sessão", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => json({ message: "Sala não encontrada" }, 401)));
    await expect(s().join("jwt", "zzzz")).resolves.toBeNull();
    expect(s().error).toBe("Sala não encontrada");
    expect({ roomCode: s().roomCode, ticket: s().ticket, busy: s().busy }).toEqual({ roomCode: null, ticket: null, busy: false });
  });

  it("nome da sala não passa de 30 caracteres e usa o padrão de 6 lugares", () => {
    expect(createInput("Maximiliano Albuquerque de Souza")).toEqual({
      roomName: "Sala de Maximiliano Albuquerqu",
      maxPlayers: DEFAULT_MAX_PLAYERS,
    });
  });

  it("lê o código do convite na URL no formato hexa do servidor", () => {
    expect(roomCodeFromSearch("?sala=1A2B3C")).toBe("1a2b3c");
    expect(roomCodeFromSearch("?cena=table&sala=%20a1b2%20")).toBe("a1b2");
    expect(roomCodeFromSearch("?sala=zzzz")).toBeNull();
    expect(roomCodeFromSearch("?sala=abc")).toBeNull();
    expect(roomCodeFromSearch("")).toBeNull();
  });

  it("lê o username do JWT e sobrevive a token inválido", () => {
    const token = `h.${btoa(JSON.stringify({ id: "me", username: "Matheus" }))}.s`;
    expect(usernameFromToken(token)).toBe("Matheus");
    expect(usernameFromToken("nada")).toBe("");
  });

  it("lê nome com acento no JWT (UTF-8)", () => {
    const bytes = new TextEncoder().encode(JSON.stringify({ username: "João" }));
    const token = `h.${btoa(String.fromCharCode(...bytes))}.s`;
    expect(usernameFromToken(token)).toBe("João");
  });

  it("lê o playerId do ticket (ou do JWT do backend, no login)", () => {
    const ticket = `h.${btoa(JSON.stringify({ playerId: "eu", roomId: "a1b2" }))}.s`;
    const login = `h.${btoa(JSON.stringify({ id: "eu", username: "Matheus" }))}.s`;
    expect(playerIdFromToken(ticket)).toBe("eu");
    expect(playerIdFromToken(login)).toBe("eu");
    expect(playerIdFromToken("nada")).toBe("");
  });
});
