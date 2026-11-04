import { describe, expect, it, vi } from "vitest";
import { parseRoom, playOrder } from "./protocol";

// Formato real de PublicRoomForUpdates (rooms/room.go), com gameEvent na grafia do Go.
const serverMessage = {
  id: "K7Q2",
  turn: 4,
  tax: 1,
  doubledCardValues: {
    POLITICAL: { timesValueDoubled: 1, roundsUntilDecrease: 3, usedThisTurn: false },
  },
  players: {
    p2: { id: "p2", name: "Ana", moveTime: 2, cards: [{ name: "CROUPIER", index: 0, protected: false, dead: false }, { index: 1, protected: false, dead: false }], coins: 7, investments: null, pendingInvestmentCoin: false, alive: true },
    p1: { id: "p1", name: "Você", moveTime: 1, cards: [{ name: "GUARDIAN", index: 0, protected: true, dead: false }, { name: "TRILLIONAIRE", index: 1, protected: false, dead: false }], coins: 7, investments: [2], pendingInvestmentCoin: false, alive: true },
  },
  deck: 12,
  currentPlayer: "p2",
  gameEvent: { PlayerID: "p2", Type: "CARD_PLAYED_ASSASSIN", ExpiresAt: "2026-11-04T21:38:10Z", Payload: { targetPlayer: "p1", targetCardIndex: 1 } },
  pendingEffects: [],
  gameOver: false,
  startTime: "2026-11-04T21:30:00Z",
  chat: null,
};

describe("parseRoom", () => {
  it("normaliza a mensagem real do servidor", () => {
    expect(parseRoom(serverMessage)).toEqual({
      id: "K7Q2",
      name: "",
      turn: 4,
      tax: 1,
      doubledCardValues: { POLITICAL: { timesValueDoubled: 1, roundsUntilDecrease: 3, usedThisTurn: false } },
      players: {
        p2: { id: "p2", name: "Ana", moveTime: 2, cards: [{ name: "CROUPIER", index: 0, protected: false, dead: false }, { index: 1, protected: false, dead: false }], coins: 7, investments: [], pendingInvestmentCoin: false, alive: true },
        p1: { id: "p1", name: "Você", moveTime: 1, cards: [{ name: "GUARDIAN", index: 0, protected: true, dead: false }, { name: "TRILLIONAIRE", index: 1, protected: false, dead: false }], coins: 7, investments: [2], pendingInvestmentCoin: false, alive: true },
      },
      deck: 12,
      currentPlayer: "p2",
      gameEvent: { playerId: "p2", type: "CARD_PLAYED_ASSASSIN", expiresAt: Date.parse("2026-11-04T21:38:10Z"), payload: { targetPlayer: "p1", targetCardIndex: 1 } },
      gameOver: false,
      chat: [],
    });
  });

  it("aceita gameEvent em camelCase e deck como lista (modo depuração)", () => {
    const r = parseRoom({ ...serverMessage, deck: ["ASSASSIN", "REBEL"], gameEvent: { playerId: "p1", type: "WAITING_FIRST_ACTION", expiresAt: "2026-11-04T21:38:10Z", payload: null } });
    expect(r?.deck).toBe(2);
    expect(r?.gameEvent?.type).toBe("WAITING_FIRST_ACTION");
  });

  it("rejeita mensagem com campo faltando sem lançar", () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
    const { players: _omit, ...broken } = serverMessage;
    expect(parseRoom(broken)).toBeNull();
    expect(parseRoom("não é json")).toBeNull();
    expect(parseRoom({ ...serverMessage, gameEvent: { ...serverMessage.gameEvent, ExpiresAt: "ontem" } })).toBeNull();
    expect(warn).toHaveBeenCalledTimes(3);
    warn.mockRestore();
  });

  it("ordena jogadores por moveTime", () => {
    expect(playOrder(parseRoom(serverMessage)!).map((p) => p.id)).toEqual(["p1", "p2"]);
  });
});
