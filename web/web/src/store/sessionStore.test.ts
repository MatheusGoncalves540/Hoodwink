import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createInput, DEFAULT_MAX_PLAYERS, useSessionStore, usernameFromToken } from "./sessionStore";

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

  it("lê o username do JWT e sobrevive a token inválido", () => {
    const token = `h.${btoa(JSON.stringify({ id: "me", username: "Matheus" }))}.s`;
    expect(usernameFromToken(token)).toBe("Matheus");
    expect(usernameFromToken("nada")).toBe("");
  });
});
