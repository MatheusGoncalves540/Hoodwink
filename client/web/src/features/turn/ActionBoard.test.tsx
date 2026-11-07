import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture } from "../../game/fixtures";
import { ActionBoard } from "./ActionBoard";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const f = fixture("Acoes", NOW);
const out = renderToStaticMarkup(<ActionBoard view={{ me: f.me, room: f.room, log: f.log }} layout="mobile" onPick={() => {}} />);

describe("ActionBoard", () => {
  it("mostra a rodada, a taxa e suas moedas", () => {
    expect(out).toContain("Sua vez");
    expect(out).toContain("Rodada 4 · Taxa +1");
    expect(out).toContain("Escolha uma ação");
    expect(out).toContain('aria-label="Você tem 5 moedas"');
    expect(out).toContain('role="progressbar" aria-label="Tempo da jogada"');
  });

  it("usa o preço de pricing.ts e marca o que está na mão", () => {
    expect(out).toContain('aria-label="Assassino, custa 4"');
    expect(out).toContain('aria-label="Trilionário, recebe 4 moedas"');
    // sem Kamikaze vivo na mão o servidor recusaria o sacrifício: botão desligado com o motivo
    expect(out).toContain('aria-label="Kamikaze, Precisa de um Kamikaze vivo na mão"');
    expect(out).toContain("☞ na mão");
    expect(out).toContain('aria-label="Guardiã, custa 7, faltam 2 moedas"');
  });

  it("ações do turno no topo, com o mesmo botão: Pegar moeda, Matar, Trocar cartas", () => {
    const i = out.indexOf("Ações do turno");
    expect(i).toBeGreaterThan(-1);
    expect(i).toBeLessThan(out.indexOf("Para ganhar moedas"));
    expect(out).toContain('aria-label="Pegar moeda, recebe 2 moedas"'); // taxa +1
    expect(out).toContain('aria-label="Matar, custa 6, faltam 1 moeda"');
    expect(out).toContain('aria-label="Trocar cartas, Precisa de 6 moedas"');
  });

  it("Egoísta fica fora do turno, desabilitado, com 1+", () => {
    expect(out).toContain('aria-label="Egoísta, fora do seu turno"');
    expect(out).toContain('aria-disabled="true"');
    expect(out).toContain("1+");
  });

  it("liga o motivo ao botão desabilitado pelo aria-describedby", () => {
    expect(out).toContain("aria-describedby=");
    expect(out).toContain('aria-live="polite"');
  });
});
