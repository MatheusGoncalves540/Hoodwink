import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture, type FixtureName } from "../../game/fixtures";
import { selectChat, selectSeatSheet, selectYouSheet, type GameView } from "../../game/selectors";
import { ChatSheetBody } from "./ChatSheet";
import { RoomMenuContent } from "./RoomMenu";
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

describe("menu e conversa (fumaça em Node)", () => {
  it("menu da sala: sala, preferências e saída", () => {
    const out = renderToStaticMarkup(
      <RoomMenuContent
        code="K7Q2"
        turn={4}
        sound
        vibration
        motion={false}
        onSound={() => {}}
        onVibration={() => {}}
        onMotion={() => {}}
        onLeave={() => {}}
      />,
    );
    expect(out).toContain("Sala nº");
    expect(out).toContain("K7Q2");
    expect(out).toContain("rodada 4");
    expect(out).toContain("Como se joga");
    expect(out).toContain("os 12 papéis");
    expect(out).toContain("Som");
    expect(out).toContain("Vibração");
    expect(out).toContain("quando chega a sua deixa");
    expect(out).toContain("Menos movimento");
    expect(out).toContain("cartas e cortinas sem animação");
    expect(out).toContain("Sair de cena");
    expect(out).toContain("Seu camarote fica na mesa e o tempo corre sem você. Para voltar, entre de novo com o código K7Q2.");
    expect(out.match(/aria-checked="true"/g)).toHaveLength(2);
    expect(out.match(/aria-checked="false"/g)).toHaveLength(1);
  });

  it("conversa: falas, horas e acontecimentos", () => {
    const out = renderToStaticMarkup(<ChatSheetBody items={selectChat(view("Conversa"))} />);
    expect(out).toContain('role="log" aria-label="Falas"');
    expect(out).toContain("Conversa da mesa");
    expect(out).toContain("ficam as 30 últimas");
    expect(out).toContain("Carla.");
    expect(out).toContain("Ninguém aqui tem Guardiã, né?");
    expect(out).toContain("21:38");
    expect(out).toContain("Você.");
  });

  it("ConversaCheia entra com a lista cheia", () => {
    const items = selectChat(view("ConversaCheia"));
    expect(items.length).toBeGreaterThanOrEqual(30);
    expect(renderToStaticMarkup(<ChatSheetBody items={items} />)).toContain("kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk");
  });
});
