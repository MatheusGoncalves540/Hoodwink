/**
 * View models: o que as telas consomem. Nenhuma tela lê `room` cru. Funções puras de `GameView`.
 *
 * GameView = { me, room, log } (o mesmo formato do gameStore).
 *
 * selectScene(v)       → 'lobby' | 'room' | 'table' | 'finale'
 * selectSeats(v)       → SeatVM[] na ordem de jogo a partir de quem vem depois de você:
 *                        { id, name, short, coins, cards: MiniCard[2], inScene, out, hasMore, label }
 * selectStage(v)       → { activeId, activeName, fala, declared: null | { actorId, actorName, role, phase, targetId } }
 *                        phase: 'declared' (CARD_PLAYED_*), 'contested' (CONTEST), 'bluff' | 'truth' (CONTEST_PENALTY)
 * selectFootlights(v)  → { deck, tax, suddenDeath, effects: [{ role, name, times, rounds, label }] }
 * selectYou(v)         → { name, coins, limit, cards: YouCard[], investments, nextInvestment, alive, timer }
 *                        timer: { expiresAt, total, ownerName } | null (o número na tela sai de expiresAt − agora)
 * selectDecision(v)    → qual painel mostrar (tabela em 05-protocolo.md):
 *                        none | myTurn | react | pickCardToKill | waitPick | waitOther | outOfScene | finale
 *                        (o aviso de morte súbita é de UI: toca uma vez quando selectFootlights.suddenDeath liga)
 * selectLog(v, id?)    → [{ key, turn, text }] do mais novo ao mais velho; com id, só as linhas daquele ator,
 *                        no texto da gaveta ("Roubou 3 de Eva como Trapaceiro.")
 * selectRolesUsed(v, id) → papéis declarados por um ator: [{ role, name, turn, text }], o mais recente primeiro
 * selectChat(v)        → as 30 últimas falas e acontecimentos, em ordem: { kind: 'line' | 'event', … }
 * selectFinale(v)      → { winnerName, round, coins, exits: [{ place, name, round, me }], fala }
 * selectWaitingRoom(v) → { code, name, max, seats, castLine, isHost }
 */
import { copy, firstName } from "./copy";
import { exitsOf, logSelfText, logText, type LogEntry } from "./log";
import { playOrder, type PublicCard, type PublicRoom, type RoleId } from "./protocol";
import { cardValue } from "./pricing";
import { declaredRole, roleName, ROLES } from "./roles";
import { CLASSIC_RULES } from "./rules.classic";

export type GameView = { me: string | null; room: PublicRoom | null; log: LogEntry[] };

export type MiniCard = { state: "back" | "revealed" | "protected" | "dead"; role?: RoleId };
export type SeatVM = {
  id: string;
  name: string;
  short: string;
  coins: number;
  cards: MiniCard[];
  inScene: boolean;
  out: boolean;
  hasMore: boolean;
  label: string;
};

const miniCard = (c: PublicCard): MiniCard =>
  c.dead ? { state: "dead", role: c.name } : c.name ? { state: "revealed", role: c.name } : c.protected ? { state: "protected" } : { state: "back" };

const fname = (v: GameView, id: string) => firstName(v.room?.players[id]?.name ?? "");
const payload = (v: GameView) => (v.room?.gameEvent?.payload ?? {}) as Record<string, unknown>;

export function selectScene(v: GameView): "lobby" | "room" | "table" | "finale" {
  if (!v.room) return "lobby";
  if (v.room.gameOver) return "finale";
  if (!v.room.gameEvent || v.room.gameEvent.type === "WAITING_START") return "room";
  return "table";
}

/** Adversários na ordem de jogo, começando por quem vem depois de você. */
export function opponents(v: GameView) {
  if (!v.room) return [];
  const order = playOrder(v.room);
  const i = order.findIndex((p) => p.id === v.me);
  return i < 0 ? order : [...order.slice(i + 1), ...order.slice(0, i)];
}

export function selectSeats(v: GameView): SeatVM[] {
  return opponents(v).map((p) => {
    const out = !p.alive;
    const cards = p.cards.map(miniCard);
    return {
      id: p.id,
      name: p.name,
      short: firstName(p.name),
      coins: p.coins,
      cards,
      inScene: v.room?.currentPlayer === p.id,
      out,
      hasMore:
        cards.some((c) => c.state === "revealed" || c.state === "protected") ||
        p.investments.length > 0 ||
        v.log.some((e) => e.actorId === p.id),
      label: copy.seatLabel(p.name, p.coins, out),
    };
  });
}

export type Declared = { actorId: string; actorName: string; role: RoleId; phase: "declared" | "contested" | "bluff" | "truth"; targetId: string | null };

