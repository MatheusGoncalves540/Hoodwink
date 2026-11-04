/**
 * Estado da partida vindo do servidor (04-arquitetura.md → gameStore).
 * applyRoom(room, now): guarda a sala, monta o diário (game/log.ts) e põe o evento novo na fila de apresentação.
 * Reaplicar a mesma sala (reconexão) não duplica diário nem fila.
 * loadFixture(name): carrega um estado das pranchas, só para desenvolvimento.
 */
import { create } from "zustand";
import { fixture, type FixtureName } from "../game/fixtures";
import { appendLog, eventKey, type LogEntry } from "../game/log";
import type { GameEvent, PublicRoom } from "../game/protocol";
import { useUiStore } from "./uiStore";

export type Connection = "idle" | "connecting" | "open" | "closed" | "error";

export type GameState = {
  me: string | null;
  room: PublicRoom | null;
  presentation: GameEvent[];
  log: LogEntry[];
  connection: Connection;
  setMe(id: string | null): void;
  setConnection(c: Connection): void;
  applyRoom(room: PublicRoom, now?: number): void;
  shift(): void;
  reset(): void;
  loadFixture(name: FixtureName, now?: number): void;
};

/**
 * ponytail: numa jogada com dados confidenciais o servidor também manda a versão pública da sala a você,
 * sem os nomes das suas cartas (rooms/roomBroadcast.go). Mantemos o nome que você já conhecia enquanto
 * a carta não mudar de estado; se o servidor corrigir o envio, isto vira no-op.
 */
function keepMyCardNames(prev: PublicRoom | null, next: PublicRoom, me: string | null): PublicRoom {
  const was = me ? prev?.players[me] : undefined;
  const now = me ? next.players[me] : undefined;
  if (!was || !now || now.cards.every((c) => c.name)) return next;
  const cards = now.cards.map((c) => {
    const old = was.cards.find((o) => o.index === c.index);
    return !c.name && old?.name && old.dead === c.dead ? { ...c, name: old.name } : c;
  });
  return { ...next, players: { ...next.players, [now.id]: { ...now, cards } } };
}

export const useGameStore = create<GameState>((set, get) => ({
  me: null,
  room: null,
  presentation: [],
  log: [],
  connection: "idle",
  setMe: (me) => set({ me }),
  setConnection: (connection) => set({ connection }),
  applyRoom: (incoming, now = Date.now()) => {
    const { room: prev, log, presentation, me } = get();
    const room = keepMyCardNames(prev, incoming, me);
    const ev = room.gameEvent;
    // Evento já vencido (reconexão/reload) entra no diário mas não na fila: a janela de animação acabou.
    const fresh =
      !!ev &&
      eventKey(ev) !== eventKey(prev?.gameEvent ?? null) &&
      ev.expiresAt > now &&
      !presentation.some((e) => eventKey(e) === eventKey(ev));
    set({
      room,
      log: appendLog(log, prev, room, now),
      presentation: fresh && !log.some((e) => e.key === eventKey(ev)) ? [...presentation, ev] : presentation,
    });
  },
  shift: () => set({ presentation: get().presentation.slice(1) }),
  reset: () => set({ room: null, log: [], presentation: [] }),
  loadFixture: (name, now = Date.now()) => {
    const f = fixture(name, now);
    set({ me: f.me, room: f.room, log: f.log, presentation: [], connection: "open" });
    useUiStore.setState({ openSheet: f.ui.openSheet, chatSeen: f.ui.chatSeen, targeting: null });
  },
}));
