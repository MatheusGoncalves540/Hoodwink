import { describe, expect, it } from "vitest";
import { fixture, FIXTURE_NAMES, type FixtureName } from "./fixtures";
import type { PublicRoom } from "./protocol";
import {
  selectChat, selectDecision, selectFinale, selectFootlights, selectLog, selectRolesUsed, selectScene,
  selectSeatSheet, selectSeats, selectStage, selectWaitingRoom, selectYou, selectYouSheet, type GameView,
} from "./selectors";

const NOW = Date.UTC(2026, 10, 4, 21, 41);
const view = (name: FixtureName, me?: string): GameView => {
  const f = fixture(name, NOW);
  return { me: me ?? f.me, room: f.room, log: f.log };
};
const withEvent = (v: GameView, type: string, playerId: string, payload: unknown = null): GameView => ({
  ...v,
  room: { ...(v.room as PublicRoom), gameEvent: { type, playerId, expiresAt: NOW, payload } },
});

describe("Mesa6", () => {
  const v = view("Mesa6");

  it("camarotes na ordem de jogo, com estados das miniaturas", () => {
    expect(selectSeats(v)).toEqual([
      { id: "ana", name: "Ana", short: "Ana", coins: 7, cards: [{ state: "revealed", role: "CROUPIER" }, { state: "back" }], inScene: false, out: false, hasMore: true, label: "Ver camarote de Ana, 7 moedas" },
      { id: "carla", name: "Carla", short: "Carla", coins: 11, cards: [{ state: "protected" }, { state: "back" }], inScene: true, out: false, hasMore: true, label: "Ver camarote de Carla, 11 moedas" },
      { id: "bruno", name: "Bruno", short: "Bruno", coins: 2, cards: [{ state: "dead", role: "ASSASSIN" }, { state: "back" }], inScene: false, out: false, hasMore: true, label: "Ver camarote de Bruno, 2 moedas" },
      { id: "davi", name: "Davi", short: "Davi", coins: 0, cards: [{ state: "dead", role: "TRICKSTER" }, { state: "dead", role: "GRAVEDIGGER" }], inScene: false, out: true, hasMore: false, label: "Ver camarote de Davi, fora do jogo" },
      { id: "eva", name: "Eva", short: "Eva", coins: 4, cards: [{ state: "back" }, { state: "back" }], inScene: false, out: false, hasMore: true, label: "Ver camarote de Eva, 4 moedas" },
    ]);
  });

  it("palco, ribalta e faixa Você", () => {
    expect(selectStage(v)).toEqual({ activeId: "carla", activeName: "Carla", fala: "", declared: null });
    expect(selectFootlights(v)).toEqual({
      deck: 12,
      tax: 1,
      suddenDeath: false,
      effects: [
        { role: "POLITICAL", name: "Político", times: 1, rounds: 3, label: "Político custa 2 vezes mais; baixa em 3 rodadas" },
        { role: "REBEL", name: "Rebelde", times: 2, rounds: 1, label: "Rebelde custa 4 vezes mais; baixa em 1 rodada" },
      ],
    });
    expect(selectYou(v)).toEqual({
      name: "Matheus",
      coins: 7,
      limit: 20,
      alive: true,
      cards: [
        { role: "GUARDIAN", protected: true, dead: false, index: 0 },
        { role: "TRILLIONAIRE", protected: false, dead: false, index: 1 },
      ],
      investments: [2],
      nextInvestment: 2,
      timer: { expiresAt: NOW + 8000, total: 10, ownerName: "Carla" },
    });
  });

  it("diário do ator, do mais novo ao mais velho", () => {
    expect(selectLog(v, "bruno").map((l) => [l.turn, l.text])).toEqual([
      [3, "Se disse Assassino, foi contestado e perdeu a carta."],
      [2, "Roubou 3 de Eva como Trapaceiro."],
      [1, "Recebeu 4 como Trilionário."],
    ]);
  });
});

