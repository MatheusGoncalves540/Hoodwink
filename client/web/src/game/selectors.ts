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
import { exitsOf, logSelfText, logText, revealedCards, type LogEntry } from "./log";
import { playOrder, type PublicCard, type PublicRoom, type RoleId } from "./protocol";
import { cardValue, guardianPrice } from "./pricing";
import { declaredRole, roleName, ROLES } from "./roles";
import { CLASSIC_RULES } from "./rules.classic";

export type GameView = { me: string | null; room: PublicRoom | null; log: LogEntry[] };

export type MiniCard = { state: "back" | "revealed" | "protected" | "dead"; role?: RoleId; index?: number };
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

/** A mesa viu a carta: o estado atual manda; se ela já voltou a ficar escondida, o diário lembra o papel. */
const cardWithReveal = (c: PublicCard, known: Map<number, RoleId> | undefined): MiniCard => {
  const base = miniCard(c);
  const role = base.state === "back" ? known?.get(c.index) : undefined;
  return role ? { state: "revealed", role } : base;
};

const fname = (v: GameView, id: string) => firstName(v.room?.players[id]?.name ?? "");
const payload = (v: GameView) => (v.room?.gameEvent?.payload ?? {}) as Record<string, unknown>;
/** startTime zero (Go: "0001-01-01T00:00:00Z") = sala ainda não começou; gameEvent nulo só marca o intervalo entre deixas. */
const hasStarted = (r: PublicRoom | null) => !!r && Date.parse(r.startTime) > 0;

export function selectScene(v: GameView): "lobby" | "room" | "table" | "finale" {
  if (!v.room) return "lobby";
  if (v.room.gameOver) return "finale";
  return hasStarted(v.room) ? "table" : "room";
}

/** Adversários na ordem de jogo, começando por quem vem depois de você. */
export function opponents(v: GameView) {
  if (!v.room) return [];
  const order = playOrder(v.room);
  const i = order.findIndex((p) => p.id === v.me);
  return i < 0 ? order : [...order.slice(i + 1), ...order.slice(0, i)];
}

