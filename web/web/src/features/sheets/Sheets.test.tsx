import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture, type FixtureName } from "../../game/fixtures";
import { selectSeatSheet, selectYouSheet, type GameView } from "../../game/selectors";
import { SeatSheetBody } from "./SeatSheet";
import { YouSheetBody } from "./YouSheet";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const view = (name: FixtureName): GameView => {
  const f = fixture(name, NOW);
  return { me: f.me, room: f.room, log: f.log };
};

describe("gavetas (fumaça em Node)", () => {
  it("camarote: cartas no estado da mesa, diário do ator e rodadas", () => {
    const vm = selectSeatSheet(view("Camarote"), "bruno")!;
    const out = renderToStaticMarkup(<SeatSheetBody vm={vm} />);
    expect(out).toContain("Bruno");
    expect(out).toContain("morta");
    expect(out).toContain("Assassino");
    expect(out).toContain("escondida");
    expect(out).toContain("O que a mesa já viu");
    expect(out).toContain("Roubou 3 de Eva como Trapaceiro.");
    expect(out).toContain("Recebeu 4 como Trilionário.");
    expect(out).toContain("Se disse Assassino, foi contestado e perdeu a carta.");
  });

  it("CamaroteLongo traz as 14 linhas do diário", () => {
    expect(selectSeatSheet(view("CamaroteLongo"), "max")!.log).toHaveLength(14);
  });

  it("seu camarote: cartas, investimento e papéis que a mesa viu", () => {
    const vm = selectYouSheet(view("Voce"))!;
    const out = renderToStaticMarkup(<YouSheetBody vm={vm} />);
    expect(out).toContain("Você");
    expect(out).toContain("Passou de 20 moedas, perde uma carta.");
    expect(out).toContain("a mesa vê: carta protegida");
    expect(out).toContain("a mesa vê: carta virada");
    expect(out).toContain("1 investimento, cada um rende meia moeda por rodada");
    expect(out).toContain("Acaba em 2 rodadas");
    expect(out).toContain("Papéis que a mesa já viu você usar");
    expect(out).toContain("Guardiã");
    expect(out).toContain("rodada 2");
    expect(out).toContain("Protegeu uma carta.");
  });
});
