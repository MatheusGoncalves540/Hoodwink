import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture, type FixtureName } from "../../game/fixtures";
import { selectDecision, selectFinale, type GameView } from "../../game/selectors";
import { Finale } from "./Finale";
import { OutOfScene } from "./OutOfScene";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const view = (name: FixtureName): GameView => {
  const f = fixture(name, NOW);
  return { me: f.me, room: f.room, log: f.log };
};

describe("fim de jogo", () => {
  it("ForaDeCena: cartas mortas, rodada da saída e as duas ações", () => {
    const v = view("ForaDeCena");
    expect(selectDecision(v)).toEqual({ kind: "outOfScene" });
    const out = renderToStaticMarkup(<OutOfScene view={v} layout="mobile" onChat={() => {}} onLeave={() => {}} />);
    expect(out).toContain("Fora da partida");
    expect(out).toContain("Suas cartas caíram na rodada 14.");
    expect(out).toContain('aria-label="Suas duas cartas, mortas: Guardiã e Trilionário"');
    expect(out).toContain("Abrir a conversa");
    expect(out).toContain("Sair da sala");
  });

  it("Fim: vencedor, ordem de saída e as ações", () => {
    const vm = selectFinale(view("Fim"));
    const out = renderToStaticMarkup(<Finale vm={vm} layout="mobile" onNewRoom={() => {}} onLobby={() => {}} />);
    expect(out).toContain("Fim da partida");
    expect(out).toContain(">Carla<");
    expect(out).toContain("Venceu na rodada 17, com 14 moedas.");
    expect(out).toContain("Ordem de saída");
    expect(out).toContain("2º");
    expect(out).toContain("rodada 17");
    expect(out).toContain("Você");
    expect(out).toContain("Criar outra sala");
    expect(out).toContain("☜ Voltar à entrada");
  });
});
