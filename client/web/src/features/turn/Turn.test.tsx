import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture, type FixtureName } from "../../game/fixtures";
import { selectDecision, type Decision, type GameView } from "../../game/selectors";
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

  it("Clarividente escolhe se a carta aparece para todos; os outros não têm a chave", () => {
    expect(html(<TargetBar role="CLAIRVOYANT" onBack={() => {}} showToAll={false} onShowToAll={() => {}} />)).toContain(
      'role="switch" aria-checked="false"',
    );
    expect(html(<TargetBar role="KILL_CARD" onBack={() => {}} />)).not.toContain('role="switch"');
  });
});