export function selectStage(v: GameView) {
  const r = v.room;
  const activeId = r?.currentPlayer ?? "";
  const ev = r?.gameEvent;
  const p = payload(v);
  let declared: Declared | null = null;
  const role = declaredRole(ev?.type);
  if (ev && role) {
    declared = { actorId: ev.playerId, actorName: fname(v, ev.playerId), role, phase: "declared", targetId: typeof p.targetPlayer === "string" ? p.targetPlayer : null };
  } else if (ev && (ev.type === "CONTEST" || ev.type === "CONTEST_PENALTY")) {
    const actorId = String(p.contestedPlayer ?? "");
    const contested = declaredRole(`CARD_PLAYED_${String(p.contestedCard ?? "")}`);
    if (contested) {
      declared = {
        actorId,
        actorName: fname(v, actorId),
        role: contested,
        phase: ev.type === "CONTEST" ? "contested" : p.hasCard ? "truth" : "bluff",
        targetId: null,
      };
    }
  }
  return { activeId, activeName: fname(v, activeId), fala: "", declared };
}

export function selectFootlights(v: GameView) {
  const r = v.room;
  const started = !!r?.gameEvent && r.gameEvent.type !== "WAITING_START";
  const effects = Object.entries(r?.doubledCardValues ?? {})
    .filter(([id, d]) => d.timesValueDoubled > 0 && id in ROLES)
    .map(([id, d]) => {
      const name = roleName(id as RoleId);
      return { role: id as RoleId, name, times: d.timesValueDoubled, rounds: d.roundsUntilDecrease, label: copy.effectLabel(name, d.timesValueDoubled, d.roundsUntilDecrease) };
    });
  return { deck: r?.deck ?? 0, tax: r?.tax ?? 0, suddenDeath: started && r?.deck === 0, effects };
}

export type YouCard = { role?: RoleId; protected: boolean; dead: boolean; index: number };

export function selectYou(v: GameView) {
  const r = v.room;
  const me = v.me ? r?.players[v.me] : undefined;
  const ev = r?.gameEvent;
  const total = r?.deck === 0 ? CLASSIC_RULES.waitingAction.suddenDeath : CLASSIC_RULES.waitingAction.default;
  const owner = ev?.playerId ?? r?.currentPlayer ?? "";
  return {
    name: me?.name ?? "",
    coins: me?.coins ?? 0,
    limit: CLASSIC_RULES.general.maxCoins,
    alive: me?.alive ?? false,
    cards: (me?.cards ?? []).map((c): YouCard => ({ role: c.name, protected: c.protected, dead: c.dead, index: c.index })),
    investments: me?.investments ?? [],
    nextInvestment: me?.investments.length ? Math.min(...me.investments) : null,
    timer: ev && ev.type !== "WAITING_START" ? { expiresAt: ev.expiresAt, total, ownerName: fname(v, owner) } : null,
  };
}

export type Block = { role: RoleId; cost: number; payload: "guardian" | "selfish" };
export type Decision =
  | { kind: "none" | "myTurn" | "outOfScene" | "finale" }
  | { kind: "react"; actorId: string; role: RoleId | null; canContest: boolean; block: Block | null; targetsMe: boolean }
  | { kind: "pickCardToKill"; targetId: string; role: RoleId }
  | { kind: "waitPick"; chooserId: string; role: RoleId }
  | { kind: "waitOther"; actorId: string };

