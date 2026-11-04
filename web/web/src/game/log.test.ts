import { describe, expect, it } from "vitest";
import { appendLog, exitsOf, logSelfText, logText, type LogEntry } from "./log";
import type { PublicRoom } from "./protocol";

const NAMES = { a: "Ana", b: "Bruno", c: "Carla", d: "Davi", e: "Eva" } as const;
type Id = keyof typeof NAMES;

const room = (turn: number, type: string | null, playerId: Id = "a", payload: unknown = null, dead: Id[] = []): PublicRoom => ({
  id: "K7Q2",
  name: "",
  turn,
  tax: 1,
  doubledCardValues: {},
  players: Object.fromEntries(
    (Object.keys(NAMES) as Id[]).map((id, i) => [
      id,
      { id, name: NAMES[id], moveTime: i, cards: [], coins: 3, investments: [], pendingInvestmentCoin: false, alive: !dead.includes(id) },
    ]),
  ),
  deck: 12,
  currentPlayer: playerId,
  gameEvent: type ? { playerId, type, expiresAt: turn * 1000 + type.length, payload } : null,
  gameOver: false,
  chat: [],
});

const run = (rooms: PublicRoom[]) =>
  rooms.reduce<{ log: LogEntry[]; prev: PublicRoom | null }>(
    ({ log, prev }, r, i) => ({ log: appendLog(log, prev, r, i), prev: r }),
    { log: [], prev: null },
  ).log;

describe("appendLog", () => {
  it("gera as frases do diário de 05-protocolo.md", () => {
    const log = run([
      room(1, "CARD_PLAYED_GUARDIAN", "c"),
      room(1, "WAITING_FIRST_ACTION", "a"),
      room(1, "CHANGED_CARD", "a", { cause: "CROUPIER" }),
      room(2, "CARD_PLAYED_ASSASSIN", "b"),
      room(2, "CONTEST", "c", { contestedPlayer: "b", contestedCard: "ASSASSIN" }),
      room(2, "CONTEST_PENALTY", "c", { contestedPlayer: "b", contestedCard: "ASSASSIN", hasCard: false }),
      room(2, "CARD_KILLED", "c", { cause: "CONTEST_PENALTY", targetPlayer: "b", targetCardIndex: 0 }),
      room(3, "CARD_KILLED", "a", { cause: "ASSASSIN", targetPlayer: "d", targetCardIndex: 1 }, ["d"]),
      room(3, "STEAL_COINS", "b", { cause: "TRICKSTER", lostCoins: 3, targetPlayer: "e" }, ["d"]),
    ]);
    expect(log.map(logText)).toEqual([
      "Carla se disse Guardiã. Ninguém contestou.",
      "Ana puxou uma carta nova como Crupiê.",
      "Bruno se disse Assassino, foi contestado e perdeu a carta.",
      "Davi perdeu uma carta.",
      "Davi saiu de cena.",
      "Bruno roubou 3 de Eva como Trapaceiro.",
    ]);
  });

  it("guarda ator, rodada e papel; a gaveta usa o predicado", () => {
    const log = run([room(4, "CARD_PLAYED_TRILLIONAIRE", "e"), room(4, "EARN_COINS", "e", { cause: "TRILLIONAIRE", earnedCoins: 4 })]);
    expect(log).toEqual([
      { key: "CARD_PLAYED_TRILLIONAIRE|e|4024", turn: 4, at: 0, actorId: "e", actorName: "Eva", pred: "se disse Trilionário. Ninguém contestou.", open: undefined, role: "TRILLIONAIRE" },
      { key: "EARN_COINS|e|4010", turn: 4, at: 1, actorId: "e", actorName: "Eva", pred: "recebeu 4 como Trilionário." },
    ]);
    expect(logSelfText(log[1])).toBe("Recebeu 4 como Trilionário.");
  });

  it("contestação em que o contestado tinha a carta", () => {
    const log = run([
      room(5, "CARD_PLAYED_GUARDIAN", "c"),
      room(5, "CONTEST", "a", { contestedPlayer: "c" }),
      room(5, "CONTEST_PENALTY", "a", { contestedPlayer: "c", hasCard: true }),
    ]);
    expect(log.map(logText)).toEqual(["Carla se disse Guardiã, foi contestado e mostrou a carta."]);
  });

  it("a mesma sala reaplicada (reconexão) não duplica linhas", () => {
    const r = room(1, "CARD_PLAYED_REBEL", "a");
    const once = appendLog([], null, r, 0);
    expect(appendLog(once, null, r, 5)).toEqual(once);
    expect(appendLog(once, r, r, 5)).toEqual(once);
  });

  it("sem evento e sem mudança: diário igual", () => {
    expect(appendLog([], null, room(1, null), 0)).toEqual([]);
  });

  it("exitsOf lista as saídas em ordem", () => {
    const log = run([room(1, null), room(5, "CARD_KILLED", "a", { targetPlayer: "d" }, ["d"]), room(9, "CARD_KILLED", "a", { targetPlayer: "b" }, ["d", "b"])]);
    expect(exitsOf(log)).toEqual([
      { playerId: "d", name: "Davi", turn: 5 },
      { playerId: "b", name: "Bruno", turn: 9 },
    ]);
  });
});
