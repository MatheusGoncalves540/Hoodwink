import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("../config/envs", () => ({ ENV: { VITE_BACKEND_URL: "http://backend" } }));

import loginLocal from "./loginLocal";

const jwt = (claims: object) => `h.${btoa(JSON.stringify(claims))}.s`;

function run(body: object | null) {
  const fetchMock = vi.fn(async (_url: string, _init?: RequestInit) => ({
    json: async () => body,
  }));
  vi.stubGlobal("fetch", fetchMock);
  const alertMock = vi.fn();
  vi.stubGlobal("alert", alertMock);
  const setAuth = vi.fn();
  return { setAuth, alertMock, fetchMock, done: loginLocal("Ana", setAuth) };
}

afterEach(() => vi.unstubAllGlobals());

describe("loginLocal", () => {
  it("manda o nome para /auth/external/local e grava token e id", async () => {
    const token = jwt({ id: "p1" });
    const t = run({ message: "logged_in", data: { token } });
    await t.done;
    expect(t.fetchMock.mock.calls[0][0]).toBe("http://backend/auth/external/local");
    expect(JSON.parse(t.fetchMock.mock.calls[0][1]!.body as string)).toEqual({ username: "Ana" });
    expect(t.setAuth).toHaveBeenCalledWith(token, "p1");
  });

  it("sem token (backend fora de local) avisa e não grava login", async () => {
    const t = run({ message: "Provedor não suportado" });
    await t.done;
    expect(t.setAuth).not.toHaveBeenCalled();
    expect(t.alertMock).toHaveBeenCalledWith("Erro no login: Provedor não suportado");
  });

  it("servidor fora do ar avisa e não grava login", async () => {
    const fetchMock = vi.fn(async () => {
      throw new TypeError("Failed to fetch");
    });
    vi.stubGlobal("fetch", fetchMock);
    const alertMock = vi.fn();
    vi.stubGlobal("alert", alertMock);
    const setAuth = vi.fn();

    await loginLocal("Ana", setAuth);

    expect(setAuth).not.toHaveBeenCalled();
    expect(alertMock).toHaveBeenCalledWith(
      "Não deu para falar com o servidor do jogo. Confira se ele está no ar.",
    );
  });

  it("resposta sem corpo não quebra", async () => {
    const t = run(null);
    await t.done;
    expect(t.setAuth).not.toHaveBeenCalled();
    expect(t.alertMock).toHaveBeenCalledWith("Erro no login: resposta sem token");
  });
});
