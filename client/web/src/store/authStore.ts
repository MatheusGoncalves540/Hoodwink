import { create } from "zustand";
import { persist } from "zustand/middleware";
import isTokenExpired from "../Auth/isTokenExpired";

interface AuthState {
  jwtToken: string;
  playerId: string;
  isLoginOpen: boolean;

  setAuth: (token: string, playerId: string) => void;
  openLogin: () => void;
  closeLogin: () => void;
}

/** O que sobrevive ao reload; o modal de login sempre começa fechado. */
type PersistedAuth = Pick<AuthState, "jwtToken" | "playerId">;

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      jwtToken: "",
      playerId: "",
      isLoginOpen: false,

      // login concluído fecha a bilheteria
      setAuth: (jwtToken, playerId) => set({ jwtToken, playerId, isLoginOpen: false }),

      openLogin: () => set({ isLoginOpen: true }),

      closeLogin: () => set({ isLoginOpen: false }),
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
