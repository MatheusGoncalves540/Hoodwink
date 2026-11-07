import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { BUG_DRAFTS_KEY, addFlag, clear, draftFor, pendingOnOpen, type BugDraft } from "./drafts";

const DAY = 24 * 60 * 60 * 1000;
const NOW = Date.parse("2026-11-06T12:00:00Z");

const memory = new Map<string, string>();
const storage = {
  getItem: (key: string) => memory.get(key) ?? null,
  setItem: (key: string, value: string) => void memory.set(key, value),
  removeItem: (key: string) => void memory.delete(key),
};

const flag = (over: Partial<BugDraft> = {}): BugDraft => ({
  reportId: "relato-1",
  matchId: "partida-1",
  turns: [3],
  at: NOW,
  userId: "usuario-1",
  ...over,
});

beforeEach(() => {
  memory.clear();
  vi.stubGlobal("localStorage", storage);
});

afterEach(() => vi.unstubAllGlobals());

describe("rascunhos de bug", () => {
  it("junta dois flags da mesma partida em um rascunho com os dois turnos", () => {
    addFlag(flag({ turns: [3], at: NOW }));
    addFlag(flag({ turns: [5], at: NOW + 1000 }));

    expect(pendingOnOpen(NOW + 2000, "usuario-1")).toEqual([
      { reportId: "relato-1", matchId: "partida-1", turns: [3, 5], at: NOW + 1000, userId: "usuario-1" },
    ]);
  });

  it("não duplica o mesmo turno", () => {
    addFlag(flag({ turns: [3] }));
    const draft = addFlag(flag({ turns: [3] }));

    expect(draft.turns).toEqual([3]);
  });

  it("mantém o at mais recente quando o flag antigo chega depois", () => {
    addFlag(flag({ turns: [3], at: NOW }));
    const draft = addFlag(flag({ turns: [5], at: NOW - 5000 }));

    expect(draft.at).toBe(NOW);
  });

  it("guarda no máximo 10 turnos", () => {
    addFlag(flag({ turns: Array.from({ length: 10 }, (_, i) => i + 1) }));
    const draft = addFlag(flag({ turns: [11] }));

    expect(draft.turns).toHaveLength(10);
    expect(draft.turns).not.toContain(11);
  });

  it("separa partidas e usuários diferentes", () => {
    addFlag(flag());
    addFlag(flag({ matchId: "partida-2" }));
    addFlag(flag({ userId: "usuario-2" }));

    expect(pendingOnOpen(NOW, "usuario-1")).toHaveLength(2);
    expect(pendingOnOpen(NOW, "usuario-2")).toHaveLength(1);
  });

  it("draftFor acha só o rascunho da partida e do usuário", () => {
    addFlag(flag());
    addFlag(flag({ matchId: "partida-2" }));
    addFlag(flag({ userId: "usuario-2" }));

    expect(draftFor("partida-2", "usuario-1")?.matchId).toBe("partida-2");
    expect(draftFor("partida-1", "usuario-2")?.userId).toBe("usuario-2");
    expect(draftFor("partida-1", "usuario-3")).toBeUndefined();
  });

  it("ignora rascunho com 8 dias e aceita com 6", () => {
    addFlag(flag({ at: NOW - 8 * DAY }));

    expect(pendingOnOpen(NOW, "usuario-1")).toEqual([]);
    expect(pendingOnOpen(NOW - 2 * DAY, "usuario-1")).toHaveLength(1);
  });

  it("ignora rascunho de outro usuário", () => {
    addFlag(flag());

    expect(pendingOnOpen(NOW, "outro-usuario")).toEqual([]);
  });

  it("clear apaga só o relato pedido", () => {
    addFlag(flag());
    addFlag(flag({ reportId: "relato-2", matchId: "partida-2" }));

    clear("relato-1");

    expect(draftFor("partida-1", "usuario-1")).toBeUndefined();
    expect(draftFor("partida-2", "usuario-1")?.reportId).toBe("relato-2");
  });

  it("não lança com JSON corrompido", () => {
    storage.setItem(BUG_DRAFTS_KEY, "{isso não é json");

    expect(() => draftFor("partida-1", "usuario-1")).not.toThrow();
    expect(draftFor("partida-1", "usuario-1")).toBeUndefined();
    expect(pendingOnOpen(NOW, "usuario-1")).toEqual([]);
    expect(addFlag(flag()).turns).toEqual([3]);
  });

  it("não lança quando o JSON não é uma lista", () => {
    storage.setItem(BUG_DRAFTS_KEY, JSON.stringify({ draft: flag() }));

    expect(pendingOnOpen(NOW, "usuario-1")).toEqual([]);
  });

  it("descarta entradas com formato inválido sem perder as boas", () => {
    storage.setItem(BUG_DRAFTS_KEY, JSON.stringify([flag(), { matchId: "sem-turnos", turns: "3" }]));

    expect(pendingOnOpen(NOW, "usuario-1")).toEqual([flag()]);
  });

  it("corta turnos extras que já estavam no armazenamento", () => {
    storage.setItem(BUG_DRAFTS_KEY, JSON.stringify([flag({ turns: Array.from({ length: 12 }, (_, i) => i) })]));

    expect(pendingOnOpen(NOW, "usuario-1")[0]?.turns).toEqual(Array.from({ length: 10 }, (_, i) => i));
  });

  it("não lança quando o localStorage está bloqueado", () => {
    vi.stubGlobal("localStorage", {
      getItem: () => {
        throw new Error("bloqueado");
      },
      setItem: () => {
        throw new Error("bloqueado");
      },
      removeItem: () => {
        throw new Error("bloqueado");
      },
    });

    expect(() => addFlag(flag())).not.toThrow();
    expect(draftFor("partida-1", "usuario-1")).toBeUndefined();
  });
});
