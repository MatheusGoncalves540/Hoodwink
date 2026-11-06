import { afterEach, describe, expect, it, vi } from "vitest";
import { ENV } from "../config/envs";
import { getTicket, newRoom } from "./api";

const base = ENV.VITE_GAMESERVER_URL.replace(/\/$/, "");
const json = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
const stubFetch = (res: Response) => {
  const fn = vi.fn(async (_url: string, _init?: RequestInit) => res);
  vi.stubGlobal("fetch", fn);
  return fn;
};

afterEach(() => vi.unstubAllGlobals());

describe("api", () => {
  it("cria a sala com o JWT e o corpo camelCase do servidor", async () => {
    const fetch = stubFetch(json({ RoomId: "K7Q2", message: "Sala criada" }, 201));
    await expect(newRoom("jwt", { roomName: "Sala", maxPlayers: 6 })).resolves.toBe("K7Q2");

    const [url, init] = fetch.mock.calls[0]!;
    expect(url).toBe(`${base}/newRoom`);
    expect(init).toEqual({
      method: "POST",
      headers: { Authorization: "Bearer jwt", "Content-Type": "application/json" },
      body: JSON.stringify({ roomName: "Sala", password: "", maxPlayers: 6 }),
    });
  });

  it("pede o ticket da sala e devolve só o token", async () => {
    const fetch = stubFetch(json({ message: "Ticket gerado", data: { ticket: "abc.def.ghi" } }));
    await expect(getTicket("jwt", "K7Q2")).resolves.toBe("abc.def.ghi");
    expect(fetch.mock.calls[0]![0]).toBe(`${base}/getTicket/K7Q2`);
  });

  it("escapa o id da sala no caminho", async () => {
    const fetch = stubFetch(json({ data: { ticket: "t" } }));
    await getTicket("jwt", "a/b c");
    expect(fetch.mock.calls[0]![0]).toBe(`${base}/getTicket/a%2Fb%20c`);
  });

  it("propaga a mensagem de erro do servidor", async () => {
    stubFetch(json({ error: "Erro ao processar a requisição", message: "Essa sala está cheia." }, 401));
    await expect(getTicket("jwt", "K7Q2")).rejects.toThrow("Essa sala está cheia.");
  });

  it("recusa resposta de sucesso em formato inesperado", async () => {
    stubFetch(json({ message: "Ticket gerado", data: {} }));
    await expect(getTicket("jwt", "K7Q2")).rejects.toThrow();
  });
});
