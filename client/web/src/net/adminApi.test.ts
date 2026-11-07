import { afterEach, describe, expect, it, vi } from "vitest";
import { ENV } from "../config/envs";
import {
  AdminApiError,
  adminBugReportsQuery,
  adminRequestLogsQuery,
  listBugReports,
  listRequestLogs,
  updateBugReportStatus,
  type AdminBugReport,
  type AdminRequestLog,
} from "./adminApi";

const base = ENV.VITE_BACKEND_URL.replace(/\/$/, "");
const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
const stubFetch = (res: Response) => {
  const fn = vi.fn(async (_url: string, _init?: RequestInit) => res);
  vi.stubGlobal("fetch", fn);
  return fn;
};

const item: AdminRequestLog = {
  id: 7,
  at: "2026-11-06T12:00:00Z",
  service: "backend",
  method: "GET",
  path: "/getTicket/K7Q2",
  route: "/getTicket/{RoomId}",
  status: 200,
  duration_ms: 12,
  ip: "127.0.0.1",
  user_agent: "Mozilla/5.0",
  user_id: null,
  username: "ana",
  request_id: "6f1c1d3e-2a4b-4c5d-8e9f-0123456789ab",
  body: null,
  error: null,
};

afterEach(() => vi.unstubAllGlobals());

describe("adminRequestLogsQuery", () => {
  it("serializa todos os filtros com URLSearchParams", () => {
    const params = adminRequestLogsQuery({
      from: "2026-11-05T00:00:00Z",
      to: "2026-11-06T00:00:00Z",
      route: "/getTicket",
      status: "4xx",
      userId: "6f1c1d3e-2a4b-4c5d-8e9f-0123456789ab",
      username: "ana",
      service: "gameServer",
      ip: "127.0.0.1",
      archived: true,
      page: 2,
      limit: 100,
    });

    expect(Object.fromEntries(params)).toEqual({
      from: "2026-11-05T00:00:00Z",
      to: "2026-11-06T00:00:00Z",
      route: "/getTicket",
      status: "4xx",
      userId: "6f1c1d3e-2a4b-4c5d-8e9f-0123456789ab",
      username: "ana",
      service: "gameServer",
      ip: "127.0.0.1",
      archived: "true",
      page: "2",
      limit: "100",
    });
  });

  it("omite vazios, booleanos falsos e undefined", () => {
    const params = adminRequestLogsQuery({
      from: "",
      to: "  ",
      route: "",
      status: "",
      userId: "",
      username: "",
      service: "",
      ip: "",
      archived: false,
    });
    expect(params.toString()).toBe("");
  });

  it("escapa caracteres especiais no valor", () => {
    const params = adminRequestLogsQuery({ route: "/a b?c=d&100%" });
    expect(params.get("route")).toBe("/a b?c=d&100%");
    expect(params.toString()).toContain("route=%2Fa+b%3Fc%3Dd%26100%25");
  });
});

