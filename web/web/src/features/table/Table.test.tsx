import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture, type FixtureName } from "../../game/fixtures";
import type { GameView } from "../../game/selectors";
import { Table } from "./Table";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const view = (name: FixtureName): GameView => {
  const f = fixture(name, NOW);
  return { me: f.me, room: f.room, log: f.log };
};
const html = (name: FixtureName) => renderToStaticMarkup(<Table view={view(name)} unread={2} />);

describe("Mesa (fumaça em Node)", () => {
  it("Mesa6 desenha cabeçalho, camarotes, palco, ribalta e faixa Você", () => {
    const out = html("Mesa6");
    expect(out).toContain('aria-label="Camarotes, toque para ver detalhes"');
    expect(out).toContain('aria-label="Ver camarote de Ana, 7 moedas"');
    expect(out).toContain('aria-label="Ver camarote de Davi, fora do jogo"');
    expect(out).toContain('aria-label="Palco"');
    expect(out).toContain("Rodada 4 · Taxa +1");
    expect(out).toContain("Conversa da mesa, 2 novas");
    expect(out).toContain(">Carla</h1>");
    expect(out).toContain('aria-label="12 cartas no baralho"');
    expect(out).toContain('aria-label="Taxa atual: 1"');
    expect(out).toContain("Político");
    expect(out).toContain("×2");
    expect(out).toContain("Rebelde");
    expect(out).toContain("×4");
    expect(out).toContain('aria-label="Você"');
    expect(out).toContain("Guardiã");
    expect(out).toContain("Trilionário");
    expect(out).toContain('aria-label="7 moedas"');
    expect(out).toContain("limite 20");
    expect(out).toContain("1 investimento, rende em 2");
    expect(out).toContain('aria-label="Tempo da jogada de Carla"');
  });

  it("Mesa3 e Mesa10 mudam o número de camarotes", () => {
    expect(html("Mesa3").match(/Ver camarote de/g)).toHaveLength(2);
    const out10 = html("Mesa10");
    expect(out10.match(/Ver camarote de/g)).toHaveLength(9);
    expect(out10).toContain("Iracema");
  });
});
