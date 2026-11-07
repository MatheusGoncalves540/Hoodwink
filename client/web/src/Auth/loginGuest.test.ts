import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("../config/envs", () => ({ ENV: { VITE_BACKEND_URL: "http://backend" } }));

import loginGuest from "./loginGuest";
import { copy } from "../game/copy";

const jwt = (claims: object) => `h.${btoa(JSON.stringify(claims))}.s`;

function run(body: object | null) {
  const fetchMock = vi.fn(async (_url: string, _init?: RequestInit) => ({
    json: async () => body,
  }));
  vi.stubGlobal("fetch", fetchMock);
  const setAuth = vi.fn();
  return { setAuth, fetchMock, done: loginGuest("Ana", setAuth) };
}

afterEach(() => vi.unstubAllGlobals());

describe("loginGuest", () => {
  it("manda o nome para /auth/external/guest e grava token e id", async () => {
    const token = jwt({ id: "p1" });
    const t = run({ message: "logged_in", data: { token } });
    await expect(t.done).resolves.toEqual({ status: "ok" });
    expect(t.fetchMock.mock.calls[0][0]).toBe("http://backend/auth/external/guest");
    expect(JSON.parse(t.fetchMock.mock.calls[0][1]!.body as string)).toEqual({ username: "Ana" });
    expect(t.setAuth).toHaveBeenCalledWith(token, "p1");
  });

  it("sem token devolve erro e não grava login", async () => {
    const t = run({ message: "Provedor não suportado" });
    await expect(t.done).resolves.toEqual({
      status: "error",
      message: copy.loginModal.error("Provedor não suportado"),
    });
    expect(t.setAuth).not.toHaveBeenCalled();
  });

  it("servidor fora do ar devolve aviso e não grava login", async () => {
    const fetchMock = vi.fn(async () => {
      throw new TypeError("Failed to fetch");
    });
    vi.stubGlobal("fetch", fetchMock);
    const setAuth = vi.fn();

    await expect(loginGuest("Ana", setAuth)).resolves.toEqual({
      status: "error",
      message: copy.loginModal.offline,
    });

    expect(setAuth).not.toHaveBeenCalled();
  });

  it("resposta sem corpo não quebra", async () => {
    const t = run(null);
    await expect(t.done).resolves.toEqual({
      status: "error",
      message: copy.loginModal.error(copy.loginModal.noToken),
    });
    expect(t.setAuth).not.toHaveBeenCalled();
  });
});