describe("selectDecision", () => {
  const base = view("Mesa6");

  it("vez: sua, de outro", () => {
    expect(selectDecision(withEvent(base, "WAITING_FIRST_ACTION", "me"))).toEqual({ kind: "myTurn" });
    expect(selectDecision(base)).toEqual({ kind: "none" });
  });

  it("declaração: você reage; quem declarou espera", () => {
    expect(selectDecision(view("Main"))).toEqual({ kind: "react", actorId: "ana", role: "ASSASSIN", canContest: true, block: null, targetsMe: true });
    expect(selectDecision(view("Main", "ana"))).toEqual({ kind: "waitOther", actorId: "ana" });
    expect(selectDecision(withEvent(base, "CARD_PLAYED_CLAIRVOYANT", "ana", { targetPlayer: "me" }))).toEqual({
      kind: "react", actorId: "ana", role: "CLAIRVOYANT", canContest: true, block: { role: "GUARDIAN", cost: 6, payload: "guardian" }, targetsMe: true,
    });
  });

  it("contestação: blefe, verdade e espectador", () => {
    expect(selectDecision(view("ContestaBlefe"))).toEqual({ kind: "pickCardToKill", targetId: "ana", role: "ASSASSIN" });
    expect(selectDecision(view("ContestaVerdade"))).toEqual({ kind: "waitPick", chooserId: "ana", role: "ASSASSIN" });
    expect(selectDecision(view("ContestaVerdade", "ana"))).toEqual({ kind: "pickCardToKill", targetId: "me", role: "ASSASSIN" });
    expect(selectDecision(view("ContestaBlefe", "carla"))).toEqual({ kind: "waitOther", actorId: "me" });
  });

  it("carta morta por ataque: alvo pode bloquear com Guardiã; por contestação, não", () => {
    expect(selectDecision(withEvent(base, "CARD_KILLED", "ana", { cause: "ASSASSIN", targetPlayer: "me" }))).toMatchObject({ kind: "react", block: { role: "GUARDIAN" }, canContest: false });
    expect(selectDecision(withEvent(base, "CARD_KILLED", "ana", { cause: "CONTEST_PENALTY", targetPlayer: "me" }))).toEqual({ kind: "none" });
  });

  it("morte causada por você mesmo (Greed) também aceita Guardiã", () => {
    expect(selectDecision(withEvent(base, "CARD_KILLED", "me", { cause: "GREED", targetPlayer: "me" }))).toMatchObject({
      kind: "react", block: { role: "GUARDIAN" }, targetsMe: true,
    });
  });

  it("ganho de outro: Egoísta pode barrar, com o preço de cada caso", () => {
    expect(selectDecision(withEvent(base, "EARN_COINS", "ana"))).toMatchObject({ kind: "react", block: { role: "SELFISH", cost: 1 } });
    expect(selectDecision(withEvent(base, "INVESTMENT", "ana"))).toMatchObject({ block: { cost: 2 } });
    expect(selectDecision(withEvent(base, "CHANGED_CARD", "ana"))).toMatchObject({ block: { cost: 7 } });
    expect(selectDecision(withEvent(base, "EARN_COINS", "me"))).toEqual({ kind: "none" });
  });

  it("Egoísta não aparece em roubo pequeno (lostCoins/2 <= 1)", () => {
    expect(selectDecision(withEvent(base, "STEAL_COINS", "ana", { lostCoins: 3, targetPlayer: "me" }))).toEqual({ kind: "none" });
    expect(selectDecision(withEvent(base, "STEAL_COINS", "ana", { lostCoins: 6, targetPlayer: "me" }))).toMatchObject({
      kind: "react", block: { role: "SELFISH" },
    });
  });

  it("fora de cena, fim e sem sala", () => {
    expect(selectDecision(view("ForaDeCena"))).toEqual({ kind: "outOfScene" });
    expect(selectDecision(view("Fim"))).toEqual({ kind: "finale" });
    expect(selectDecision(view("Lobby"))).toEqual({ kind: "none" });
  });
});

describe("relógio e cena", () => {
  const base = view("Mesa6");

  it("gameEvent nulo no meio da partida continua mesa; sala que não começou fica sala", () => {
    expect(selectScene({ ...base, room: { ...(base.room as PublicRoom), gameEvent: null } })).toBe("table");
    const waiting = view("Sala");
    expect(selectScene({ ...waiting, room: { ...(waiting.room as PublicRoom), gameEvent: null } })).toBe("room");
  });

  it("timer usa a janela do evento (ação 10 s, contestação 6 s, guardiã 8 s)", () => {
    expect(selectYou(withEvent(base, "WAITING_FIRST_ACTION", "me")).timer?.total).toBe(10);
    expect(selectYou(withEvent(base, "CONTEST", "ana")).timer?.total).toBe(6);
    expect(selectYou(withEvent(base, "CARD_PLAYED_GUARDIAN", "ana")).timer?.total).toBe(8);

    const sudden = view("MesaSubita");
    expect(selectYou(withEvent(sudden, "WAITING_FIRST_ACTION", "me")).timer?.total).toBe(5);
    expect(selectYou(withEvent(sudden, "CONTEST", "ana")).timer?.total).toBe(3);
    expect(selectYou(withEvent(sudden, "CARD_PLAYED_GUARDIAN", "ana")).timer?.total).toBe(4);
  });
});

