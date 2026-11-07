import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture, type FixtureName } from "../../game/fixtures";
import { selectDecision, type Decision, type GameView } from "../../game/selectors";
import { play } from "../../net/actions";
import { Table } from "../table/Table";
import { CenterCurtain } from "./CenterCurtain";
import { RevealChoice } from "./RevealChoice";
import { ContestPick } from "./ContestPick";
import { DeclaredCard } from "./DeclaredCard";
import { ReactionPanel } from "./ReactionPanel";
import { TargetBar } from "./TargetBar";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const view = (name: FixtureName): GameView => {
  const f = fixture(name, NOW);
  return { me: f.me, room: f.room, log: f.log };
};
const html = (node: React.ReactElement) => renderToStaticMarkup(node);

describe("reação e contestação", () => {
  it("Main: palco com a carta declarada e a sua deixa", () => {
    const v = view("Main");
    const decision = selectDecision(v);
    if (decision.kind !== "react") throw new Error("Main deveria reagir");
    const out = html(
      <>
        <DeclaredCard view={v} layout="mobile" />
        <ReactionPanel view={v} decision={decision} layout="mobile" onContest={() => {}} onBlock={() => {}} onPass={() => {}} />
      </>,
    );
    expect(out).toContain("Ana se diz");
    expect(out).toContain(">Assassino<");
    expect(out).toContain("Ana paga 4 moedas e mira uma carta sua.");
    expect(out).toContain("dita, não mostrada");
    expect(out).toContain("Sua vez de reagir");
    expect(out).toContain("Na mão: Guardiã e Trilionário");
    expect(out).toContain("Duvidar");
    expect(out).toContain("Deixar passar");
    expect(out).toContain('role="progressbar" aria-label="Tempo para reagir"');
  });

  it("ContestaBlefe: carimbo e a escolha da carta que cai", () => {
    const v = view("ContestaBlefe");
    const decision = selectDecision(v);
    if (decision.kind !== "pickCardToKill") throw new Error("ContestaBlefe deveria escolher carta");
    const out = html(
      <>
        <DeclaredCard view={v} layout="mobile" />
        <ContestPick view={v} decision={decision} layout="mobile" onPick={() => {}} />
      </>,
    );
    expect(out).toContain("Ana se disse");
    expect(out).toContain("Blefe?");
    expect(out).toContain("Ana não tinha Assassino.");
    expect(out).toContain("Derrube uma carta de Ana");
    expect(out).toContain('aria-label="Derrubar a carta revelada de Ana, Crupiê"');
    expect(out).toContain('aria-label="Derrubar a carta escondida de Ana"');
    expect(out).toContain("a sorte escolhe por você.");
  });

  it("ContestaVerdade: verdade e a escolha do outro sobre as suas cartas", () => {
    const v = view("ContestaVerdade");
    const decision = selectDecision(v);
    if (decision.kind !== "waitPick") throw new Error("ContestaVerdade deveria esperar a escolha");
    const out = html(
      <>
        <DeclaredCard view={v} layout="mobile" />
        <ContestPick view={v} decision={decision} layout="mobile" onPick={() => {}} />
      </>,
    );
    expect(out).toContain("Ana era mesmo");
    expect(out).toContain("Verdade");
    expect(out).toContain("Ana escolhe qual carta sua cai");
    expect(out).toContain("Guardiã");
    expect(out).toContain("se for esta, cai só a proteção");
    // a carta protegida mostra o estado do servidor, não “na mão”
    expect(out).toContain(">protegida<");
    expect(out).toContain(">na mão<");
    expect(out).toContain("Depois, a sorte escolhe.");
  });
});

describe("carta revelada", () => {
  it("Clarividente: o palco mostra a carta e o alvo", () => {
    const base = view("Mesa6");
    const v: GameView = {
      ...base,
      room: {
        ...base.room!,
        gameEvent: {
          type: "REVEALED_CARD",
          playerId: "ana",
          expiresAt: NOW + 6000,
          payload: { targetPlayer: "bruno", targetCardIndex: 0, showToAllPlayers: true, revealedCard: "KAMIKAZE" },
        },
      },
    };
    const out = html(<DeclaredCard view={v} layout="mobile" />);
    expect(out).toContain("Ana revelou uma carta de Bruno.");
    expect(out).toContain("Kamikaze");
    expect(out).toContain("A mesa toda viu a carta.");
  });
});

