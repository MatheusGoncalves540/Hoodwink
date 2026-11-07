import { afterEach, describe, expect, it, vi } from "vitest";
import { ENV } from "../config/envs";
import { copy } from "../game/copy";
import {
  ReplayApiError,
  clampMatchesLimit,
  downloadReplay,
  listMyMatches,
  replayFileName,
  type MyMatch,
} from "./replayApi";

const base = ENV.VITE_BACKEND_URL.replace(/\/$/, "");
const matchId = "01936f5e-7a2b-7c3d-8e9f-0123456789ab";

const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });

const stubFetch = (res: Response) => {
  const fn = vi.fn(async (_url: string, _init?: RequestInit) => res);
  vi.stubGlobal("fetch", fn);
  return fn;
};

const match: MyMatch = {
  match_id: matchId,
  room_id: "sala-1",
  engine_version: 1,
  started_at: "2026-11-06T12:00:00Z",
  ended_at: "2026-11-06T12:20:00Z",
  winner_id: "jogador-a",
  player_ids: ["jogador-a", "jogador-b"],
};

afterEach(() => vi.unstubAllGlobals());

describe("clampMatchesLimit", () => {
  it("prende o limit em 1..100 e usa 30 quando o valor não é número", () => {
    expect(clampMatchesLimit(30)).toBe(30);
    expect(clampMatchesLimit(0)).toBe(1);
    expect(clampMatchesLimit(500)).toBe(100);
    expect(clampMatchesLimit(12.9)).toBe(12);
    expect(clampMatchesLimit(Number.NaN)).toBe(30);
  });
});

describe("listMyMatches", () => {
  it("pede 30 por padrão, com o JWT no Authorization e cache no-store", async () => {
    const fetch = stubFetch(json({ error: false, message: "", data: [match] }));

    await expect(listMyMatches("jwt")).resolves.toEqual([match]);

    const [url, init] = fetch.mock.calls[0]!;
    expect(url).toBe(`${base}/me/matches?limit=30`);
    expect(init?.headers).toEqual({ Authorization: "Bearer jwt" });
    expect(init?.cache).toBe("no-store");
    expect(url).not.toContain("jwt");
  });

  it("respeita o limit pedido e nunca passa de 100", async () => {
    const fetch = vi.fn(async () => json({ data: [] }));
    vi.stubGlobal("fetch", fetch);

    await listMyMatches("jwt", 500);
    expect(fetch.mock.calls[0]![0]).toBe(`${base}/me/matches?limit=100`);

    await listMyMatches("jwt", 0);
    expect(fetch.mock.calls[1]![0]).toBe(`${base}/me/matches?limit=1`);
  });

  it("aceita winner_id nulo e room_id ausente", async () => {
    const { room_id: _room, winner_id: _winner, ...withoutOptional } = match;
    stubFetch(json({ data: [{ ...withoutOptional, winner_id: null }] }));

    await expect(listMyMatches("jwt")).resolves.toEqual([{ ...withoutOptional, winner_id: null }]);
  });

  it("sem token devolve 401 sem chamar a rede", async () => {
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);

    const error = await listMyMatches("").catch((e: unknown) => e);
    expect(error).toBeInstanceOf(ReplayApiError);
    expect(error).toMatchObject({ status: 401, message: copy.perfil.needLogin });
    expect(fetch).not.toHaveBeenCalled();
  });

  it("401, 403 e 404 viram ReplayApiError com o status certo", async () => {
    for (const [status, message] of [
      [401, copy.perfil.needLogin],
      [403, copy.perfil.forbidden],
      [404, copy.perfil.notFound],
    ] as const) {
      stubFetch(json({ message: "erro" }, status));
      const error = await listMyMatches("jwt").catch((e: unknown) => e);
      expect(error).toBeInstanceOf(ReplayApiError);
      expect(error).toMatchObject({ status, message });
    }
  });

  it("rede fora do ar vira status 0", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => {
      throw new Error("rede");
    }));

    const error = await listMyMatches("jwt").catch((e: unknown) => e);
    expect(error).toMatchObject({ status: 0, message: copy.perfil.network });
  });

  it("resposta fora do schema é recusada", async () => {
    stubFetch(json({ data: [{ match_id: matchId }] }));

    const error = await listMyMatches("jwt").catch((e: unknown) => e);
    expect(error).toBeInstanceOf(ReplayApiError);
    expect(error).toMatchObject({ status: 200, message: copy.perfil.badResponse });
  });

  it("abort não vira erro de rede", async () => {
    const aborted = Object.assign(new Error("aborted"), { name: "AbortError" });
    vi.stubGlobal("fetch", vi.fn(async () => {
      throw aborted;
    }));

    await expect(listMyMatches("jwt", 30, new AbortController().signal)).rejects.toBe(aborted);
  });
});

