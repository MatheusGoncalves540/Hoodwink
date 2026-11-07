import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture } from "../../game/fixtures";
import { ActionBoard, ActionPeek } from "./ActionBoard";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const f = fixture("Acoes", NOW);
const out = renderToStaticMarkup(<ActionBoard view={{ me: f.me, room: f.room, log: f.log }} layout="mobile" onPick={() => {}} />);

describe("ActionBoard", () => {
  it("o cabeçalho da gaveta (handle) mostra o título, suas moedas e o tempo da jogada", () => {
    const peek = renderToStaticMarkup(<ActionPeek view={{ me: f.me, room: f.room, log: f.log }} open />);
    expect(peek).toContain("Escolha uma ação");
    expect(peek).toContain("Ver palco");
    expect(peek).toContain('aria-label="Você tem 5 moedas"');
    expect(peek).toContain('role="progressbar" aria-label="Tempo da jogada"');
    expect(renderToStaticMarkup(<ActionPeek view={{ me: f.me, room: f.room, log: f.log }} open={false} />)).toContain("Abrir");
  });

  it("Na sua mão vem no topo, com a carta ao lado, e sai dos grupos", () => {
    const hand = out.indexOf("Na sua mão");
    expect(hand).toBeGreaterThan(-1);
    expect(hand).toBeLessThan(out.indexOf("Ações do turno"));
    expect(out).toContain("Jogar estas não é blefe.");
    // Trilionário e Guardiã estão na mão: aparecem uma vez só, antes dos grupos
    expect(out.split('aria-label="Trilionário, recebe 4 moedas"')).toHaveLength(2);
    expect(out.indexOf('aria-label="Trilionário, recebe 4 moedas"')).toBeLessThan(out.indexOf("Ações do turno"));
    expect(out).toContain('class="hw-carta');
  });

  it("usa o preço de pricing.ts e marca o que está na mão", () => {
    expect(out).toContain('aria-label="Assassino, custa 5"');
    expect(out).toContain('aria-label="Trilionário, recebe 4 moedas"');
    // sem Kamikaze vivo na mão o servidor recusaria o sacrifício: botão desligado com o motivo
    expect(out).toContain('aria-label="Kamikaze, Precisa de um Kamikaze vivo na mão"');
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

  it("mostra a fileira só de leitura com suas cartas", () => {
    expect(out).toContain('aria-label="Suas cartas"');
    expect(out).toContain('data-slot="hand"');
    // a protegida aparece com o estado do servidor, e nada aqui é seletor: quem escolhe é a gaveta Você
    expect(out).toContain(">protegida<");
    expect(out).not.toContain("aria-label=\"Escolher Guardiã\"");
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