describe("reação com bloqueio", () => {
  it("Clarividente: a Guardiã aparece com o preço da taxa", () => {
    const base = view("Mesa6");
    const v: GameView = {
      ...base,
      room: {
        ...base.room!,
        gameEvent: { type: "CARD_PLAYED_CLAIRVOYANT", playerId: "ana", expiresAt: NOW + 8000, payload: { targetPlayer: "me", targetCardIndex: 1 } },
      },
    };
    const decision = selectDecision(v);
    if (decision.kind !== "react" || !decision.block) throw new Error("deveria oferecer a Guardiã");
    const out = html(
      <ReactionPanel view={v} decision={decision} layout="mobile" onContest={() => {}} onBlock={() => {}} onPass={() => {}} />,
    );
    expect(out).toContain("Bloquear com Guardiã");
    expect(out).toContain("por 5 moedas");
  });
});

describe("barra de alvo (regras v2)", () => {
  it("Trocar cartas oferece trocar as duas, com o aviso de custo", () => {
    const out = html(<TargetBar role="SWAP_CARDS" onBack={() => {}} onSwapBoth={() => {}} />);
    expect(out).toContain("Trocar as duas");
    expect(out).toContain("Custa todas as suas moedas.");
    expect(out).toContain("Escolha a carta sua que vai trocar.");
  });

  it("os outros papéis não têm a pergunta de revelar", () => {
    expect(html(<TargetBar role="KILL_CARD" onBack={() => {}} />)).not.toContain("Revelar para");
  });
});

describe("escolha do Clarividente", () => {
  it("oferece revelar para todos ou só para você, sem escolha pronta", () => {
    const out = html(<RevealChoice onPick={() => {}} onBack={() => {}} />);
    expect(out).toContain('role="dialog"');
    expect(out).toContain("Revelar para quem");
    expect(out).toContain(">Todos<");
    expect(out).toContain(">Apenas você<");
    expect(out).toContain("A mesa toda vê a carta.");
    expect(out).toContain("Só você vê a carta.");
    expect(out).not.toContain("aria-pressed");
    // o valor escolhido vai no payload até a jogada confirmar
    expect(play("CLAIRVOYANT", "me", { playerId: "bruno", cardIndex: 0, showToAll: true }).payload).toMatchObject({
      showToAllPlayers: true,
    });
  });

  it("depois da escolha, a mesa dos alvos fica visível com o que foi decidido", () => {
    const out = html(
      <Table
        view={view("Mesa6")}
        targeting
        slots={{
          curtain: (
            <CenterCurtain open>
              <TargetBar role="CLAIRVOYANT" onBack={() => {}} note="A carta vai aparecer para todos." />
            </CenterCurtain>
          ),
        }}
      />,
    );
    expect(out).toContain("Escolha o alvo");
    expect(out).toContain("A carta vai aparecer para todos.");
    // HUD continua: a faixa Você está na tela (a barra de alvo para acima dela)
    expect(out).toContain('aria-label="Você"');
    // durante o alvo a cortina não recolhe: sem handle
    expect(out).not.toContain("Fechar ações");
  });

  it("o holofote global traz a pergunta por cima da mesa e da faixa Você, sem as poças de alvo", () => {
    const out = html(<Table view={view("Mesa6")} slots={{ center: <RevealChoice onPick={() => {}} onBack={() => {}} /> }} />);
    expect(out).not.toContain('data-slot="target-light"');
    expect(out).toContain('data-slot="global-question"');
    expect(out).toContain('data-lugar="pergunta"');
    expect(out).toContain("Revelar para quem");
    expect(out).toContain("Voltar às ações");
    // o holofote cobre a faixa Você: vem depois dela
    expect(out.indexOf('data-slot="center"')).toBeGreaterThan(out.indexOf('aria-label="Você"'));
  });
});
