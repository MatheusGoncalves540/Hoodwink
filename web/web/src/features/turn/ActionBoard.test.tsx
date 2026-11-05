import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture } from "../../game/fixtures";
import { ActionBoard } from "./ActionBoard";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const f = fixture("Acoes", NOW);
const out = renderToStaticMarkup(<ActionBoard view={{ me: f.me, room: f.room, log: f.log }} layout="mobile" onPick={() => {}} />);

describe("ActionBoard", () => {
  it("mostra a rodada, a taxa e suas moedas", () => {
    expect(out).toContain("Sua vez em cena");
    expect(out).toContain("Rodada 4 · Taxa +1");
    expect(out).toContain("Declare um papel");
    expect(out).toContain('aria-label="Você tem 5 moedas"');
    expect(out).toContain('role="progressbar" aria-label="Tempo da jogada"');
  });

  it("usa o preço de pricing.ts e marca o que está na mão", () => {
    expect(out).toContain('aria-label="Assassino, custa 4"');
    expect(out).toContain('aria-label="Trilionário, recebe 4 moedas"');
    expect(out).toContain('aria-label="Kamikaze, custa 0"');
    expect(out).toContain("☞ na mão");
    expect(out).toContain('aria-label="Guardiã, custa 6, faltam 1 moeda"');
  });

  it("Egoísta fica fora do turno, desabilitado, com 1+", () => {
    expect(out).toContain('aria-label="Egoísta, fora do seu turno"');
    expect(out).toContain('aria-disabled="true"');
    expect(out).toContain("1+");
  });
});
