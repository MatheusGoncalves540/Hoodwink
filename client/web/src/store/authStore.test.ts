import { afterEach, describe, expect, it, vi } from "vitest";
import { useAuthStore } from "./authStore";

const mkToken = (payload: object) => `h.${btoa(JSON.stringify(payload))}.s`;

afterEach(() => {
  vi.unstubAllGlobals();
  useAuthStore.setState({ jwtToken: "", playerId: "", isLoginOpen: false });
});

describe("authStore", () => {
  it("login concluído fecha o modal de login", () => {
    useAuthStore.getState().openLogin();
    expect(useAuthStore.getState().isLoginOpen).toBe(true);
    useAuthStore.getState().setAuth("token", "p1");
    expect(useAuthStore.getState()).toMatchObject({ jwtToken: "token", playerId: "p1", isLoginOpen: false });
  });

  it("401 do refresh limpa a sessão e abre o login", async () => {
    const now = Math.floor(Date.now() / 1000);
    const token = mkToken({ id: "p1", iat: now - 7200, exp: now + 3600 });
    useAuthStore.setState({ jwtToken: token, playerId: "p1", isLoginOpen: false });

    vi.stubGlobal("fetch", vi.fn(async () => new Response(JSON.stringify({ message: "inválido" }), { status: 401 })));

    await useAuthStore.getState().refreshIfNeeded();
    expect(useAuthStore.getState().jwtToken).toBe("");
    expect(useAuthStore.getState().isLoginOpen).toBe(true);
  });
});
