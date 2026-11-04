/**
 * Diário da cena, montado no cliente (o servidor não guarda passado). 05-protocolo.md → "Diário da cena".
 *
 * appendLog(log, prev, next, now) compara duas salas consecutivas e devolve o diário novo:
 * - um `gameEvent` novo (chave tipo|jogador|expiresAt) vira uma linha, ou completa a linha da declaração
 *   ("se disse Guardiã." + " Ninguém contestou." / ", foi contestado e perdeu a carta.");
 * - um ator que passou de vivo a morto ganha "saiu de cena.";
 * - a mesma sala aplicada de novo (reconexão) não muda nada.
 *
 * Cada entrada guarda o ator e o predicado: o diário geral mostra `${firstName(actorName)} ${pred}`,
 * a gaveta do camarote mostra `capitalize(pred)`.
 */
import { capitalize, copy, firstName } from "./copy";
import type { GameEvent, PublicRoom, RoleId } from "./protocol";
import { declaredRole, roleName } from "./roles";

export type LogEntry = {
  key: string;
  turn: number;
  at: number;
  actorId: string;
  actorName: string;
  pred: string;
  /** a linha é de uma declaração ainda aberta (pode ganhar o desfecho) */
  open?: "declared" | "contested";
  /** papel declarado nesta linha, para "Papéis que a mesa já viu você usar" */
  role?: RoleId;
};

export const eventKey = (e: GameEvent | null) => (e ? `${e.type}|${e.playerId}|${e.expiresAt}` : "");
export const logText = (e: LogEntry) => (e.actorId ? `${firstName(e.actorName)} ${e.pred}` : capitalize(e.pred));
export const logSelfText = (e: LogEntry) => capitalize(e.pred);

const obj = (p: unknown): Record<string, unknown> => (p && typeof p === "object" ? (p as Record<string, unknown>) : {});
const str = (v: unknown) => (typeof v === "string" ? v : "");
const num = (v: unknown) => (typeof v === "number" ? v : 0);

export function appendLog(log: LogEntry[], prev: PublicRoom | null, next: PublicRoom, now: number): LogEntry[] {
  const out = [...log];
  const ev = next.gameEvent;
  const key = eventKey(ev);
  const nameOf = (id: string) => next.players[id]?.name ?? prev?.players[id]?.name ?? "";
  const fname = (id: string) => firstName(nameOf(id));

  if (ev && key !== eventKey(prev?.gameEvent ?? null) && !out.some((e) => e.key === key)) {
    const last = out.length - 1;
    const open = last >= 0 ? out[last] : undefined;
    const p = obj(ev.payload);
    const push = (actorId: string, pred: string, extra: Partial<LogEntry> = {}) =>
      out.push({ key, turn: next.turn, at: now, actorId, actorName: nameOf(actorId), pred, ...extra });
    const amend = (pred: string, state?: LogEntry["open"]) => {
      if (open) out[last] = { ...open, pred, open: state };
    };

    // a declaração anterior não foi contestada: fecha a linha dela
    if (open?.open === "declared" && ev.type !== "CONTEST") amend(open.pred + copy.log.noContest);

    const role = declaredRole(ev.type);
    if (role) {
      push(ev.playerId, copy.log.declared(roleName(role)), { open: "declared", role });
    } else {
      switch (ev.type) {
        case "CONTEST":
          if (open?.open === "declared" && open.role) amend(copy.log.contested(roleName(open.role)), "contested");
          else push(ev.playerId, copy.log.contestedBy(fname(str(p.contestedPlayer))));
          break;
        case "CONTEST_PENALTY":
          if (open?.open === "contested") amend(open.pred + (p.hasCard ? copy.log.showedCard : copy.log.lostCard));
          break;
        case "CARD_KILLED":
          if (str(p.cause) !== "CONTEST") push(str(p.targetPlayer) || ev.playerId, copy.log.lostACard);
          break;
        case "REVEALED_CARD":
          push(ev.playerId, copy.log.revealed(fname(str(p.targetPlayer))));
          break;
        case "REVIVED_CARD":
          push(ev.playerId, copy.log.revived);
          break;
        case "CHANGED_CARD":
          push(ev.playerId, copy.log.changed);
          break;
        case "EARN_COINS":
          push(ev.playerId, copy.log.earned(num(p.earnedCoins), roleName(declaredRole(`CARD_PLAYED_${str(p.cause)}`) ?? "TRILLIONAIRE")));
          break;
        case "STEAL_COINS":
          push(ev.playerId, copy.log.stole(num(p.lostCoins), fname(str(p.targetPlayer))));
          break;
        case "INVESTMENT":
          push(ev.playerId, copy.log.invested);
          break;
        case "INCREMENT_TAXES":
          push(ev.playerId, copy.log.taxUp);
          break;
        case "DECREMENT_TAXES":
          push(ev.playerId, copy.log.taxDown);
          break;
        case "RICHNESS_DENIED":
          push(ev.playerId, copy.log.denied(fname(str(p.targetPlayer))));
          break;
        case "SUDDEN_DEATH":
          push("", copy.log.suddenDeath);
          break;
      }
    }
  }

  // quem saiu de cena entre as duas salas
  if (prev) {
    for (const p of Object.values(next.players)) {
      const exitKey = `EXIT|${p.id}`;
      if (!p.alive && prev.players[p.id]?.alive && !out.some((e) => e.key === exitKey)) {
        out.push({ key: exitKey, turn: next.turn, at: now, actorId: p.id, actorName: p.name, pred: copy.log.exited });
      }
    }
  }
  return out;
}

/** Saídas de cena, da primeira à última, a partir do diário (para a ordem de saída no Fim). */
export const exitsOf = (log: LogEntry[]) =>
  log.filter((e) => e.key.startsWith("EXIT|")).map((e) => ({ playerId: e.actorId, name: e.actorName, turn: e.turn }));
