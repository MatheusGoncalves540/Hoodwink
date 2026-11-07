import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture, type FixtureName } from "../../game/fixtures";
import { selectDecision, selectFinale, type GameView } from "../../game/selectors";
import { Finale, type FinaleReplay } from "./Finale";
import { OutOfScene } from "./OutOfScene";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const view = (name: FixtureName): GameView => {
  const f = fixture(name, NOW);
  return { me: f.me, room: f.room, log: f.log };
};

const replay = (patch: Partial<FinaleReplay> = {}): FinaleReplay => ({
  ready: false,
  checking: false,
  pending: false,
  unauthorized: false,
  downloading: false,
  error: "",
  onWatch() {},
  onDownload() {},
  onRetry() {},
  onLogin() {},
  ...patch,
});

describe("fim de jogo", () => {
  it("ForaDeCena: cartas mortas, rodada da saída e as duas ações", () => {
    const v = view("ForaDeCena");
    expect(selectDecision(v)).toEqual({ kind: "outOfScene" });
    const out = renderToStaticMarkup(<OutOfScene view={v} layout="mobile" onChat={() => {}} onLeave={() => {}} />);
    expect(out).toContain("Fora da partida");
    expect(out).toContain("Suas cartas caíram na rodada 14.");
    expect(out).toContain('aria-label="Suas duas cartas, mortas: Guardiã e Trilionário"');
    expect(out).toContain("Abrir a conversa");
    expect(out).toContain("Sair da partida");
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
    expect(out).toContain("Jogar novamente");
    expect(out).toContain("☜ Voltar à entrada");
  });

  it("Fim: com replay pronto mostra Assistir e Baixar habilitados", () => {
    const vm = selectFinale(view("Fim"));
    const out = renderToStaticMarkup(
      <Finale vm={vm} layout="mobile" onNewRoom={() => {}} onLobby={() => {}} replay={replay({ ready: true })} />,
    );

    expect(out).toContain("Replay da partida");
    expect(out).toContain(">Assistir<");
    expect(out).toContain(">Baixar<");
    expect(out).not.toContain('aria-disabled="true"');
  });

  it("Fim: enquanto procura o replay, os botões ficam desabilitados", () => {
    const vm = selectFinale(view("Fim"));
    const out = renderToStaticMarkup(
      <Finale vm={vm} layout="mobile" onNewRoom={() => {}} onLobby={() => {}} replay={replay({ checking: true })} />,
    );

    expect(out).toContain("Procurando o replay…");
    expect(out.match(/aria-disabled="true"/g)?.length).toBe(2);
  });

  it("Fim: replay demorou mostra o aviso e deixa tentar de novo", () => {
    const vm = selectFinale(view("Fim"));
    const out = renderToStaticMarkup(
      <Finale vm={vm} layout="mobile" onNewRoom={() => {}} onLobby={() => {}} replay={replay({ pending: true })} />,
    );

    expect(out).toContain("O replay ainda está sendo processado.");
    expect(out).toContain("Tentar de novo");
  });

  it("Fim: sem login, o cartaz convida a entrar", () => {
    const vm = selectFinale(view("Fim"));
    const out = renderToStaticMarkup(
      <Finale vm={vm} layout="mobile" onNewRoom={() => {}} onLobby={() => {}} replay={replay({ unauthorized: true })} />,
    );

    expect(out).toContain("Entre na sua conta para ver seus replays.");
    expect(out).toMatch(/>Entrar</);
  });

  it("Fim: erro de download aparece sem esconder as ações", () => {
    const vm = selectFinale(view("Fim"));
    const out = renderToStaticMarkup(
      <Finale
        vm={vm}
        layout="mobile"
        onNewRoom={() => {}}
        onLobby={() => {}}
        replay={replay({ ready: true, error: "Não foi possível baixar o replay. Tente de novo." })}
      />,
    );

    expect(out).toContain('role="alert"');
    expect(out).toContain("Não foi possível baixar o replay. Tente de novo.");
    expect(out).toContain(">Assistir<");
  });
});
