/**
 * Estados de exemplo idênticos aos das pranchas (design-handoff/referencias/pranchas/*.dc.html, <script>).
 * `fixture(name, now)` monta { me, room, log, ui } com os timers relativos a `now` (ms), para que
 * a contagem regressiva funcione no navegador e os testes fiquem determinísticos.
 */
import type { LogEntry } from "./log";
import type { PublicCard, PublicPlayer, PublicRoom, RoleId } from "./protocol";

export type OpenSheet = null | "menu" | "chat" | "you" | { seat: string };
export type Fixture = {
  me: string;
  room: PublicRoom | null;
  log: LogEntry[];
  ui: { openSheet: OpenSheet; chatSeen: number };
};

export const FIXTURE_NAMES = [
  "Lobby", "Sala", "SalaCheia",
  "Mesa6", "Mesa3", "Mesa10",
  "Camarote", "CamaroteLongo", "Menu", "Conversa", "ConversaCheia", "Voce", "VoceLongo",
  "Acoes", "Main", "ContestaBlefe", "ContestaVerdade",
  "ForaDeCena", "MorteSubita", "MesaSubita", "Fim",
] as const;
export type FixtureName = (typeof FIXTURE_NAMES)[number];

type C = { name?: RoleId; protected?: boolean; dead?: boolean };
const cards = (cs: C[]): PublicCard[] => cs.map((c, index) => ({ name: c.name, index, protected: !!c.protected, dead: !!c.dead }));
const player = (id: string, name: string, moveTime: number, coins: number, cs: C[], extra: Partial<PublicPlayer> = {}): PublicPlayer => ({
  id, name, moveTime, coins, cards: cards(cs), investments: [], pendingInvestmentCoin: false, alive: true, ...extra,
});

// Mesa6.dc.html → roster, na ordem de jogo depois de você
const ROSTER: PublicPlayer[] = [
  player("ana", "Ana", 1, 7, [{ name: "CROUPIER" }, {}]),
  player("carla", "Carla", 2, 11, [{ protected: true }, {}]),
  player("bruno", "Bruno", 3, 2, [{ name: "ASSASSIN", dead: true }, {}]),
  player("davi", "Davi", 4, 0, [{ name: "TRICKSTER", dead: true }, { name: "GRAVEDIGGER", dead: true }], { alive: false }),
  player("eva", "Eva", 5, 4, [{}, {}], { investments: [3, 5] }),
  player("max", "Maximiliano Albuquerque", 6, 20, [{}, { name: "INVESTOR" }]),
  player("ira", "Iracema", 7, 13, [{ protected: true }, { name: "POLITICAL" }]),
  player("flo", "Florisvaldo", 8, 1, [{ name: "REBEL", dead: true }, {}]),
  player("jo", "Jô", 9, 9, [{}, {}]),
];
const ME = player("me", "Matheus", 0, 7, [{ name: "GUARDIAN", protected: true }, { name: "TRILLIONAIRE" }], { investments: [2] });

const local = (h: number, m: number) => new Date(2026, 10, 4, h, m).getTime();
const iso = (h: number, m: number) => new Date(local(h, m)).toISOString();

const entry = (actorId: string, turn: number, pred: string, at: number, extra: Partial<LogEntry> = {}): LogEntry => ({
  key: `fx|${actorId}|${turn}|${pred}`,
  turn,
  at,
  actorId,
  actorName: [...ROSTER, ME].find((p) => p.id === actorId)?.name ?? "",
  pred,
  ...extra,
});

// "O que a mesa já viu" de Bruno (Mesa6.dc.html → logs)
const BRUNO_LOG = [
  entry("bruno", 1, "recebeu 4 como Trilionário.", local(21, 30), { role: "TRILLIONAIRE" }),
  entry("bruno", 2, "roubou 3 de Eva como Trapaceiro.", local(21, 34)),
  entry("bruno", 3, "se disse Assassino, foi contestado e perdeu a carta.", local(21, 39) + 30_000, { role: "ASSASSIN", chat: true }),
];

// CamaroteLongo.dc.html → logs de Maximiliano, 14 jogadas
const MAX_LOG: [number, string, RoleId?][] = [
  [1, "recebeu 4 como Trilionário.", "TRILLIONAIRE"], [2, "protegeu uma carta como Guardiã.", "GUARDIAN"],
  [3, "subiu a taxa como Político.", "POLITICAL"], [4, "bloqueou as moedas de Jô como Egoísta.", "SELFISH"],
  [5, "recebeu 4 como Trilionário.", "TRILLIONAIRE"], [6, "trocou uma carta como Crupiê.", "CROUPIER"],
  [7, "investiu como Investidor.", "INVESTOR"], [8, "recebeu 2 como Trapaceiro.", "TRICKSTER"],
  [9, "deixou passar o Assassino de Iracema."], [10, "protegeu uma carta como Guardiã.", "GUARDIAN"],
  [11, "foi contestada como Guardiã e mostrou a carta.", "GUARDIAN"], [12, "investiu como Investidor.", "INVESTOR"],
  [13, "recebeu 4 como Trilionário.", "TRILLIONAIRE"], [14, "contestado ao se dizer Político. Perdeu a carta.", "POLITICAL"],
];

