import { describe, expect, it } from "vitest";
import { useAuthStore } from "./authStore";

describe("authStore", () => {
  it("login concluído fecha o modal de login", () => {
    useAuthStore.getState().openLogin();
    expect(useAuthStore.getState().isLoginOpen).toBe(true);
    useAuthStore.getState().setAuth("token", "p1");
    expect(useAuthStore.getState()).toMatchObject({ jwtToken: "token", playerId: "p1", isLoginOpen: false });
  });
});
