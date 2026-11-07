/**
 * Rascunhos dos relatos marcados na partida (plano 06, T4).
 * O flag só vira rascunho depois que o backend responde 200/201; o cartão do fim
 * (ou do Lobby, na próxima abertura) lê daqui para pedir os detalhes.
 * Tudo é tolerante a localStorage bloqueado e a JSON corrompido: nunca lança.
 */
import { readPref, setPref } from "../../app/prefs";

export type BugDraft = {
  reportId: string;
  matchId: string;
  turns: number[];
  at: number;
  /** Quem marcou. O rascunho é ignorado se outro usuário estiver logado. */
  userId?: string;
};

export const BUG_DRAFTS_KEY = "hoodwink-bug-drafts";

const MAX_TURNS = 10;
const SEVEN_DAYS = 7 * 24 * 60 * 60 * 1000;
const MAX_TURN = 10_000;

const isTurn = (n: unknown): n is number => typeof n === "number" && Number.isInteger(n) && n >= 0 && n <= MAX_TURN;

const isDraft = (value: unknown): value is BugDraft => {
  if (!value || typeof value !== "object") return false;
  const d = value as Partial<BugDraft>;
  return (
    typeof d.reportId === "string" &&
    typeof d.matchId === "string" &&
    typeof d.at === "number" &&
    Number.isFinite(d.at) &&
    Array.isArray(d.turns) &&
    d.turns.every(isTurn) &&
    (d.userId === undefined || typeof d.userId === "string")
  );
};

const uniqueTurns = (turns: number[]) => [...new Set(turns.filter(isTurn))].sort((a, b) => a - b).slice(0, MAX_TURNS);

const readAll = (): BugDraft[] => {
  const raw = readPref(BUG_DRAFTS_KEY);
  if (!raw) return [];
  try {
    const parsed: unknown = JSON.parse(raw);
    return Array.isArray(parsed)
      ? parsed.filter(isDraft).map((d) => ({ ...d, turns: uniqueTurns(d.turns) }))
      : [];
  } catch {
    return [];
  }
};

const writeAll = (drafts: BugDraft[]) => setPref(BUG_DRAFTS_KEY, drafts.length ? JSON.stringify(drafts) : null);

const normalize = (draft: BugDraft): BugDraft => {
  const normalized: BugDraft = {
    reportId: draft.reportId,
    matchId: draft.matchId,
    turns: uniqueTurns(draft.turns),
    at: draft.at,
  };
  if (draft.userId !== undefined) normalized.userId = draft.userId;
  return normalized;
};

/**
 * Grava um flag. Flags do mesmo jogo e do mesmo usuário viram um rascunho só:
 * turnos únicos, no máximo 10, e o `at` mais recente.
 */
export function addFlag(draft: BugDraft): BugDraft {
  const next = normalize(draft);
  const drafts = readAll();
  const index = drafts.findIndex((d) => d.matchId === next.matchId && d.userId === next.userId);

  if (index < 0) {
    drafts.push(next);
    writeAll(drafts);
    return next;
  }

  const current = drafts[index]!;
  const merged: BugDraft = {
    reportId: next.reportId,
    matchId: next.matchId,
    turns: uniqueTurns([...current.turns, ...next.turns]),
    at: Math.max(current.at, next.at),
  };
  if (next.userId !== undefined) merged.userId = next.userId;
  drafts[index] = merged;
  writeAll(drafts);
  return merged;
}

/** O rascunho daquela partida e daquele usuário (o mais recente, se houver repetido). */
export function draftFor(matchId: string, userId?: string): BugDraft | undefined {
  return readAll()
    .filter((d) => d.matchId === matchId && d.userId === userId)
    .sort((a, b) => b.at - a.at)[0];
}

/** Rascunhos do usuário atual com menos de 7 dias, do mais novo para o mais antigo. */
export function pendingOnOpen(now: number, userId?: string): BugDraft[] {
  return readAll()
    .filter((d) => d.userId === userId && now - d.at < SEVEN_DAYS)
    .sort((a, b) => b.at - a.at);
}

/** Apaga o rascunho pelo id do relato (enviar ou pular no cartão). */
export function clear(reportId: string): void {
  const drafts = readAll();
  const kept = drafts.filter((d) => d.reportId !== reportId);
  if (kept.length !== drafts.length) writeAll(kept);
}