// Voce.dc.html → "Papéis que a mesa já viu você usar"
const MY_LOG = [
  entry("me", 1, "recebeu 4 moedas.", local(21, 31), { role: "TRILLIONAIRE" }),
  entry("me", 2, "protegeu uma carta.", local(21, 35), { role: "GUARDIAN" }),
];
// VoceLongo.dc.html
const MY_LONG_LOG: [number, string, RoleId][] = [
  [3, "mexeu na taxa.", "POLITICAL"], [4, "bloqueou as moedas de Jô.", "SELFISH"], [6, "trocou uma carta.", "CROUPIER"],
  [8, "recebeu 2 moedas.", "TRICKSTER"], [10, "contestada, mostrou a carta; protegeu 2 vezes.", "GUARDIAN"],
  [11, "investiu 2 vezes.", "INVESTOR"], [13, "recebeu 12 moedas.", "TRILLIONAIRE"],
];

// Conversa.dc.html
const CHAT = [
  { playerName: "Carla", timestamp: iso(21, 38), message: "Ninguém aqui tem Guardiã, né?" },
  { playerName: "Ana", timestamp: iso(21, 39), message: "Esse Crupiê é meu, juro pela cortina." },
  { playerName: "Bruno", timestamp: iso(21, 40), message: "Isso foi pessoal." },
  { playerName: "Matheus", timestamp: iso(21, 41), message: "Foi teatro, Bruno." },
  { playerName: "Eva", timestamp: iso(21, 41), message: "Alguém conta as moedas da Carla, por favor." },
];

// ConversaCheia.dc.html: o mesmo ciclo de falas, de 22:10 a 22:39, com Florisvaldo contestado a cada 7 min
const LONG_LINES: [string, string][] = [
  ["Maximiliano Albuquerque", "kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk"],
  ["Iracema", "Eu vou contestar, ninguém me segura, já avisei que esse Político é mentira desde a rodada dois e vocês não ouviram."],
  ["Jô", "ok"],
  ["Florisvaldo", "Quem tem o Coveiro?"],
  ["Carla", "Mentira."],
  ["Ana", "Juro."],
  ["Matheus", "Mentira."],
];

type TableOpts = {
  n: number;
  active: string;
  event?: { type: string; playerId?: string; left?: number; payload?: unknown };
  deck?: number;
  turn?: number;
  me?: PublicPlayer;
  doubled?: Record<string, [number, number]>;
  gameOver?: boolean;
  roster?: PublicPlayer[];
};

function table(now: number, o: TableOpts): PublicRoom {
  const players = [o.me ?? ME, ...(o.roster ?? ROSTER).slice(0, o.n - 1)];
  const doubled = o.doubled ?? { POLITICAL: [1, 3], REBEL: [2, 1] };
  const ev = o.event ?? { type: "WAITING_FIRST_ACTION" };
  return {
    id: "K7Q2",
    name: "Quinta dos mentirosos",
    turn: o.turn ?? 4,
    tax: 1,
    doubledCardValues: Object.fromEntries(
      Object.entries(doubled).map(([k, [t, r]]) => [k, { timesValueDoubled: t, roundsUntilDecrease: r, usedThisTurn: false }]),
    ),
    players: Object.fromEntries(players.map((p) => [p.id, p])),
    deck: o.deck ?? 24 - 2 * o.n,
    currentPlayer: o.active,
    gameEvent: { type: ev.type, playerId: ev.playerId ?? o.active, expiresAt: now + (ev.left ?? 8) * 1000, payload: ev.payload ?? null },
    gameOver: !!o.gameOver,
    startTime: new Date(now).toISOString(),
    chat: CHAT,
  };
}

const waiting = (now: number, names: string[], max: number, roomName: string): PublicRoom => ({
  id: "K7Q2",
  name: roomName,
  maxPlayers: max,
  turn: 0,
  tax: 0,
  doubledCardValues: {},
  players: Object.fromEntries(
    names.map((name, i) => {
      const p = i === 0 ? { ...ME, cards: [] } : player(`p${i}`, name, i, 3, []);
      return [p.id, p];
    }),
  ),
  deck: 24,
  currentPlayer: "",
  gameEvent: { type: "WAITING_START", playerId: "", expiresAt: now, payload: null },
  gameOver: false,
  startTime: "",
  chat: [],
});

const SALA_NAMES = ["Matheus", "Ana", "Bruno", "Maximiliano Albuquerque de Souza", "Iracema", "Florisvaldo", "Jô", "Carla", "Davi"];