describe("downloadReplay", () => {
  const binary = (disposition?: string) =>
    new Response(new Blob([new Uint8Array([1, 2, 3])]), {
      status: 200,
      headers: disposition ? { "Content-Disposition": disposition } : undefined,
    });

  it("devolve o Blob e o nome do Content-Disposition", async () => {
    const fetch = stubFetch(binary('attachment; filename="hoodwink-2026-11-06-01936f5e.hwr"'));

    const file = await downloadReplay("jwt", matchId);

    expect(file.filename).toBe("hoodwink-2026-11-06-01936f5e.hwr");
    expect(file.blob.size).toBe(3);
    expect(new Uint8Array(await file.blob.arrayBuffer())).toEqual(new Uint8Array([1, 2, 3]));

    const [url, init] = fetch.mock.calls[0]!;
    expect(url).toBe(`${base}/replays/${matchId}`);
    expect(init?.headers).toEqual({ Authorization: "Bearer jwt" });
    expect(init?.cache).toBe("no-store");
    expect(url).not.toContain("jwt");
  });

  it("sem Content-Disposition usa hoodwink-<matchId8>.hwr", async () => {
    stubFetch(binary());
    await expect(downloadReplay("jwt", matchId)).resolves.toMatchObject({
      filename: "hoodwink-01936f5e.hwr",
    });
  });

  it("matchId vazio devolve 404 sem chamar a rede", async () => {
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);

    const error = await downloadReplay("jwt", "  ").catch((e: unknown) => e);
    expect(error).toMatchObject({ status: 404, message: copy.perfil.notFound });
    expect(fetch).not.toHaveBeenCalled();
  });

  it("404 do servidor vira ReplayApiError 404", async () => {
    stubFetch(json({ message: "replay não encontrado" }, 404));

    const error = await downloadReplay("jwt", matchId).catch((e: unknown) => e);
    expect(error).toBeInstanceOf(ReplayApiError);
    expect(error).toMatchObject({ status: 404, message: copy.perfil.notFound });
  });

  it("rede fora do ar vira status 0", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => {
      throw new Error("rede");
    }));

    const error = await downloadReplay("jwt", matchId).catch((e: unknown) => e);
    expect(error).toMatchObject({ status: 0, message: copy.perfil.network });
  });
});

describe("replayFileName", () => {
  it("lê filename* com UTF-8 e o filename simples", () => {
    expect(replayFileName("attachment; filename*=UTF-8''hoodwink%2D2026.hwr", matchId)).toBe(
      "hoodwink-2026.hwr",
    );
    expect(replayFileName('attachment; filename="hoodwink-2026.hwr"', matchId)).toBe("hoodwink-2026.hwr");
  });

  it("remove separadores de caminho e caracteres de controle", () => {
    expect(replayFileName('attachment; filename="../../etc/passwd"', matchId)).toBe("....etcpasswd");
    expect(replayFileName('attachment; filename="a\u0000b.hwr"', matchId)).toBe("ab.hwr");
  });

  it("cai no fallback quando o cabeçalho não traz nome útil", () => {
    expect(replayFileName(null, matchId)).toBe("hoodwink-01936f5e.hwr");
    expect(replayFileName('attachment; filename=""', matchId)).toBe("hoodwink-01936f5e.hwr");
    expect(replayFileName('attachment; filename="///"', matchId)).toBe("hoodwink-01936f5e.hwr");
  });
});
