import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture, type FixtureName } from "../../game/fixtures";
import type { GameView } from "../../game/selectors";
import type { PublicCard } from "../../game/protocol";
import { CenterCurtain } from "../turn/CenterCurtain";
import { Table, ownCardDisabled } from "./Table";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const view = (name: FixtureName): GameView => {
  const f = fixture(name, NOW);
  return { me: f.me, room: f.room, log: f.log };
};
const html = (name: FixtureName) => renderToStaticMarkup(<Table view={view(name)} unread={2} />);

describe("Mesa (fumaça em Node)", () => {
  it("Mesa6 desenha cabeçalho, camarotes, palco, ribalta e faixa Você", () => {
    const out = html("Mesa6");
    expect(out).toContain('aria-label="Lugares na mesa, toque para ver detalhes"');
    expect(out).toContain('aria-label="Ana, 7 moedas. Ver o lugar"');
    expect(out).toContain('aria-label="Davi, fora do jogo. Ver o lugar"');
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

  it("o slot curtain vive dentro da área da mesa, fora do cabeçalho e da faixa Você", () => {
    const out = renderToStaticMarkup(<Table view={view("Mesa6")} slots={{ curtain: <p>ações na cortina</p> }} />);
    const area = out.indexOf('data-slot="table-area"');
    const curtain = out.indexOf('data-slot="curtain"');
    expect(area).toBeGreaterThan(-1);
    expect(curtain).toBeGreaterThan(area);
    expect(out).toContain("ações na cortina");
    // cabeçalho vem antes da área; a faixa Você (HUD) fica depois da cortina
    expect(out.indexOf("Rodada 4 · Taxa +1")).toBeLessThan(area);
    expect(out.indexOf('aria-label="Você"')).toBeGreaterThan(curtain);
  });

  it("o aviso de tela inteira fica fora da área da mesa e da faixa Você", () => {
    const out = renderToStaticMarkup(
      <Table view={view("Mesa6")} dark slots={{ curtain: <p>ações na cortina</p>, screenOverlay: <p>aviso de morte súbita</p> }} />,
    );
    const overlay = out.indexOf('data-slot="screen-overlay"');
    expect(overlay).toBeGreaterThan(out.indexOf('data-slot="curtain"'));
    expect(overlay).toBeGreaterThan(out.indexOf('aria-label="Você"'));
    expect(out).toContain("aviso de morte súbita");
  });

  it("com a cortina fechada sobra o handle e a faixa Você continua no lugar", () => {
    const out = renderToStaticMarkup(
      <Table
        view={view("Mesa6")}
        slots={{
          curtain: (
            <CenterCurtain open={false} onToggle={() => {}}>
              <p>ações na cortina</p>
            </CenterCurtain>
          ),
        }}
      />,
    );
    expect(out).toContain('aria-expanded="false"');
    expect(out).toContain("Abrir ações");
    expect(out).toContain('aria-label="Você"');
    expect(out.indexOf('aria-label="Você"')).toBeGreaterThan(out.indexOf('data-slot="curtain"'));
  });

  it("Mesa3 e Mesa10 mudam o número de camarotes", () => {
    expect(html("Mesa3").match(/Ver o lugar/g)).toHaveLength(2);
    const out10 = html("Mesa10");
    expect(out10.match(/Ver o lugar/g)).toHaveLength(9);
    expect(out10).toContain("Iracema");
  });

  it("Crupê aceita carta morta; Coveiro só morta; Guardiã só viva e sem proteção", () => {
    const card = (dead: boolean, protectedAlso = false): PublicCard => ({
      name: "ASSASSIN",
      index: 0,
      protected: protectedAlso,
      dead,
    });
    expect(ownCardDisabled("CROUPIER", card(true))).toBe(false);
    expect(ownCardDisabled("CROUPIER", card(false))).toBe(false);
    expect(ownCardDisabled("GRAVEDIGGER", card(true))).toBe(false);
    expect(ownCardDisabled("GRAVEDIGGER", card(false))).toBe(true);
    expect(ownCardDisabled("GUARDIAN", card(true))).toBe(true);
    expect(ownCardDisabled("GUARDIAN", card(false, true))).toBe(true);
    expect(ownCardDisabled("GUARDIAN", card(false))).toBe(false);
    expect(ownCardDisabled("SWAP_CARDS", card(true))).toBe(true);
  });
});
