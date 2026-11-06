import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture, type FixtureName } from "../../game/fixtures";
import type { GameView } from "../../game/selectors";
import { WaitingRoom } from "./WaitingRoom";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const view = (name: FixtureName, me?: string): GameView => {
  const f = fixture(name, NOW);
  return { me: me ?? f.me, room: f.room, log: f.log };
};
const html = (v: GameView) => renderToStaticMarkup(<WaitingRoom view={v} />);

describe("Sala de espera (fumaça em Node)", () => {
  it("mostra convite, elenco e regra da casa", () => {
    const out = html(view("Sala"));
    expect(out).toContain('aria-label="Convite"');
    expect(out).toContain("K7Q2");
    expect(out).toContain("Copiar convite");
    expect(out).toContain("Quinta dos mentirosos");
    expect(out).toContain("Jogadores: 3 de 6 lugares ocupados");
    expect(out).toContain("anfitrião, você");
    expect(out).toContain("pronto");
    expect(out).toContain("Lugar vago");
    expect(out).toContain("aguardando");
    expect(out).toContain("Cada jogador começa com 3 moedas.");
    expect(out).toContain("Quem passar de 20 perde uma carta.");
    expect(out).toContain("Abrir as cortinas");
    expect(out).toContain("Só o anfitrião começa, com pelo menos 2 jogadores.");
  });

  it("só o anfitrião com dois atores abre as cortinas", () => {
    expect(html(view("Sala"))).not.toContain('aria-disabled="true"');
    expect(html(view("Sala", "p1"))).toContain('aria-disabled="true"');
  });

  it("SalaCheia usa linhas compactas e dez lugares", () => {
    const out = html(view("SalaCheia"));
    expect(out).toContain("Jogadores: 9 de 10 lugares ocupados");
    expect(out.match(/<li/g)).toHaveLength(10);
    expect(out).toContain("tight");
  });
});
