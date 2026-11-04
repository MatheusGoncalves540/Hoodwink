import { beforeEach, describe, expect, it } from "vitest";
import { fixture } from "../game/fixtures";
import type { PublicRoom } from "../game/protocol";
import { useGameStore } from "./gameStore";
import { useUiStore } from "./uiStore";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const s = () => useGameStore.getState();
const event = (room: PublicRoom, type: string, playerId: string, at: number, payload: unknown = null): PublicRoom => ({
  ...room,
  gameEvent: { type, playerId, expiresAt: at, payload },
});

beforeEach(() => {
  useGameStore.setState({ me: "me", room: null, log: [], presentation: [] });
});

describe("gameStore", () => {
  it("loadFixture carrega sala, diário e gaveta da prancha", () => {
    s().loadFixture("Camarote", NOW);
    const f = fixture("Camarote", NOW);
    expect({ me: s().me, room: s().room, log: s().log, presentation: s().presentation }).toEqual({ me: "me", room: f.room, log: f.log, presentation: [] });
    expect(useUiStore.getState().openSheet).toEqual(f.ui.openSheet);
  });

  it("evento novo entra no diário e na fila; a mesma sala de novo (reconexão) não duplica", () => {
    const base = fixture("Mesa6", NOW).room!;
    const declared = event(base, "CARD_PLAYED_REBEL", "ana", NOW + 1);
    s().applyRoom(declared, NOW);
    s().applyRoom(declared, NOW + 50);
    expect(s().presentation).toEqual([declared.gameEvent]);
    expect(s().log).toHaveLength(1);

    s().shift();
    expect(s().presentation).toEqual([]);
    s().applyRoom(declared, NOW + 90);
    expect(s().presentation).toEqual([]);
    expect(s().log).toHaveLength(1);
  });

  it("mantém o nome das suas cartas quando a versão pública chega sem eles", () => {
    const base = fixture("Mesa6", NOW).room!;
    s().applyRoom(base, NOW);
    const me = base.players.me;
    const blind: PublicRoom = { ...base, players: { ...base.players, me: { ...me, cards: me.cards.map(({ name: _n, ...c }) => c) } } };
    s().applyRoom(blind, NOW + 1);
    expect(s().room).toEqual(base);
  });
});
