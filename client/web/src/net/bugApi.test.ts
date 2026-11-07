import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ENV } from "../config/envs";
import { useAuthStore } from "../store/authStore";
import { BugApiError, flagBug, patchBugReport, sendBugReport } from "./bugApi";

const base = ENV.VITE_BACKEND_URL.replace(/\/$/, "");
const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
const stubFetch = (res: Response) => {
  const fn = vi.fn(async (_url: string, _init?: RequestInit) => res);
  vi.stubGlobal("fetch", fn);
  return fn;
};

beforeEach(() => {
  useAuthStore.setState({ jwtToken: "", playerId: "", isLoginOpen: false });
});

afterEach(() => vi.unstubAllGlobals());

describe("flagBug", () => {
  it("manda o JWT, o matchId e o turno e devolve o id do relato", async () => {
    const fetch = stubFetch(json({ id: 7, match_id: "partida-1", turns: [3, 5] }, 201));
    await expect(flagBug("jwt", { matchId: "partida-1", turn: 5 })).resolves.toEqual({
      reportId: "7",
      turns: [3, 5],
    });

    const [url, init] = fetch.mock.calls[0]!;
    expect(url).toBe(`${base}/bugReports/flag`);
    expect(init).toEqual({
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: "Bearer jwt" },
      body: JSON.stringify({ matchId: "partida-1", turn: 5 }),
    });
  });

  it("aceita id como string no corpo de resposta", async () => {
    stubFetch(json({ id: "9", match_id: "p", turns: [] }));
    await expect(flagBug("jwt", { matchId: "p", turn: 1 })).resolves.toEqual({ reportId: "9", turns: [] });
  });

  it("401 vira BugApiError unauthorized e abre o login", async () => {
    stubFetch(json({ message: "Token inválido" }, 401));
    const error = await flagBug("jwt", { matchId: "p", turn: 1 }).catch((e: unknown) => e);
    expect(error).toBeInstanceOf(BugApiError);
    expect(error).toMatchObject({ status: 401, result: "unauthorized" });
    expect(useAuthStore.getState().isLoginOpen).toBe(true);
  });

  it("429 vira limit e mantém o token intacto", async () => {
    stubFetch(json({ message: "limite" }, 429));
    const error = await flagBug("jwt", { matchId: "p", turn: 1 }).catch((e: unknown) => e);
    expect(error).toMatchObject({ status: 429, result: "limit" });
  });
});

describe("sendBugReport", () => {
  it("sem token não manda Authorization (relato anônimo)", async () => {
    const fetch = stubFetch(json({ id: 1, source: "lobby" }, 201));
    await sendBugReport("", { category: "outro", description: "não entra na sala 123", context: null });

    const [url, init] = fetch.mock.calls[0]!;
    expect(url).toBe(`${base}/bugReports`);
    expect(init?.headers).toEqual({ "Content-Type": "application/json" });
    expect(JSON.parse((init?.body as string) ?? "{}")).toEqual({
      category: "outro",
      description: "não entra na sala 123",
      context: null,
    });
  });

  it("com token manda o Bearer", async () => {
    const fetch = stubFetch(json({ id: 1 }, 201));
    await sendBugReport("jwt", {
      category: "tela",
      description: "tela quebrada",
      context: { scene: "lobby", buildVersion: "dev", screen: "390x844", errors: [] },
    });
    expect((fetch.mock.calls[0]![1]?.headers as Record<string, string>).Authorization).toBe("Bearer jwt");
  });
});

describe("patchBugReport", () => {
  it("monta o PATCH no id com o Bearer e o corpo", async () => {
    const fetch = stubFetch(json({ id: 7, source: "partida" }));
    await patchBugReport("jwt", "7", { category: "moedas", description: "errei", context: null });

    const [url, init] = fetch.mock.calls[0]!;
    expect(url).toBe(`${base}/bugReports/7`);
    expect(init?.method).toBe("PATCH");
    expect((init?.headers as Record<string, string>).Authorization).toBe("Bearer jwt");
    expect(JSON.parse((init?.body as string) ?? "{}")).toEqual({ category: "moedas", description: "errei", context: null });
  });

  it("404 e 403 viram error com o status preservado (o cartão apaga)", async () => {
    for (const status of [403, 404]) {
      stubFetch(json({ message: "não existe" }, status));
      const error = await patchBugReport("jwt", "7", { category: null, description: null, context: null }).catch((e: unknown) => e);
      expect(error).toMatchObject({ status, result: "error" });
    }
  });

  it("rede fora do ar vira status 0", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => { throw new Error("rede"); }));
    const error = await patchBugReport("jwt", "7", { category: null, description: null, context: null }).catch((e: unknown) => e);
    expect(error).toMatchObject({ status: 0, result: "error" });
  });
});