export function selectDecision(v: GameView): Decision {
  const r = v.room;
  const me = v.me ? r?.players[v.me] : undefined;
  if (!r) return { kind: "none" };
  if (r.gameOver) return { kind: "finale" };
  if (me && !me.alive) return { kind: "outOfScene" };
  const ev = r.gameEvent;
  if (!ev || !me) return { kind: "none" };
  const p = payload(v);
  const mine = ev.playerId === me.id;
  const role = declaredRole(ev.type);
  const aliveCards = me.cards.filter((c) => !c.dead).length;
  const guardian = (): Block => ({ role: "GUARDIAN", cost: cardValue("GUARDIAN", r, CLASSIC_RULES, aliveCards === 1 ? 1 : 0), payload: "guardian" });

  if (ev.type === "WAITING_FIRST_ACTION") return mine ? { kind: "myTurn" } : { kind: "none" };
  if (role) {
    if (mine) return { kind: "waitOther", actorId: me.id };
    const targetsMe = p.targetPlayer === me.id;
    // Guardiã bloqueia Clarividente (guardianValidation.go: CanDefendClairvoyant)
    const block = role === "CLAIRVOYANT" && targetsMe ? guardian() : null;
    return { kind: "react", actorId: ev.playerId, role, canContest: true, block, targetsMe };
  }
  if (ev.type === "CONTEST_PENALTY") {
    const contested = String(p.contestedPlayer ?? "");
    const card = declaredRole(`CARD_PLAYED_${String(p.contestedCard ?? "")}`) ?? "ASSASSIN";
    // contestPenaltyValidation.go: sem a carta, quem contestou escolhe; com a carta, o contestado escolhe
    const chooser = p.hasCard ? contested : ev.playerId;
    const victim = p.hasCard ? ev.playerId : contested;
    if (chooser === me.id) return { kind: "pickCardToKill", targetId: victim, role: card };
    if (victim === me.id) return { kind: "waitPick", chooserId: chooser, role: card };
    return { kind: "waitOther", actorId: chooser };
  }
  if (!mine && ev.type === "CARD_KILLED" && p.targetPlayer === me.id && p.cause !== "CONTEST_PENALTY" && p.cause !== "KAMIKAZE") {
    return { kind: "react", actorId: ev.playerId, role: null, canContest: false, block: guardian(), targetsMe: true };
  }
  if (!mine && ["EARN_COINS", "STEAL_COINS", "INVESTMENT", "CHANGED_CARD"].includes(ev.type)) {
    const variant = ev.type === "INVESTMENT" ? 5 : 0;
    const cost = ev.type === "CHANGED_CARD" || ev.type === "INVESTMENT" ? cardValue("SELFISH", r, CLASSIC_RULES, variant) : CLASSIC_RULES.cards.SELFISH.fixedValue ?? 1;
    return { kind: "react", actorId: ev.playerId, role: null, canContest: false, block: { role: "SELFISH", cost, payload: "selfish" }, targetsMe: false };
  }
  return { kind: "none" };
}

export function selectLog(v: GameView, playerId?: string) {
  return v.log
    .filter((e) => !playerId || e.actorId === playerId)
    .map((e) => ({ key: e.key, turn: e.turn, text: playerId ? logSelfText(e) : logText(e) }))
    .reverse();
}

export function selectRolesUsed(v: GameView, playerId: string) {
  return v.log
    .filter((e) => e.actorId === playerId && e.role)
    .map((e) => ({ key: e.key, role: e.role!, name: roleName(e.role), turn: e.turn, text: logSelfText(e) }))
    .reverse();
}

export type ChatItem =
  | { kind: "line"; key: string; name: string; text: string; time: string; mine: boolean }
  | { kind: "event"; key: string; text: string };

const hhmm = (iso: string) => {
  const d = new Date(iso);
  return `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`;
};

export function selectChat(v: GameView): ChatItem[] {
  const myName = v.me ? v.room?.players[v.me]?.name : undefined;
  const lines = (v.room?.chat ?? []).map((m, i) => ({
    at: Date.parse(m.timestamp),
    item: { kind: "line" as const, key: `c${i}|${m.timestamp}`, name: m.playerName === myName ? copy.you : m.playerName, text: m.message, time: hhmm(m.timestamp), mine: m.playerName === myName },
  }));
  const events = v.log.filter((e) => e.chat).map((e) => ({ at: e.at, item: { kind: "event" as const, key: e.key, text: `(${logText(e)})` } }));
  return [...lines, ...events]
    .sort((a, b) => a.at - b.at)
    .slice(-30)
    .map((x) => x.item);
}

export function selectFinale(v: GameView) {
  const r = v.room;
  const winner = r ? Object.values(r.players).find((p) => p.alive) : undefined;
  const exits = exitsOf(v.log)
    .reverse()
    .map((e, i) => ({ place: i + 2, name: e.playerId === v.me ? copy.you : firstName(e.name), round: e.turn, me: e.playerId === v.me }));
  return { winnerName: winner ? firstName(winner.name) : "", round: r?.turn ?? 0, coins: winner?.coins ?? 0, exits, fala: "" };
}

export function selectWaitingRoom(v: GameView) {
  const r = v.room;
  const order = r ? playOrder(r) : [];
  // ponytail: o servidor ainda não manda maxPlayers; sem ele, vale o padrão de ClassicRules.yaml
  const max = Math.max(r?.maxPlayers ?? CLASSIC_RULES.general.maxPlayers, order.length);
  const hostId = order[0]?.id;
  const seats = Array.from({ length: max }, (_, i) => {
    const p = order[i];
    if (!p) return { key: `vago${i}`, name: copy.vacantRole, tag: copy.waiting, vacant: true };
    const tag = p.id === hostId ? (p.id === v.me ? copy.hostYou : copy.host) : p.id === v.me ? copy.you.toLowerCase() : copy.ready;
    return { key: p.id, name: p.name, tag, vacant: false };
  });
  const tight = max > 6;
  return {
    code: r?.id ?? "",
    name: r?.name || (r ? `${copy.room} ${r.id}` : ""),
    max,
    count: order.length,
    seats,
    castLine: copy.castLine(order.length, max),
    tight,
    isHost: !!v.me && hostId === v.me,
  };
}