describe("listRequestLogs", () => {
  it("monta a URL com query e o JWT no Authorization", async () => {
    const fetch = stubFetch(json({ items: [], total: 0, page: 1, limit: 50 }));
    await listRequestLogs("jwt", {
      from: "2026-11-05T00:00:00Z",
      to: "2026-11-06T00:00:00Z",
      route: "/getTicket",
      archived: true,
      page: 2,
      limit: 100,
    });

    const [url, init] = fetch.mock.calls[0]!;
    const parsed = new URL(url);
    expect(`${parsed.origin}${parsed.pathname}`).toBe(`${base}/admin/requestLogs`);
    expect(Object.fromEntries(parsed.searchParams)).toEqual({
      from: "2026-11-05T00:00:00Z",
      to: "2026-11-06T00:00:00Z",
      route: "/getTicket",
      archived: "true",
      page: "2",
      limit: "100",
    });
    expect(init?.headers).toEqual({ Authorization: "Bearer jwt" });
  });

  it("sem filtros não deixa interrogação sobrando", async () => {
    const fetch = stubFetch(json({ items: [], total: 0, page: 1, limit: 50 }));
    await listRequestLogs("jwt");
    expect(fetch.mock.calls[0]![0]).toBe(`${base}/admin/requestLogs`);
  });

  it("devolve a resposta tipada com os campos snake_case", async () => {
    stubFetch(json({ items: [item], total: 1, page: 1, limit: 50 }));
    await expect(listRequestLogs("jwt")).resolves.toEqual({ items: [item], total: 1, page: 1, limit: 50 });
  });

  it("401 vira AdminApiError com status 401 e mensagem própria", async () => {
    stubFetch(json({ error: true, message: "Token inválido" }, 401));
    const error = await listRequestLogs("jwt").catch((e: unknown) => e);
    expect(error).toBeInstanceOf(AdminApiError);
    expect(error).toMatchObject({ status: 401, message: "Sessão expirada. Entre de novo." });
  });

  it("403 vira AdminApiError com status 403 e 'Sem acesso'", async () => {
    stubFetch(json({ error: true, message: "Acesso negado" }, 403));
    const error = await listRequestLogs("jwt").catch((e: unknown) => e);
    expect(error).toBeInstanceOf(AdminApiError);
    expect(error).toMatchObject({ status: 403, message: "Sem acesso" });
  });

  it("outro erro usa a mensagem do servidor", async () => {
    stubFetch(json({ error: true, message: "erro ao consultar logs" }, 500));
    const error = await listRequestLogs("jwt").catch((e: unknown) => e);
    expect(error).toMatchObject({ status: 500, message: "erro ao consultar logs" });
  });

  it("resposta fora do schema é recusada", async () => {
    stubFetch(json({ items: [{ id: 1 }], total: 1, page: 1, limit: 50 }));
    const error = await listRequestLogs("jwt").catch((e: unknown) => e);
    expect(error).toBeInstanceOf(AdminApiError);
    expect(error).toMatchObject({ status: 200, message: "Resposta inesperada do servidor." });
  });

  it("rede fora do ar vira status 0", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => { throw new Error("rede"); }));
    const error = await listRequestLogs("jwt").catch((e: unknown) => e);
    expect(error).toMatchObject({ status: 0, message: "Não foi possível falar com o servidor." });
  });

  it("abort não vira erro de rede", async () => {
    const aborted = Object.assign(new Error("aborted"), { name: "AbortError" });
    vi.stubGlobal("fetch", vi.fn(async () => { throw aborted; }));
    await expect(listRequestLogs("jwt", {}, new AbortController().signal)).rejects.toBe(aborted);
  });
});

const matchId = "01936f5e-7a2b-7c3d-8e9f-0123456789ab";

const bug: AdminBugReport = {
  id: 9,
  created_at: "2026-11-06T12:00:00Z",
  updated_at: "2026-11-06T12:05:00Z",
  user_id: "6f1c1d3e-2a4b-4c5d-8e9f-0123456789ab",
  username: "ana",
  source: "partida",
  match_id: matchId,
  turns: [3, 4],
  category: "moedas",
  description: '<img src=x onerror=alert(1)>',
  context: { build: "abc", errors: [] },
  ip: "127.0.0.1",
  user_agent: "Mozilla/5.0",
  status: "novo",
};

describe("adminBugReportsQuery", () => {
  it("serializa status, origem, partida e paginação", () => {
    const params = adminBugReportsQuery({ status: "novo", source: "partida", matchId, page: 2, limit: 100 });

    expect(Object.fromEntries(params)).toEqual({
      status: "novo",
      source: "partida",
      matchId,
      page: "2",
      limit: "100",
    });
  });

  it("omite vazios e undefined", () => {
    const params = adminBugReportsQuery({ status: "", source: "", matchId: "  " });
    expect(params.toString()).toBe("");
  });
});

