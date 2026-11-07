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
  startTime: "2026-11-04T00:00:00Z",
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
      "Carla se disse Guardiã. Ninguém duvidou.",
      "Ana puxou uma carta nova como Crupiê.",
      "Bruno se disse Assassino, Carla duvidou. Bruno perdeu a carta e Carla escolhe qual cai.",
      "Ana matou uma carta de Davi como Assassino.",
      "Davi saiu da partida.",
      "Bruno roubou 3 de Eva como Trapaceiro.",
    ]);
  });

  it("ações do turno e limite de 20, com os nomes reais do servidor (02 §6)", () => {
    const log = run([
      room(6, "EARN_COINS", "a", { cause: "TAKE_COINS", earnedCoins: 2 }),
      room(6, "CARD_KILLED", "b", { cause: "KILL_CARD", targetPlayer: "c", targetCardIndex: 0 }),
      room(6, "CHANGED_CARD", "d", { cause: "SWAP_CARDS" }),
      room(7, "CARD_KILLED", "e", { cause: "GREED", targetPlayer: "e", targetCardIndex: 1 }),
    ]);
    expect(log.map(logText)).toEqual([
      "Ana pegou 2 moedas.",
      "Bruno pagou e matou uma carta de Carla.",
      "Davi trocou uma carta com o baralho.",
      "Eva passou de 20 moedas e perdeu uma carta.",
    ]);
  });

  it("guarda ator, rodada e papel; a gaveta usa o predicado", () => {
    const log = run([room(4, "CARD_PLAYED_TRILLIONAIRE", "e"), room(4, "EARN_COINS", "e", { cause: "TRILLIONAIRE", earnedCoins: 4 })]);
    expect(log).toEqual([
      { key: "CARD_PLAYED_TRILLIONAIRE|e|4024", turn: 4, at: 0, actorId: "e", actorName: "Eva", pred: "se disse Trilionário. Ninguém duvidou.", open: undefined, role: "TRILLIONAIRE" },
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
    expect(log.map(logText)).toEqual(["Carla se disse Guardiã, Ana duvidou. Carla mostrou a carta e escolhe uma carta de Ana."]);
  });

  it("contestação perdida na reconexão ainda fecha a declaração com o desfecho", () => {
    const log = run([
      room(6, "CARD_PLAYED_REBEL", "b"),
      room(6, "CONTEST_PENALTY", "a", { contestedPlayer: "b", hasCard: false }),
    ]);
    expect(log.map(logText)).toEqual(["Bruno se disse Rebelde, Ana duvidou. Bruno perdeu a carta e Ana escolhe qual cai."]);
  });

  it("troca de carta por contestação não se disfarça de Crupiê", () => {
    const log = run([
      room(7, "CARD_PLAYED_CROUPIER", "c"),
      room(7, "CHANGED_CARD", "c", { cause: "CONTEST_PENALTY" }),
    ]);
    expect(log.map(logText)).toEqual([
      "Carla se disse Crupiê. Ninguém duvidou.",
      "Carla trocou a carta mostrada depois da dúvida.",
    ]);
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