describe("outras telas", () => {
  it("cena de cada fixture", () => {
    expect(Object.fromEntries(FIXTURE_NAMES.map((n) => [n, selectScene(view(n))]))).toMatchObject({
      Lobby: "lobby", Sala: "room", SalaCheia: "room", Mesa6: "table", Main: "table", MesaSubita: "table", Fim: "finale",
    });
  });

  it("palco na contestação", () => {
    expect(selectStage(view("ContestaBlefe")).declared).toEqual({ actorId: "ana", actorName: "Ana", role: "ASSASSIN", phase: "bluff", targetId: null });
    expect(selectStage(view("ContestaVerdade")).declared?.phase).toBe("truth");
    expect(selectStage(view("Main")).declared).toEqual({ actorId: "ana", actorName: "Ana", role: "ASSASSIN", phase: "declared", targetId: "me" });
  });

  it("morte súbita liga com baralho vazio e deixa de 5 s", () => {
    const v = view("MesaSubita");
    expect(selectFootlights(v).suddenDeath).toBe(true);
    expect(selectYou(v).timer?.total).toBe(5);
    expect(selectFootlights(view("Sala")).suddenDeath).toBe(false);
  });

  it("conversa: falas e acontecimentos em ordem, você como 'Você'", () => {
    expect(selectChat(view("Conversa")).map((c) => (c.kind === "line" ? `${c.name}: ${c.text}` : c.text))).toEqual([
      "Carla: Ninguém aqui tem Guardiã, né?",
      "Ana: Esse Crupiê é meu, juro pela cortina.",
      "(Bruno se disse Assassino, foi contestado e perdeu a carta.)",
      "Bruno: Isso foi pessoal.",
      "Você: Foi teatro, Bruno.",
      "Eva: Alguém conta as moedas da Carla, por favor.",
    ]);
    expect(selectChat(view("ConversaCheia"))).toHaveLength(30);
  });

  it("fim: vencedor e ordem de saída do último ao primeiro", () => {
    expect(selectFinale(view("Fim"))).toEqual({
      winnerName: "Carla",
      round: 17,
      coins: 14,
      fala: "",
      exits: [
        { place: 2, name: "Ana", round: 17, me: false },
        { place: 3, name: "Você", round: 14, me: true },
        { place: 4, name: "Eva", round: 12, me: false },
        { place: 5, name: "Bruno", round: 9, me: false },
        { place: 6, name: "Davi", round: 5, me: false },
      ],
    });
  });

  it("sala de espera: lugares vagos e anfitrião", () => {
    const s = selectWaitingRoom(view("Sala"));
    expect(s).toMatchObject({ code: "K7Q2", name: "Quinta dos mentirosos", max: 6, count: 3, castLine: "três de seis papéis preenchidos", isHost: true });
    expect(s.seats.map((x) => [x.name, x.tag])).toEqual([
      ["Matheus", "anfitrião, você"], ["Ana", "pronto"], ["Bruno", "pronto"], ["Papel vago", "aguardando"], ["Papel vago", "aguardando"], ["Papel vago", "aguardando"],
    ]);
    expect(s.tight).toBe(false);
    expect(selectWaitingRoom(view("SalaCheia", "p1")).isHost).toBe(false);
  });

  it("papéis que a mesa já viu você usar", () => {
    expect(selectRolesUsed(view("Voce"), "me").map((r) => [r.name, r.turn, r.text])).toEqual([
      ["Guardiã", 2, "Protegeu uma carta."],
      ["Trilionário", 1, "Recebeu 4 moedas."],
    ]);
  });
});

describe("gavetas", () => {
  it("camarote: cartas, diário do ator e investimentos", () => {
    const seat = selectSeatSheet(view("Camarote"), "bruno")!;
    expect(seat).toMatchObject({ id: "bruno", name: "Bruno", coins: 2, alive: true, investments: [] });
    expect(seat.cards).toEqual([{ state: "dead", role: "ASSASSIN" }, { state: "back" }]);
    expect(seat.log.map((l) => [l.turn, l.text])).toEqual([
      [3, "Se disse Assassino, foi contestado e perdeu a carta."],
      [2, "Roubou 3 de Eva como Trapaceiro."],
      [1, "Recebeu 4 como Trilionário."],
    ]);
  });

  it("seu camarote: cartas, limite e papéis que a mesa viu", () => {
    const you = selectYouSheet(view("Voce"))!;
    expect(you).toMatchObject({ name: "Você", coins: 7, limit: 20, investments: [2] });
    expect(you.cards).toEqual([
      { role: "GUARDIAN", protected: true, dead: false, index: 0 },
      { role: "TRILLIONAIRE", protected: false, dead: false, index: 1 },
    ]);
    expect(you.roles.map((r) => [r.name, r.turn, r.text])).toEqual([
      ["Guardiã", 2, "Protegeu uma carta."],
      ["Trilionário", 1, "Recebeu 4 moedas."],
    ]);
  });
});
