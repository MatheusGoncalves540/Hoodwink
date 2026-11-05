import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("../config/envs", () => ({ ENV: { VITE_BACKEND_URL: "http://backend" } }));

import handleCredentialResponse from "./handleCredentialResponse";

const jwt = (claims: object) => `h.${btoa(JSON.stringify(claims))}.s`;
const reply = (body: object) => ({ json: async () => body }) as Response;

function run(replies: object[], username: string | null = "Ana") {
  const fetchMock = vi.fn();
  replies.forEach((r) => fetchMock.mockResolvedValueOnce(reply(r)));
  vi.stubGlobal("fetch", fetchMock);
  vi.stubGlobal("prompt", () => username);
  const alertMock = vi.fn();
  vi.stubGlobal("alert", alertMock);
  const setAuth = vi.fn();
  return { setAuth, alertMock, fetchMock, done: handleCredentialResponse({ credential: "g" }, setAuth) };
}

afterEach(() => vi.unstubAllGlobals());

describe("handleCredentialResponse", () => {
  it("login existente grava token e id", async () => {
    const token = jwt({ id: "p1" });
    const t = run([{ message: "logged_in", data: { token } }]);
    await t.done;
    expect(t.setAuth).toHaveBeenCalledWith(token, "p1");
  });

  it("conta nova: manda o nome e grava o token da segunda chamada", async () => {
    const token = jwt({ id: "p2" });
    const t = run([{ message: "need_additional_data" }, { message: "logged_in", data: { token } }]);
    await t.done;
    expect(JSON.parse(t.fetchMock.mock.calls[1][1].body)).toEqual({ IdToken: "g", username: "Ana" });
    expect(t.setAuth).toHaveBeenCalledWith(token, "p2");
  });

  it("prompt cancelado: não faz a segunda chamada nem grava nada", async () => {
    const t = run([{ message: "need_additional_data" }], null);
    await t.done;
    expect(t.fetchMock).toHaveBeenCalledTimes(1);
    expect(t.setAuth).not.toHaveBeenCalled();
  });

  it("segunda chamada sem token (ou token sem id) não grava login", async () => {
    const t = run([{ message: "need_additional_data" }, { message: "username_taken" }]);
    await t.done;
    expect(t.setAuth).not.toHaveBeenCalled();
    expect(t.alertMock).toHaveBeenCalledWith("Erro no login: username_taken");

    const semId = run([{ message: "logged_in", data: { token: jwt({}) } }]);
    await semId.done;
    expect(semId.setAuth).not.toHaveBeenCalled();
  });
});
