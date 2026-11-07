import { create } from "zustand";
import { persist } from "zustand/middleware";
import decodeJWT from "../Auth/decodeJWT";
import isTokenExpired from "../Auth/isTokenExpired";
import shouldRefresh from "../Auth/shouldRefresh";
import { refreshToken } from "../net/api";

interface AuthState {
  jwtToken: string;
  playerId: string;
  isLoginOpen: boolean;

  setAuth: (token: string, playerId: string) => void;
  openLogin: () => void;
  closeLogin: () => void;
  clearSession: () => void;
  refreshIfNeeded: () => Promise<void>;
}

/** O que sobrevive ao reload; o modal de login sempre começa fechado. */
type PersistedAuth = Pick<AuthState, "jwtToken" | "playerId">;

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      jwtToken: "",
      playerId: "",
      isLoginOpen: false,

      // login concluído fecha a bilheteria
      setAuth: (jwtToken, playerId) => set({ jwtToken, playerId, isLoginOpen: false }),

      openLogin: () => set({ isLoginOpen: true }),

      closeLogin: () => set({ isLoginOpen: false }),

      // 401 do refresh (ou logout): apaga a sessão e abre o login de novo
      clearSession: () => set({ jwtToken: "", playerId: "", isLoginOpen: true }),

      // renova ao carregar se já passou da metade da vida; rede falhou mantém o token
      refreshIfNeeded: async () => {
        const { jwtToken, setAuth, clearSession } = get();
        if (!jwtToken || isTokenExpired(jwtToken)) return;
        if (!shouldRefresh(jwtToken, Math.floor(Date.now() / 1000))) return;

        const outcome = await refreshToken(jwtToken);
        if (outcome.status === "ok") {
          const id = decodeJWT<{ id?: string }>(outcome.token)?.id ?? "";
          if (id) setAuth(outcome.token, id);
        } else if (outcome.status === "unauthorized") {
          clearSession();
        }
      },
    }),
    {
      name: "hoodwink-auth",
      partialize: ({ jwtToken, playerId }): PersistedAuth => ({ jwtToken, playerId }),
      merge: (persisted, current) => {
        const saved = persisted as Partial<PersistedAuth> | undefined;
        if (!saved?.jwtToken || isTokenExpired(saved.jwtToken)) return current;
        return { ...current, jwtToken: saved.jwtToken, playerId: saved.playerId ?? "" };
      },
    },
  ),
);
