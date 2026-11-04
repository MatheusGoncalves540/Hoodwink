/**
 * Estado só de interface (04-arquitetura.md → uiStore).
 * - openSheet: a gaveta aberta (uma por vez).
 * - chatSeen: quantas falas você já viu (o resto vira "N novas" no balão).
 * - targeting: o fluxo de alvo inferido (05-protocolo.md): papel escolhido esperando o camarote.
 * - suddenDeathShown: o aviso de morte súbita toca uma vez por partida (guarda o id da sala).
 */
import { create } from "zustand";
import type { OpenSheet } from "../game/fixtures";
import type { RoleId } from "../game/protocol";

export type UiState = {
  openSheet: OpenSheet;
  chatDraft: string;
  chatSeen: number;
  targeting: { role: RoleId; target?: { playerId: string; cardIndex: number } } | null;
  suddenDeathShown: string | null;
  open(sheet: OpenSheet): void;
  close(): void;
  setDraft(text: string): void;
  markChatSeen(count: number): void;
  setTargeting(t: UiState["targeting"]): void;
  setSuddenDeathShown(roomId: string): void;
};

export const useUiStore = create<UiState>((set) => ({
  openSheet: null,
  chatDraft: "",
  chatSeen: 0,
  targeting: null,
  suddenDeathShown: null,
  open: (openSheet) => set({ openSheet }),
  close: () => set({ openSheet: null }),
  setDraft: (chatDraft) => set({ chatDraft }),
  markChatSeen: (chatSeen) => set({ chatSeen }),
  setTargeting: (targeting) => set({ targeting }),
  setSuddenDeathShown: (suddenDeathShown) => set({ suddenDeathShown }),
}));
