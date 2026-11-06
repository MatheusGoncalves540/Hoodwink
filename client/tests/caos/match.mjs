/**
 * Partida completa de ponta a ponta: cria a sala, conecta 3 jogadores,
 * inicia o jogo e joga até gameOver. Sempre que possível contesta, para
 * acelerar a resolução (cada contestação derruba uma carta).
 */
import { connect, jwt, newRoom, ticketJwt } from "./lib.mjs";

const run = Date.now().toString(36);
const created = await newRoom(jwt(`caos-ana-${run}`, "Ana Caos"), { roomName: "Partida caos" });
const roomId = created.json.RoomId;
const players = [
  { id: `caos-ana-${run}`, name: "Ana Caos" },
  { id: `caos-bruno-${run}`, name: "Bruno Caos" },
  { id: `caos-carla-${run}`, name: "Carla Caos" },
];

const NO_TARGET = new Set(["TRILLIONAIRE", "POLITICAL", "REBEL", "INVESTOR"]);
const OTHER = new Set(["ASSASSIN", "CLAIRVOYANT", "TRICKSTER"]);
const OTHER_CARD = new Set(["KAMIKAZE"]);
const OWN = new Set(["GUARDIAN", "CROUPIER", "GRAVEDIGGER"]);

const state = players.map((p) => ({ ...p, socket: null, hand: new Map(), done: new Set(), started: false, lastKey: "" }));
const eventKey = (room) => {
  const e = room?.gameEvent;
  if (!e) return "";
  return `${e.Type}|${e.ExpiresAt}|${e.PlayerID}|${JSON.stringify(e.Payload ?? null)}`;
};
const alive = (room) => Object.values(room.players).filter((p) => p.alive);
const nextAlive = (room, me) => alive(room).find((p) => p.id !== me) ?? null;
const firstFreeCard = (player) => player?.cards?.find((c) => !c.dead)?.index ?? 0;

function rememberHand(me, room) {
  const mine = room.players[me.id];
  for (const card of mine?.cards ?? []) if (card.name) me.hand.set(card.index, card.name);
}

function actionFor(me, room) {
  const ev = room.gameEvent;
  if (!ev) return null;
  const type = ev.Type;
  const actor = ev.PlayerID;
  const payload = ev.Payload ?? {};
  const myPlayer = room.players[me.id];
  if (!myPlayer?.alive) return null;

  if (type === "WAITING_FIRST_ACTION") {
    if (room.currentPlayer !== me.id) return null;
    rememberHand(me, room);
    const names = [...me.hand.entries()].filter(([index]) => !myPlayer.cards.find((c) => c.index === index)?.dead);
    const noTarget = names.find(([, name]) => NO_TARGET.has(name));
    if (noTarget) return { type: noTarget[1], payload: null };
    const other = nextAlive(room, me.id);
    if (!other) return null;
    const otherCard = firstFreeCard(other);
    const otherCardRole = names.find(([, name]) => OTHER_CARD.has(name));
    if (otherCardRole) return { type: otherCardRole[1], payload: { targetPlayer: other.id, targetCardIndex: otherCard, targetAllyCardIndex: firstFreeCard(myPlayer) } };
    const ranged = names.find(([, name]) => OTHER.has(name));
    if (ranged) {
      const base = { targetPlayer: other.id, targetCardIndex: otherCard };
      return { type: ranged[1], payload: ranged[1] === "CLAIRVOYANT" ? { ...base, showToAllPlayers: false } : base };
    }
    const own = names.find(([, name]) => OWN.has(name));
    if (own) {
      const ownCard = own[1] === "GRAVEDIGGER" ? myPlayer.cards.find((c) => c.dead)?.index : firstFreeCard(myPlayer);
      if (ownCard != null) return { type: own[1], payload: { targetPlayer: me.id, targetCardIndex: ownCard, ...(own[1] === "CROUPIER" ? { useOnTwoCards: false } : {}) } };
    }
    return null;
  }

  if (type.startsWith("CARD_PLAYED") && actor !== me.id) {
    return { type: "CONTEST", payload: {} };
  }

  if (type === "CONTEST_PENALTY") {
    const chooser = payload.hasCard ? payload.contestedPlayer : actor;
    if (chooser !== me.id) return null;
    const victimId = payload.hasCard ? actor : payload.contestedPlayer;
    const victim = room.players[victimId];
    const card = victim?.cards?.find((c) => !c.dead)?.index;
    if (card == null) return null;
    return { type: "CONTEST_PENALTY", payload: { TargetCardIndex: card } };
  }
  return null;
}

let finished = false;
const deadline = Date.now() + 240_000;
const log = [];
for (const me of state) {
  me.socket = connect(ticketJwt(me.id, me.name, roomId));
  await me.socket.waitOpen;
}
await state[0].socket.waitFor((m) => m && m.players && Object.keys(m.players).length === 3, 15000);
state[0].socket.send({ type: "COMMAND_START_GAME", payload: {} });
log.push({ at: Date.now(), who: state[0].name, act: "START" });
for (const me of state) for (const m of me.socket.messages) if (m?.players?.[me.id]) rememberHand(me, m);
console.log("sala", roomId, "conectados", state.length);

for (const me of state) {
  const listener = () => {
    const room = [...me.socket.messages].reverse().find((m) => m && m.players);
    if (!room || finished) return;
    rememberHand(me, room);
    if (room.gameOver) {
      if (!finished) {
        finished = true;
        const winner = alive(room)[0]?.name ?? "ninguém";
        console.log("FIM", { turno: room.turn, vencedor: winner, eventos: log.length });
        for (const s of state) s.socket.close();
      }
      return;
    }
    const key = eventKey(room);
    if (key === me.lastKey) return;
    me.lastKey = key;
    if (!room.gameEvent) return;
    if (room.gameEvent.Type === "WAITING_START") return;
    const action = actionFor(me, room);
    if (!action) return;
    me.done.add(key);
    me.socket.send(action);
    log.push({ at: Date.now(), who: me.name, act: action.type, ev: room.gameEvent.Type, turn: room.turn });
  };
  me.socket.ws.addEventListener("message", listener);
}

while (!finished && Date.now() < deadline) await new Promise((r) => setTimeout(r, 250));
if (!finished) {
  const room = [...state[0].socket.messages].reverse().find((m) => m && m.players);
  console.log("TIMEOUT", { turno: room?.turn, evento: room?.gameEvent, vivos: room ? alive(room).map((p) => p.name) : [] });
  for (const s of state) s.socket.close();
  process.exitCode = 1;
}
console.log("ações", log.length, log.slice(-12));
