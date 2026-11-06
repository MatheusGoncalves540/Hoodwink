import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture, type FixtureName } from "../../game/fixtures";
import type { GameView } from "../../game/selectors";
import { DesktopRail } from "./DesktopRail";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const view = (name: FixtureName): GameView => {
  const f = fixture(name, NOW);
  return { me: f.me, room: f.room, log: f.log };
};

describe("trilho do desktop (fumaça em Node)", () => {
  it("mostra o diário, a conversa e o campo", () => {
    const out = renderToStaticMarkup(<DesktopRail view={view("Mesa6")} />);
    expect(out).toContain("Histórico da partida");
    expect(out).toContain("rodada 4");
    expect(out).toContain("Bruno roubou 3 de Eva como Trapaceiro.");
    expect(out).toContain("Conversa");
    expect(out).toContain("Eva.");
    expect(out).toContain("Alguém conta as moedas da Carla, por favor.");
    expect(out).toContain("Diga algo à mesa");
    expect(out).toContain("Enviar");
  });

  it("sem diário, mostra o aviso de mesa vazia", () => {
    const v = view("Mesa6");
    const out = renderToStaticMarkup(<DesktopRail view={{ ...v, log: [] }} />);
    expect(out).toContain("Ainda nada. A mesa observa.");
  });
});