export function selectSeats(v: GameView): SeatVM[] {
  const known = revealedCards(v.log);
  return opponents(v).map((p) => {
    const out = !p.alive;
    const cards = p.cards.map((c): MiniCard => ({ ...cardWithReveal(c, known.get(p.id)), index: c.index }));
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

export type Declared = {
  actorId: string;
  actorName: string;
  role: RoleId;
  phase: "declared" | "contested" | "bluff" | "truth";
  targetId: string | null;
  targetName: string | null;
  /** quem duvidou (CONTEST e CONTEST_PENALTY) */
  contesterId: string | null;
  contesterName: string | null;
  /** quem escolhe a carta que cai (CONTEST_PENALTY; null enquanto ninguém escolheu) */
  chooserName: string | null;
};

/** Carta revelada pelo Clarividente. `card` só existe para quem pode ver (todos, ou só quem jogou). */
export type Revealed = { actorId: string; actorName: string; targetId: string; targetName: string; card: RoleId | null; mine: boolean; toAll: boolean };

const roleOrNull = (raw: unknown): RoleId | null => (typeof raw === "string" && raw in ROLES ? (raw as RoleId) : null);

export function selectStage(v: GameView) {
  const r = v.room;
  const activeId = r?.currentPlayer ?? "";
  const ev = r?.gameEvent;
  const p = payload(v);
  let declared: Declared | null = null;
  let revealed: Revealed | null = null;
  const role = declaredRole(ev?.type);
  if (ev && role) {
    const targetId = typeof p.targetPlayer === "string" ? p.targetPlayer : null;
    declared = {
      actorId: ev.playerId,
      actorName: fname(v, ev.playerId),
      role,
      phase: "declared",
      targetId,
      targetName: targetId ? fname(v, targetId) : null,
      contesterId: null,
      contesterName: null,
      chooserName: null,
    };
  } else if (ev && ev.type === "CONTEST") {
    const contestedId = String(p.contestedPlayer ?? "");
    const contested = declaredRole(`CARD_PLAYED_${String(p.contestedCard ?? "")}`);
    if (contested) {
      declared = {
        actorId: contestedId,
        actorName: fname(v, contestedId),
        role: contested,
        phase: "contested",
        targetId: null,
        targetName: null,
        contesterId: ev.playerId,
        contesterName: fname(v, ev.playerId),
        chooserName: null,
      };
    }
  } else if (ev && ev.type === "CONTEST_PENALTY") {
    // contestPenaltyValidation.go: sem a carta, quem duvidou escolhe; com a carta, o contestado escolhe
    const contestedId = String(p.contestedPlayer ?? "");
    const contesterId = String(p.contestingPlayer ?? "") || ev.playerId;
    const contested = declaredRole(`CARD_PLAYED_${String(p.contestedCard ?? "")}`);
    if (contested) {
      declared = {
        actorId: contestedId,
        actorName: fname(v, contestedId),
        role: contested,
        phase: p.hasCard ? "truth" : "bluff",
        targetId: null,
        targetName: null,
        contesterId,
        contesterName: fname(v, contesterId),
        chooserName: fname(v, p.hasCard ? contestedId : contesterId),
      };
    }
  } else if (ev && ev.type === "REVEALED_CARD") {
    const targetId = String(p.targetPlayer ?? "");
    revealed = {
      actorId: ev.playerId,
      actorName: fname(v, ev.playerId),
      targetId,
      targetName: fname(v, targetId),
      card: roleOrNull(p.revealedCard),
      mine: !!v.me && ev.playerId === v.me,
      toAll: p.showToAllPlayers === true,
    };
  }
  return { activeId, activeName: fname(v, activeId), fala: declared || revealed ? "" : describeEvent(v), declared, revealed };
}

/**
 * Uma frase para o acontecimento atual, com os dois lados (quem fez e quem sofreu).
 * Alimenta a fala do palco e o texto do relógio: a cada evento o texto muda, mesmo quando a janela reinicia.
 */
export function describeEvent(v: GameView): string {
  const ev = v.room?.gameEvent;
  if (!ev || ev.type === "WAITING_START" || ev.type === "ACTION_TAKEN") return "";
  const p = payload(v);
  const id = (x: unknown) => String(x ?? "");
  /** Nome para a frase: "Você" quando é o jogador desta tela. */
  const name = (x: unknown) => (v.me && id(x) === v.me ? copy.you : fname(v, id(x)));
  const cardPhrase = (x: unknown) => copy.targetText({ name: fname(v, id(x)), mine: !!v.me && id(x) === v.me });
  const poss = (x: unknown) => copy.possessive(fname(v, id(x)), !!v.me && id(x) === v.me);
  const role = declaredRole(ev.type);
  if (role) return copy.stageLine.declared(name(ev.playerId), roleName(role));

  const card = declaredRole(`CARD_PLAYED_${String(p.contestedCard ?? "")}`);
  const contested = name(p.contestedPlayer);
  const contester = name(id(p.contestingPlayer) || ev.playerId);
  switch (ev.type) {
    case "WAITING_FIRST_ACTION":
      return v.me && id(ev.playerId) === v.me ? copy.stageLine.yourTurn : copy.stageLine.turn(name(ev.playerId));
    case "CONTEST":
      return copy.stageLine.doubted(name(ev.playerId), contested);
    case "CONTEST_PENALTY": {
      const shown = card ? roleName(card) : "a carta";
      const chooser = p.hasCard ? contested : contester;
      const victim = p.hasCard ? p.contestingPlayer : p.contestedPlayer;
      return p.hasCard
        ? `${copy.hadCard(contested, shown)} ${copy.stageLine.chooseCardOf(chooser, poss(victim))}`
        : `${copy.bluffed(contested, shown)} ${copy.stageLine.chooseCardOf(chooser, poss(victim))}`;
    }
    case "CARD_KILLED": {
      const target = cardPhrase(p.targetPlayer);
      switch (String(p.cause ?? "")) {
        case "KILL_CARD": return copy.stageLine.killed(name(ev.playerId), target);
        case "ASSASSIN": return copy.stageLine.killed(name(ev.playerId), target);
        case "KAMIKAZE": return copy.stageLine.kamikaze(name(ev.playerId), target);
        case "GREED": return copy.stageLine.greed(name(p.targetPlayer) || name(ev.playerId));
        default: return copy.stageLine.lostCard(name(p.targetPlayer) || name(ev.playerId));
      }
    }
    case "REVEALED_CARD": {
      const target = poss(p.targetPlayer);
      const revealedCard = roleOrNull(p.revealedCard);
      if (revealedCard && p.showToAllPlayers === true) return copy.stageLine.revealedCard(roleName(revealedCard), target);
      if (revealedCard) return v.me === ev.playerId ? copy.stageLine.revealedPrivateMe(roleName(revealedCard), target) : copy.stageLine.revealed(name(ev.playerId), target);
      return copy.stageLine.revealed(name(ev.playerId), target);
    }
    case "REVIVED_CARD": return copy.stageLine.revived(name(ev.playerId));
    case "CHANGED_CARD": return copy.stageLine.changed(name(ev.playerId));
    case "EARN_COINS":
      return String(p.cause ?? "") === "TAKE_COINS" ? copy.stageLine.took(name(ev.playerId), Number(p.earnedCoins ?? 0)) : copy.stageLine.earned(name(ev.playerId), Number(p.earnedCoins ?? 0));
    case "STEAL_COINS": return copy.stageLine.stole(name(ev.playerId), Number(p.lostCoins ?? 0), name(p.targetPlayer));
    case "INVESTMENT": return copy.stageLine.invested(name(ev.playerId));
    case "INCREMENT_TAXES": return copy.stageLine.taxUp(name(ev.playerId));
    case "DECREMENT_TAXES": return copy.stageLine.taxDown(name(ev.playerId));
    case "RICHNESS_DENIED": return copy.stageLine.denied(name(ev.playerId), name(p.targetPlayer));
    case "SUDDEN_DEATH": return copy.stageLine.suddenDeath;
    default: return "";
  }
}

export function selectFootlights(v: GameView) {
  const r = v.room;
  const started = hasStarted(r);
  const effects = Object.entries(r?.doubledCardValues ?? {})
    .filter(([id, d]) => d.timesValueDoubled > 0 && id in ROLES)
    .map(([id, d]) => {
      const name = roleName(id as RoleId);
      return { role: id as RoleId, name, times: d.timesValueDoubled, rounds: d.roundsUntilDecrease, label: copy.effectLabel(name, d.timesValueDoubled, d.roundsUntilDecrease) };
    });
  return { deck: r?.deck ?? 0, tax: r?.tax ?? 0, suddenDeath: started && r?.deck === 0, effects };
}

export type YouCard = { role?: RoleId; protected: boolean; dead: boolean; index: number };

/** Janela de cada deixa (ClassicRules.yaml): ação 10 s, mensagem 6 s, importante 8 s. */
function eventWindow(type: string | undefined, suddenDeath: boolean): number {
  const w = suddenDeath ? "suddenDeath" : "default";
  if (type === "WAITING_FIRST_ACTION" || (type?.startsWith("CARD_PLAYED_") && type !== "CARD_PLAYED_GUARDIAN")) return CLASSIC_RULES.waitingAction[w];
  if (type === "CARD_PLAYED_GUARDIAN") return CLASSIC_RULES.displayImportantMessage[w];
  return CLASSIC_RULES.displayMessage[w];
}

export function selectYou(v: GameView) {
  const r = v.room;
  const me = v.me ? r?.players[v.me] : undefined;
  const ev = r?.gameEvent;
  const total = eventWindow(ev?.type, r?.deck === 0);
  const owner = ev?.playerId ?? r?.currentPlayer ?? "";
  return {
    name: me?.name ?? "",
    coins: me?.coins ?? 0,
    limit: CLASSIC_RULES.general.maxCoins,
    alive: me?.alive ?? false,
    cards: (me?.cards ?? []).map((c): YouCard => ({ role: c.name, protected: c.protected, dead: c.dead, index: c.index })),
    investments: me?.investments ?? [],
    nextInvestment: me?.investments.length ? Math.min(...me.investments) : null,
    timer:
      ev && ev.type !== "WAITING_START"
        ? { expiresAt: ev.expiresAt, total, ownerName: fname(v, owner), caption: describeEvent(v) }
        : null,
  };
}

export type Block = { role: RoleId; cost: number; payload: "guardian" | "selfish" };
export type Decision =
  | { kind: "none" | "myTurn" | "outOfScene" | "finale" }
  | { kind: "react"; actorId: string; role: RoleId | null; canContest: boolean; block: Block | null; targetsMe: boolean; kamikaze?: true }
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
  // Guardiã: reação a ataque 5 + t (mín 4); contra Clarividente 4 + t (mín 2). 02 §3.
  const guardian = (when: "reaction" | "clairvoyant"): Block => ({ role: "GUARDIAN", cost: guardianPrice(when, r), payload: "guardian" });

  if (ev.type === "WAITING_FIRST_ACTION") return mine ? { kind: "myTurn" } : { kind: "none" };
  if (role) {
    if (mine) return { kind: "waitOther", actorId: me.id };
    const targetsMe = p.targetPlayer === me.id;
    // Guardiã bloqueia Clarividente (guardianValidation.go: CanDefendClairvoyant)
    const block = role === "CLAIRVOYANT" && targetsMe ? guardian("clairvoyant") : null;
    return { kind: "react", actorId: ev.playerId, role, canContest: true, block, targetsMe };
  }
  if (ev.type === "CONTEST_PENALTY") {
    const contested = String(p.contestedPlayer ?? "");
    const contester = String(p.contestingPlayer ?? "") || ev.playerId;
    const card = declaredRole(`CARD_PLAYED_${String(p.contestedCard ?? "")}`) ?? "ASSASSIN";
    // contestPenaltyValidation.go: sem a carta, quem duvidou escolhe; com a carta, o contestado escolhe
    const chooser = p.hasCard ? contested : contester;
    const victim = p.hasCard ? contester : contested;
    if (chooser === me.id) return { kind: "pickCardToKill", targetId: victim, role: card };
    if (victim === me.id) return { kind: "waitPick", chooserId: chooser, role: card };
    return { kind: "waitOther", actorId: chooser };
  }
  if (ev.type === "CARD_KILLED" && p.targetPlayer === me.id) {
    // Guardiã na hora não segura contestação, Kamikaze nem ganância; carta já protegida não precisa (guardianValidation.go)
    const card = me.cards.find((c) => c.index === Number(p.targetCardIndex));
    const blockable = !["CONTEST_PENALTY", "KAMIKAZE", "GREED"].includes(String(p.cause)) && !card?.protected;
    // Kamikaze revida quando é ele que está morrendo e a morte vai acontecer de fato (kamikazeValidation.go)
    const dies = !card?.protected || p.cause === "CONTEST_PENALTY" || p.cause === "GREED";
    const kamikaze = card?.name === "KAMIKAZE" && !card.dead && dies;
    if (blockable || kamikaze) {
      const block = blockable ? guardian("reaction") : null;
      return { kind: "react", actorId: ev.playerId, role: null, canContest: false, block, targetsMe: true, ...(kamikaze && { kamikaze: true as const }) };
    }
    return { kind: "none" };
  }
  // ações do turno (Pegar moeda, Trocar cartas) ninguém impede (02 §2)
  const basic = p.cause === "TAKE_COINS" || p.cause === "SWAP_CARDS";
  if (!mine && !basic && ["EARN_COINS", "STEAL_COINS", "INVESTMENT", "CHANGED_CARD"].includes(ev.type)) {
    // 02 §3: Egoísta barra o Trapaceiro só se o ganho (⌈tirado/2⌉) passar de 1, e moedas só a partir de 2 + t
    if (ev.type === "STEAL_COINS" && Math.ceil(Number(p.lostCoins ?? 0) / 2) <= 1) return { kind: "none" };
    if (ev.type === "EARN_COINS" && p.earnedCoins != null && Number(p.earnedCoins) < CLASSIC_RULES.selfishThreshold + r.tax) return { kind: "none" };
    const variant = ev.type === "INVESTMENT" ? 5 : 0;
    const cost = ev.type === "CHANGED_CARD" || ev.type === "INVESTMENT" ? cardValue("SELFISH", r, CLASSIC_RULES, variant) : CLASSIC_RULES.cards.SELFISH.fixedValue ?? 1;
    return { kind: "react", actorId: ev.playerId, role: null, canContest: false, block: { role: "SELFISH", cost, payload: "selfish" }, targetsMe: false };
  }
  return { kind: "none" };
}

/**
 * Qual Kamikaze seu morre (targetAllyCardIndex): na reação, o que está sendo morto;
 * no seu turno, o primeiro vivo. null = você não tem Kamikaze vivo (o servidor recusaria).
 */
export function kamikazeAlly(v: GameView): number | null {
  const me = v.me ? v.room?.players[v.me] : undefined;
  if (!me) return null;
  const ev = v.room?.gameEvent;
  const p = payload(v);
  const live = (c: PublicCard) => c.name === "KAMIKAZE" && !c.dead;
  const dying = ev?.type === "CARD_KILLED" && p.targetPlayer === me.id ? me.cards.find((c) => c.index === Number(p.targetCardIndex)) : undefined;
  if (dying) return live(dying) ? dying.index : null;
  return me.cards.find(live)?.index ?? null;
}

export function selectLog(v: GameView, playerId?: string) {
  return v.log
    // a revelação pública entra no diário do alvo pela linha derivada de selectPublicReveals
    .filter((e) => !playerId || e.actorId === playerId || (e.reveal?.targetId === playerId && e.reveal.showToAll !== true))
    .map((e) => ({ key: e.key, turn: e.turn, text: playerId && e.actorId === playerId ? logSelfText(e) : logText(e) }))
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
  const winnerIsMe = !!winner && winner.id === v.me;
  const exits = exitsOf(v.log)
    .reverse()
    .map((e, i) => ({ place: i + 2, name: e.playerId === v.me ? copy.you : firstName(e.name), round: e.turn, me: e.playerId === v.me }));
  return { winnerName: winner ? (winnerIsMe ? copy.you : firstName(winner.name)) : "", winnerIsMe, round: r?.turn ?? 0, coins: winner?.coins ?? 0, exits, fala: "" };
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

export type SeatSheetVM = {
  id: string;
  name: string;
  coins: number;
  alive: boolean;
  cards: (MiniCard & { index: number })[];
  investments: number[];
  /** cartas que a mesa viu deste jogador (revelação pública do Clarividente) */
  revealed: RevealedCardVM[];
  log: { key: string; turn: number; text: string }[];
};

export type RevealedCardVM = { key: string; turn: number; cardIndex: number; card?: RoleId };

/**
 * Memória do diário para a gaveta do alvo: só a revelação com "mostrar para todos"
 * conta, e duas revelações da mesma carta ficam com o turno mais recente.
 */
export function selectPublicReveals(v: GameView, playerId: string): RevealedCardVM[] {
  const byCard = new Map<number, RevealedCardVM>();
  for (const e of v.log) {
    // troca de carta apaga o que a mesa já tinha visto daquele jogador
    if (e.forgetCards === playerId) byCard.clear();
    const r = e.reveal;
    if (!r || !r.showToAll || r.targetId !== playerId) continue;
    const prev = byCard.get(r.cardIndex);
    if (!prev || e.turn >= prev.turn) byCard.set(r.cardIndex, { key: e.key, turn: e.turn, cardIndex: r.cardIndex, card: r.card });
  }
  return [...byCard.values()].sort((a, b) => b.turn - a.turn);
}

/** Gaveta de um camarote (telas 4 e 14): cartas no estado que a mesa vê, diário do ator e investimentos. */
export function selectSeatSheet(v: GameView, id: string): SeatSheetVM | null {
  const p = v.room?.players[id];
  if (!p) return null;
  // A fileira da gaveta só marca a carta vista em revelação privada; a pública
  // vive na memória do diário (bloco "Carta revelada no turno N"), sem abrir a carta.
  const known = revealedCards(v.log.filter((e) => e.reveal?.showToAll !== true)).get(p.id);
  const revealed = selectPublicReveals(v, id);
  const log = [
    ...selectLog(v, id),
    // linha derivada: a revelação pública é memória do diário do alvo, não estado da carta
    ...revealed.map((r) => ({ key: `REVEAL|${r.key}`, turn: r.turn, text: copy.revealLog(r.card ? roleName(r.card) : undefined) })),
  ].sort((a, b) => b.turn - a.turn);
  return {
    id: p.id,
    name: p.name,
    coins: p.coins,
    alive: p.alive,
    cards: p.cards.map((c) => ({ ...cardWithReveal(c, known), index: c.index })),
    investments: p.investments,
    revealed,
    log,
  };
}

export type YouSheetVM = {
  name: string;
  coins: number;
  limit: number;
  cards: YouCard[];
  investments: number[];
  roles: ReturnType<typeof selectRolesUsed>;
};

/** Seu camarote (telas 12 e 16): suas cartas, o limite, investimentos e os papéis que a mesa já viu. */
export function selectYouSheet(v: GameView): YouSheetVM | null {
  const p = v.me ? v.room?.players[v.me] : undefined;
  if (!p) return null;
  return {
    name: copy.you,
    coins: p.coins,
    limit: CLASSIC_RULES.general.maxCoins,
    cards: p.cards.map((c): YouCard => ({ role: c.name, protected: c.protected, dead: c.dead, index: c.index })),
    investments: p.investments,
    roles: selectRolesUsed(v, p.id),
  };
}
