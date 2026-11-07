import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("../config/envs", () => ({ ENV: { VITE_BACKEND_URL: "http://backend" } }));

import handleCredentialResponse from "./handleCredentialResponse";
import { copy } from "../game/copy";

const jwt = (claims: object) => `h.${btoa(JSON.stringify(claims))}.s`;
const reply = (body: object) => ({ json: async () => body }) as Response;

function run(replies: object[], username?: string) {
  const fetchMock = vi.fn();
  replies.forEach((r) => fetchMock.mockResolvedValueOnce(reply(r)));
  vi.stubGlobal("fetch", fetchMock);
  const setAuth = vi.fn();
  return {
    setAuth,
    fetchMock,
    done: handleCredentialResponse({ credential: "g" }, setAuth, username),
  };
}

afterEach(() => vi.unstubAllGlobals());

describe("handleCredentialResponse", () => {
  it("login existente grava token e id", async () => {
    const token = jwt({ id: "p1" });
    const t = run([{ message: "logged_in", data: { token } }]);
    await expect(t.done).resolves.toEqual({ status: "ok" });
    expect(t.setAuth).toHaveBeenCalledWith(token, "p1");
  });

  it("conta nova: manda o nome e grava o token da segunda chamada", async () => {
    const token = jwt({ id: "p2" });
    const t = run([{ message: "need_additional_data" }, { message: "logged_in", data: { token } }], "Ana");
    await expect(t.done).resolves.toEqual({ status: "ok" });
    expect(JSON.parse(t.fetchMock.mock.calls[1][1].body)).toEqual({ IdToken: "g", username: "Ana" });
    expect(t.setAuth).toHaveBeenCalledWith(token, "p2");
  });

  it("sem nome, pede o nome e não faz a segunda chamada nem grava nada", async () => {
    const t = run([{ message: "need_additional_data" }]);
    await expect(t.done).resolves.toEqual({ status: "need_name" });
    expect(t.fetchMock).toHaveBeenCalledTimes(1);
    expect(t.setAuth).not.toHaveBeenCalled();
  });

  it("segunda chamada sem token (ou token sem id) devolve erro e não grava login", async () => {
    const t = run([{ message: "need_additional_data" }, { message: "username_taken" }], "Ana");
    await expect(t.done).resolves.toEqual({
      status: "error",
      message: copy.loginModal.error(copy.loginModal.usernameTaken),
    });

    const semId = run([{ message: "logged_in", data: { token: jwt({}) } }]);
    await expect(semId.done).resolves.toEqual({
      status: "error",
      message: copy.loginModal.error(copy.loginModal.noToken),
    });
    expect(semId.setAuth).not.toHaveBeenCalled();
  });

  it("servidor fora do ar devolve aviso sem quebrar", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => {
      throw new TypeError("Failed to fetch");
    }));
    const setAuth = vi.fn();
    await expect(handleCredentialResponse({ credential: "g" }, setAuth)).resolves.toEqual({
      status: "error",
      message: copy.loginModal.offline,
    });
    expect(setAuth).not.toHaveBeenCalled();
  });
});