describe("listBugReports", () => {
  it("monta a URL com query e o JWT no Authorization", async () => {
    const fetch = stubFetch(json({ items: [bug], total: 1, novos: 1, page: 1, limit: 50 }));
    await listBugReports("jwt", { status: "novo", source: "partida", matchId, page: 2, limit: 50 });

    const [url, init] = fetch.mock.calls[0]!;
    const parsed = new URL(url);
    expect(`${parsed.origin}${parsed.pathname}`).toBe(`${base}/admin/bugReports`);
    expect(Object.fromEntries(parsed.searchParams)).toEqual({
      status: "novo",
      source: "partida",
      matchId,
      page: "2",
      limit: "50",
    });
    expect(init?.headers).toEqual({ Authorization: "Bearer jwt" });
  });

  it("sem filtros não deixa interrogação sobrando", async () => {
    const fetch = stubFetch(json({ items: [], total: 0, novos: 0, page: 1, limit: 50 }));
    await listBugReports("jwt");
    expect(fetch.mock.calls[0]![0]).toBe(`${base}/admin/bugReports`);
  });

  it("devolve a resposta tipada com os campos snake_case", async () => {
    stubFetch(json({ items: [bug], total: 1, novos: 1, page: 1, limit: 50 }));
    await expect(listBugReports("jwt")).resolves.toEqual({ items: [bug], total: 1, novos: 1, page: 1, limit: 50 });
  });

  it("aceita contexto nulo e campos opcionais vazios", async () => {
    const anon = { ...bug, user_id: null, username: null, match_id: null, turns: [], category: null, description: null, context: null, ip: null, user_agent: null };
    stubFetch(json({ items: [anon], total: 1, novos: 0, page: 1, limit: 50 }));
    await expect(listBugReports("jwt")).resolves.toEqual({ items: [anon], total: 1, novos: 0, page: 1, limit: 50 });
  });

  it("403 vira AdminApiError com status 403 e 'Sem acesso'", async () => {
    stubFetch(json({ error: true, message: "Acesso negado" }, 403));
    const error = await listBugReports("jwt").catch((e: unknown) => e);
    expect(error).toBeInstanceOf(AdminApiError);
    expect(error).toMatchObject({ status: 403, message: "Sem acesso" });
  });

  it("401 vira AdminApiError com status 401 e mensagem própria", async () => {
    stubFetch(json({ error: true, message: "Token inválido" }, 401));
    const error = await listBugReports("jwt").catch((e: unknown) => e);
    expect(error).toMatchObject({ status: 401, message: "Sessão expirada. Entre de novo." });
  });

  it("resposta fora do schema é recusada", async () => {
    stubFetch(json({ items: [{ id: 1 }], total: 1, novos: 1, page: 1, limit: 50 }));
    const error = await listBugReports("jwt").catch((e: unknown) => e);
    expect(error).toBeInstanceOf(AdminApiError);
    expect(error).toMatchObject({ status: 200, message: "Resposta inesperada do servidor." });
  });

  it("rede fora do ar vira status 0", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => { throw new Error("rede"); }));
    const error = await listBugReports("jwt").catch((e: unknown) => e);
    expect(error).toMatchObject({ status: 0, message: "Não foi possível falar com o servidor." });
  });

  it("abort não vira erro de rede", async () => {
    const aborted = Object.assign(new Error("aborted"), { name: "AbortError" });
    vi.stubGlobal("fetch", vi.fn(async () => { throw aborted; }));
    await expect(listBugReports("jwt", {}, new AbortController().signal)).rejects.toBe(aborted);
  });
});

describe("updateBugReportStatus", () => {
  it("faz PATCH com o status no corpo e devolve o relato atualizado", async () => {
    const updated = { ...bug, status: "resolvido" };
    const fetch = stubFetch(json(updated));

    await expect(updateBugReportStatus("jwt", 9, "resolvido")).resolves.toEqual(updated);

    const [url, init] = fetch.mock.calls[0]!;
    expect(url).toBe(`${base}/admin/bugReports/9`);
    expect(init?.method).toBe("PATCH");
    expect(init?.headers).toEqual({
      Authorization: "Bearer jwt",
      "Content-Type": "application/json",
    });
    expect(init?.body).toBe(JSON.stringify({ status: "resolvido" }));
  });

  it("400 usa a mensagem do servidor", async () => {
    stubFetch(json({ error: true, message: "status inválido" }, 400));
    const error = await updateBugReportStatus("jwt", 9, "resolvido").catch((e: unknown) => e);
    expect(error).toMatchObject({ status: 400, message: "status inválido" });
  });

  it("403 vira 'Sem acesso'", async () => {
    stubFetch(json({ error: true, message: "Acesso negado" }, 403));
    const error = await updateBugReportStatus("jwt", 9, "visto").catch((e: unknown) => e);
    expect(error).toBeInstanceOf(AdminApiError);
    expect(error).toMatchObject({ status: 403, message: "Sem acesso" });
  });

  it("404 usa a mensagem do servidor", async () => {
    stubFetch(json({ error: true, message: "relato não encontrado" }, 404));
    const error = await updateBugReportStatus("jwt", 9, "visto").catch((e: unknown) => e);
    expect(error).toMatchObject({ status: 404, message: "relato não encontrado" });
  });
});