export function fixture(name: FixtureName, now: number): Fixture {
  const base: Fixture = { me: "me", room: table(now, { n: 6, active: "carla" }), log: BRUNO_LOG, ui: { openSheet: null, chatSeen: 3 } };
  const ui = (openSheet: OpenSheet, chatSeen = 3) => ({ openSheet, chatSeen });
  const anaPlays = (event: TableOpts["event"]) => table(now, { n: 6, active: "ana", event });
  switch (name) {
    case "Lobby":
      return { ...base, room: null, log: [] };
    case "Sala":
      return { ...base, room: waiting(now, SALA_NAMES.slice(0, 3), 6, "Quinta dos mentirosos"), log: [] };
    case "SalaCheia":
      return { ...base, room: waiting(now, SALA_NAMES, 10, "A noite em que todos mentiram"), log: [] };
    case "Mesa6":
      return base;
    case "Mesa3":
      return { ...base, room: table(now, { n: 3, active: "carla" }), log: [] };
    case "Mesa10":
      return { ...base, room: table(now, { n: 10, active: "carla" }) };
    case "Camarote":
      return { ...base, ui: ui({ seat: "bruno" }) };
    case "CamaroteLongo":
      return {
        ...base,
        room: table(now, { n: 10, active: "max" }),
        log: MAX_LOG.map(([t, pred, role], i) => entry("max", t, pred, local(21, 0) + i * 60_000, { role })),
        ui: ui({ seat: "max" }),
      };
    case "Menu":
      return { ...base, ui: ui("menu") };
    case "Conversa":
      return { ...base, ui: ui("chat", CHAT.length) };
    case "ConversaCheia": {
      const chat = Array.from({ length: 30 }, (_, i) => i)
        .filter((i) => (i + 4) % 7 !== 0)
        .map((i) => ({ playerName: LONG_LINES[i % 7][0], timestamp: iso(22, 10 + i), message: LONG_LINES[i % 7][1] }));
      const flo = [3, 10, 17, 24].map((i, k) =>
        entry("flo", 5 + k, "se disse Rebelde, foi contestado e perdeu a carta.", local(22, 10 + i), { role: "REBEL", chat: true, key: `fx|flo|${k}` }),
      );
      return { ...base, room: { ...table(now, { n: 10, active: "carla" }), chat }, log: flo, ui: ui("chat", chat.length) };
    }
    case "Voce":
      return { ...base, log: [...BRUNO_LOG, ...MY_LOG], ui: ui("you") };
    case "VoceLongo":
      return {
        ...base,
        room: table(now, { n: 10, active: "carla" }),
        log: [...MY_LOG, ...MY_LONG_LOG.map(([t, pred, role]) => entry("me", t, pred, local(21, 40 + t), { role }))],
        ui: ui("you"),
      };
    case "Acoes":
      return {
        ...base,
        room: table(now, { n: 6, active: "me", me: { ...ME, coins: 5 }, event: { type: "WAITING_FIRST_ACTION", left: 9 }, doubled: {} }),
      };
    case "Main":
      return { ...base, room: anaPlays({ type: "CARD_PLAYED_ASSASSIN", left: 7, payload: { targetPlayer: "me", targetCardIndex: 1 } }) };
    case "ContestaBlefe":
      return {
        ...base,
        room: anaPlays({ type: "CONTEST_PENALTY", playerId: "me", left: 6, payload: { contestedPlayer: "ana", contestedCard: "ASSASSIN", hasCard: false } }),
      };
    case "ContestaVerdade":
      return {
        ...base,
        room: anaPlays({ type: "CONTEST_PENALTY", playerId: "me", left: 6, payload: { contestedPlayer: "ana", contestedCard: "ASSASSIN", hasCard: true } }),
      };
    case "ForaDeCena": {
      const out = { ...ME, alive: false, investments: [], cards: cards([{ name: "GUARDIAN", dead: true }, { name: "TRILLIONAIRE", dead: true }]) };
      return {
        ...base,
        room: table(now, { n: 6, active: "carla", me: out }),
        log: [...BRUNO_LOG, entry("me", 14, "saiu de cena.", local(21, 42), { key: "EXIT|me", chat: true })],
      };
    }
    case "MorteSubita":
    case "MesaSubita":
      return { ...base, room: table(now, { n: 4, active: "ana", deck: 0, turn: 19, event: { type: "WAITING_FIRST_ACTION", left: 4 } }) };
    case "Fim": {
      const dead = (p: PublicPlayer) => ({ ...p, alive: false });
      const roster = ROSTER.slice(0, 5).map((p) => (p.id === "carla" ? { ...p, coins: 14 } : dead(p)));
      const exits: [string, number][] = [["davi", 5], ["bruno", 9], ["eva", 12], ["me", 14], ["ana", 17]];
      return {
        ...base,
        room: table(now, { n: 6, active: "carla", roster, me: dead(ME), turn: 17, gameOver: true }),
        log: exits.map(([id, t], i) => entry(id, t, "saiu de cena.", local(22, i), { key: `EXIT|${id}`, chat: true })),
      };
    }
  }
}
