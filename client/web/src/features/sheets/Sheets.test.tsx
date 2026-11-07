import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture, type FixtureName } from "../../game/fixtures";
import type { LogEntry } from "../../game/log";
import { selectChat, selectSeatSheet, selectYouSheet, type GameView } from "../../game/selectors";
import { ChatSheetBody } from "./ChatSheet";
import { MotionToggles, RoomMenuContent } from "./RoomMenu";
import { SeatSheetBody } from "./SeatSheet";
import { YouSheetBody } from "./YouSheet";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const view = (name: FixtureName): GameView => {
  const f = fixture(name, NOW);
  return { me: f.me, room: f.room, log: f.log };
};

/** Entrada do diário como appendLog grava o REVEALED_CARD (alvo bruno, carta 1). */
const revealEntry = (showToAll: boolean): LogEntry => ({
  key: `REVEALED_CARD|ana|bruno|1|${showToAll}`,
  turn: 5,
  at: NOW,
  actorId: "ana",
  actorName: "Ana",
  pred: "revelou uma carta de Bruno: Kamikaze como Clarividente.",
  reveal: { targetId: "bruno", cardIndex: 1, card: "KAMIKAZE", showToAll },
});

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
    expect(out).toContain("Se disse Assassino, alguém duvidou e perdeu a carta.");
  });

  it("revelação pública: selo no alvo e linha no o que a mesa já viu", () => {
    const base = view("Mesa6");
    const entry = revealEntry(true);
    const vm = selectSeatSheet({ ...base, log: [...base.log, entry] }, "bruno")!;
    const out = renderToStaticMarkup(<SeatSheetBody vm={vm} />);
    expect(out).toContain('role="note"');
    expect(out).toContain("Carta revelada no turno 5");
    expect(out).toContain(">Kamikaze<");
    expect(out).toContain("Teve a carta Kamikaze revelada para todos.");
    // a memória é do diário: a carta continua escondida na fileira
    expect(out).toContain("escondida");
  });

  it("revelação privada não aparece na gaveta do alvo", () => {
    const base = view("Mesa6");
    const vm = selectSeatSheet({ ...base, log: [...base.log, revealEntry(false)] }, "bruno")!;
    const out = renderToStaticMarkup(<SeatSheetBody vm={vm} />);
    expect(out).not.toContain("Carta revelada no turno");
    expect(out).not.toContain("revelada para todos");
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
    expect(out).toContain("Personagens que a mesa já viu você usar");
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
        motion={{ all: true, idle: false }}
        onSound={() => {}}
        onVibration={() => {}}
        onMotion={() => {}}
        onIdle={() => {}}
        onLeave={() => {}}
      />,
    );
    expect(out).toContain("Sala nº");
    expect(out).toContain("K7Q2");
    expect(out).toContain("rodada 4");
    expect(out).toContain("Como se joga");
    expect(out).toContain("os 12 personagens");
    expect(out).toContain("Som");
    expect(out).toContain("Vibração");
    expect(out).toContain("quando chega a sua vez");
    expect(out).toContain("Animações");
    expect(out).toContain("Todas, inclusive as transições.");
    expect(out).toContain("Animações de espera");
    expect(out).toContain("Desenhos tremendo, texto respirando, lâmpadas.");
    expect(out).toContain("Sair da partida");
    expect(out).toContain("Seu lugar fica na mesa. Para voltar, use o código K7Q2.");
    expect(out.match(/aria-checked="true"/g)).toHaveLength(3);
    expect(out.match(/aria-checked="false"/g)).toHaveLength(1);
  });

  it("sem Animações, a chave de espera aparece desligada e travada, com a dica", () => {
    const out = renderToStaticMarkup(<MotionToggles motion={{ all: false, idle: false }} onMotion={() => {}} onIdle={() => {}} />);
    expect(out).toContain('role="switch" aria-checked="false" aria-disabled="true"');
    expect(out).toContain("Desligada junto com Animações");
    expect(out.match(/aria-checked="false"/g)).toHaveLength(2);
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
