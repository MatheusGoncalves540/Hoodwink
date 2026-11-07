import { describe, expect, it } from "vitest";
import { fixture, FIXTURE_NAMES, type FixtureName } from "./fixtures";
import type { LogEntry } from "./log";
import type { PublicRoom, RoleId } from "./protocol";
import {
  describeEvent, kamikazeAlly, selectChat, selectDecision, selectFinale, selectFootlights, selectLog, selectRolesUsed, selectScene,
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
      { id: "ana", name: "Ana", short: "Ana", coins: 7, cards: [{ state: "revealed", role: "CROUPIER", index: 0 }, { state: "back", index: 1 }], inScene: false, out: false, hasMore: true, label: "Ana, 7 moedas. Ver o lugar" },
      { id: "carla", name: "Carla", short: "Carla", coins: 11, cards: [{ state: "protected", index: 0 }, { state: "back", index: 1 }], inScene: true, out: false, hasMore: true, label: "Carla, 11 moedas. Ver o lugar" },
      { id: "bruno", name: "Bruno", short: "Bruno", coins: 2, cards: [{ state: "dead", role: "ASSASSIN", index: 0 }, { state: "back", index: 1 }], inScene: false, out: false, hasMore: true, label: "Bruno, 2 moedas. Ver o lugar" },
      { id: "davi", name: "Davi", short: "Davi", coins: 0, cards: [{ state: "dead", role: "TRICKSTER", index: 0 }, { state: "dead", role: "GRAVEDIGGER", index: 1 }], inScene: false, out: true, hasMore: false, label: "Davi, fora do jogo. Ver o lugar" },
      { id: "eva", name: "Eva", short: "Eva", coins: 4, cards: [{ state: "back", index: 0 }, { state: "back", index: 1 }], inScene: false, out: false, hasMore: true, label: "Eva, 4 moedas. Ver o lugar" },
    ]);
  });

  it("palco, ribalta e faixa Você", () => {
    expect(selectStage(v)).toEqual({ activeId: "carla", activeName: "Carla", fala: "Vez de Carla. Escolha uma ação.", declared: null, revealed: null });
    expect(selectFootlights(v)).toEqual({
      deck: 12,
      tax: 1,
      suddenDeath: false,
      effects: [
        { role: "POLITICAL", name: "Político", times: 1, rounds: 3, label: "Político custa 2 vezes mais. Baixa em 3 rodadas" },
        { role: "REBEL", name: "Rebelde", times: 2, rounds: 1, label: "Rebelde custa 4 vezes mais. Baixa em 1 rodada" },
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
      timer: { expiresAt: NOW + 8000, total: 10, ownerName: "Carla", caption: "Vez de Carla. Escolha uma ação." },
    });
  });

  it("diário do ator, do mais novo ao mais velho", () => {
    expect(selectLog(v, "bruno").map((l) => [l.turn, l.text])).toEqual([
      [3, "Se disse Assassino, alguém duvidou e perdeu a carta."],
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
      kind: "react", actorId: "ana", role: "CLAIRVOYANT", canContest: true, block: { role: "GUARDIAN", cost: 5, payload: "guardian" }, targetsMe: true,
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

  it("morte por ganância (Greed) não aceita Guardiã: o servidor recusa (guardianValidation.go)", () => {
    expect(selectDecision(withEvent(base, "CARD_KILLED", "me", { cause: "GREED", targetPlayer: "me" }))).toEqual({ kind: "none" });
  });

  describe("Kamikaze", () => {
    const hand = (v: GameView, cards: { name?: RoleId; dead?: boolean; protected?: boolean }[]): GameView => {
      const r = v.room as PublicRoom;
      const me = r.players.me;
      return { ...v, room: { ...r, players: { ...r.players, me: { ...me, cards: cards.map((c, index) => ({ index, dead: false, protected: false, ...c })) } } } };
    };
    const kami = hand(base, [{ name: "GUARDIAN" }, { name: "KAMIKAZE" }]);

    it("sua carta Kamikaze morrendo: pode levar uma junto, e ainda bloquear se o ataque for bloqueável", () => {
      expect(selectDecision(withEvent(kami, "CARD_KILLED", "ana", { cause: "ASSASSIN", targetPlayer: "me", targetCardIndex: 1 }))).toMatchObject({
        kind: "react", kamikaze: true, block: { role: "GUARDIAN" },
      });
      // contestação não se bloqueia, mas o Kamikaze revida
      expect(selectDecision(withEvent(kami, "CARD_KILLED", "ana", { cause: "CONTEST_PENALTY", targetPlayer: "me", targetCardIndex: 1 }))).toMatchObject({
        kind: "react", kamikaze: true, block: null,
      });
    });

    it("outra carta sua morrendo não aciona o Kamikaze", () => {
      expect(selectDecision(withEvent(kami, "CARD_KILLED", "ana", { cause: "ASSASSIN", targetPlayer: "me", targetCardIndex: 0 }))).not.toHaveProperty("kamikaze");
    });

    it("Kamikaze protegido não morre de Assassino, então não revida", () => {
      const prot = hand(base, [{ name: "GUARDIAN" }, { name: "KAMIKAZE", protected: true }]);
      expect(selectDecision(withEvent(prot, "CARD_KILLED", "ana", { cause: "ASSASSIN", targetPlayer: "me", targetCardIndex: 1 }))).toEqual({ kind: "none" });
    });

    it("kamikazeAlly: o Kamikaze que morre na reação; no turno, o primeiro vivo; sem Kamikaze, null", () => {
      expect(kamikazeAlly(withEvent(kami, "CARD_KILLED", "ana", { cause: "ASSASSIN", targetPlayer: "me", targetCardIndex: 1 }))).toBe(1);
      expect(kamikazeAlly(withEvent(kami, "CARD_KILLED", "ana", { cause: "ASSASSIN", targetPlayer: "me", targetCardIndex: 0 }))).toBeNull();
      expect(kamikazeAlly(withEvent(kami, "WAITING_FIRST_ACTION", "me"))).toBe(1);
      expect(kamikazeAlly(withEvent(hand(base, [{ name: "KAMIKAZE", dead: true }, { name: "REBEL" }]), "WAITING_FIRST_ACTION", "me"))).toBeNull();
    });
  });

  it("ganho de outro: Egoísta pode barrar, com o preço de cada caso", () => {
    expect(selectDecision(withEvent(base, "EARN_COINS", "ana"))).toMatchObject({ kind: "react", block: { role: "SELFISH", cost: 1 } });
    expect(selectDecision(withEvent(base, "INVESTMENT", "ana"))).toMatchObject({ block: { cost: 1 } });
    expect(selectDecision(withEvent(base, "CHANGED_CARD", "ana"))).toMatchObject({ block: { cost: 6 } });
    expect(selectDecision(withEvent(base, "EARN_COINS", "me"))).toEqual({ kind: "none" });
  });

  it("Egoísta não aparece em roubo pequeno (ganho ⌈lostCoins/2⌉ <= 1) nem em ganho abaixo de 2 + t", () => {
    expect(selectDecision(withEvent(base, "STEAL_COINS", "ana", { lostCoins: 2, targetPlayer: "me" }))).toEqual({ kind: "none" });
    expect(selectDecision(withEvent(base, "STEAL_COINS", "ana", { lostCoins: 3, targetPlayer: "me" }))).toMatchObject({ block: { role: "SELFISH" } });
    expect(selectDecision(withEvent(base, "EARN_COINS", "ana", { earnedCoins: 2 }))).toEqual({ kind: "none" }); // taxa +1: limiar 3
    expect(selectDecision(withEvent(base, "EARN_COINS", "ana", { earnedCoins: 3 }))).toMatchObject({ block: { role: "SELFISH", cost: 1 } });
    expect(selectDecision(withEvent(base, "STEAL_COINS", "ana", { lostCoins: 6, targetPlayer: "me" }))).toMatchObject({
      kind: "react", block: { role: "SELFISH" },
    });
  });

  it("ações do turno: ninguém barra Pegar moeda nem Trocar; Matar aceita Guardiã", () => {
    expect(selectDecision(withEvent(base, "EARN_COINS", "ana", { cause: "TAKE_COINS", earnedCoins: 5 }))).toEqual({ kind: "none" });
    expect(selectDecision(withEvent(base, "CHANGED_CARD", "ana", { cause: "SWAP_CARDS" }))).toEqual({ kind: "none" });
    expect(selectDecision(withEvent(base, "CARD_KILLED", "ana", { cause: "KILL_CARD", targetPlayer: "me" }))).toMatchObject({
      kind: "react", block: { role: "GUARDIAN", cost: 6 },
    });
  });

  it("fora da partida, fim e sem sala", () => {
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
    expect(selectStage(view("ContestaBlefe")).declared).toEqual({
      actorId: "ana", actorName: "Ana", role: "ASSASSIN", phase: "bluff",
      targetId: null, targetName: null, contesterId: "me", contesterName: "Matheus", chooserName: "Matheus",
    });
    expect(selectStage(view("ContestaVerdade")).declared?.phase).toBe("truth");
    expect(selectStage(view("ContestaVerdade")).declared?.chooserName).toBe("Ana");
    expect(selectStage(view("Main")).declared).toEqual({
      actorId: "ana", actorName: "Ana", role: "ASSASSIN", phase: "declared",
      targetId: "me", targetName: "Matheus", contesterId: null, contesterName: null, chooserName: null,
    });
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
      "(Bruno se disse Assassino, alguém duvidou e perdeu a carta.)",
      "Bruno: Isso foi pessoal.",
      "Você: Foi teatro, Bruno.",
      "Eva: Alguém conta as moedas da Carla, por favor.",
    ]);
    expect(selectChat(view("ConversaCheia"))).toHaveLength(30);
  });

  it("fim: vencedor e ordem de saída do último ao primeiro", () => {
    expect(selectFinale(view("Fim"))).toEqual({
      winnerName: "Carla",
      winnerIsMe: false,
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
    expect(s).toMatchObject({ code: "K7Q2", name: "Quinta dos mentirosos", max: 6, count: 3, castLine: "3 de 6 lugares ocupados", isHost: true });
    expect(s.seats.map((x) => [x.name, x.tag])).toEqual([
      ["Matheus", "anfitrião, você"], ["Ana", "pronto"], ["Bruno", "pronto"], ["Lugar vago", "aguardando"], ["Lugar vago", "aguardando"], ["Lugar vago", "aguardando"],
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

describe("narração: quem fez e quem sofreu", () => {
  const base = view("Mesa6");

  it("contestação sempre com os dois nomes", () => {
    expect(describeEvent(withEvent(base, "CONTEST", "bruno", { contestedPlayer: "ana", contestedCard: "ASSASSIN" }))).toBe(
      "Bruno duvidou de Ana.",
    );
    expect(
      describeEvent(withEvent(base, "CONTEST_PENALTY", "me", { contestedPlayer: "ana", contestingPlayer: "me", contestedCard: "ASSASSIN", hasCard: false })),
    ).toBe("Ana não tinha Assassino. Você escolhe qual carta de Ana cai.");
    expect(
      describeEvent(withEvent(base, "CONTEST_PENALTY", "ana", { contestedPlayer: "ana", contestingPlayer: "me", contestedCard: "ASSASSIN", hasCard: true })),
    ).toBe("Ana tinha Assassino. Ana escolhe qual carta sua cai.");
  });

  it("morte e ataque dizem quem matou e quem perdeu a carta", () => {
    expect(describeEvent(withEvent(base, "CARD_KILLED", "ana", { cause: "ASSASSIN", targetPlayer: "bruno", targetCardIndex: 0 }))).toBe(
      "Ana matou uma carta de Bruno.",
    );
    expect(describeEvent(withEvent(base, "CARD_KILLED", "bruno", { cause: "KAMIKAZE", targetPlayer: "eva", targetCardIndex: 1 }))).toBe(
      "Bruno levou uma carta de Eva junto com o Kamikaze.",
    );
  });

  it("Clarividente para todos: a carta revelada aparece no palco", () => {
    const v = withEvent(base, "REVEALED_CARD", "ana", { targetPlayer: "bruno", targetCardIndex: 0, showToAllPlayers: true, revealedCard: "KAMIKAZE" });
    expect(selectStage(v).revealed).toEqual({
      actorId: "ana", actorName: "Ana", targetId: "bruno", targetName: "Bruno", card: "KAMIKAZE", mine: false, toAll: true,
    });
    expect(describeEvent(v)).toBe("Kamikaze, carta de Bruno.");
  });

  it("Clarividente só para quem jogou: a carta aparece para você e fica escondida para os outros", () => {
    const payload = { targetPlayer: "bruno", targetCardIndex: 0, showToAllPlayers: false, revealedCard: "KAMIKAZE" };
    const mine = withEvent(base, "REVEALED_CARD", "me", payload);
    expect(selectStage(mine).revealed?.card).toBe("KAMIKAZE");
    expect(describeEvent(mine)).toBe("Só você viu: Kamikaze, carta de Bruno.");

    const other = withEvent(base, "REVEALED_CARD", "ana", { targetPlayer: "bruno", targetCardIndex: 0, showToAllPlayers: false });
    expect(selectStage(other).revealed?.card).toBeNull();
    expect(describeEvent(other)).toBe("Ana revelou uma carta de Bruno.");
  });
});

describe("gavetas", () => {
  const reveal = (targetId: string, cardIndex: number, turn: number, showToAll = true, card: RoleId = "KAMIKAZE"): LogEntry => ({
    key: `REVEALED_CARD|ana|${targetId}|${cardIndex}|${turn}`,
    turn,
    at: NOW,
    actorId: "ana",
    actorName: "Ana",
    pred: `revelou uma carta de ${targetId}: Kamikaze como Clarividente.`,
    reveal: { targetId, cardIndex, card, showToAll },
  });

  it("a revelação pública entra no alvo com selo e linha no o que a mesa já viu", () => {
    const base = view("Mesa6");
    const seat = selectSeatSheet({ ...base, log: [...base.log, reveal("bruno", 1, 5)] }, "bruno")!;
    expect(seat.revealed).toEqual([{ key: "REVEALED_CARD|ana|bruno|1|5", turn: 5, cardIndex: 1, card: "KAMIKAZE" }]);
    expect(seat.log.map((l) => [l.turn, l.text])).toEqual([
      [5, "Teve a carta Kamikaze revelada para todos."],
      [3, "Se disse Assassino, alguém duvidou e perdeu a carta."],
      [2, "Roubou 3 de Eva como Trapaceiro."],
      [1, "Recebeu 4 como Trilionário."],
    ]);
    // a carta em si continua no estado que a mesa vê
    expect(seat.cards[1]).toEqual({ state: "back", index: 1 });
  });

  it("a revelação privada não aparece na gaveta do alvo", () => {
    const base = view("Mesa6");
    const seat = selectSeatSheet({ ...base, log: [...base.log, reveal("bruno", 1, 5, false)] }, "bruno")!;
    expect(seat.revealed).toEqual([]);
    expect(seat.log.some((l) => l.text.includes("revelada"))).toBe(false);
  });

  it("duas revelações da mesma carta ficam com o turno mais recente", () => {
    const base = view("Mesa6");
    const log = [...base.log, reveal("bruno", 1, 5), reveal("bruno", 0, 4), reveal("bruno", 1, 7)];
    const seat = selectSeatSheet({ ...base, log }, "bruno")!;
    expect(seat.revealed.map((r) => [r.cardIndex, r.turn])).toEqual([
      [1, 7],
      [0, 4],
    ]);
    expect(seat.log.filter((l) => l.key.startsWith("REVEAL|"))).toHaveLength(2);
  });

  it("camarote: cartas, diário do ator e investimentos", () => {
    const seat = selectSeatSheet(view("Camarote"), "bruno")!;
    expect(seat).toMatchObject({ id: "bruno", name: "Bruno", coins: 2, alive: true, investments: [] });
    expect(seat.cards).toEqual([
      { state: "dead", role: "ASSASSIN", index: 0 },
      { state: "back", index: 1 },
    ]);
    expect(seat.log.map((l) => [l.turn, l.text])).toEqual([
      [3, "Se disse Assassino, alguém duvidou e perdeu a carta."],
      [2, "Roubou 3 de Eva como Trapaceiro."],
      [1, "Recebeu 4 como Trilionário."],
    ]);
  });

  it("a carta revelada pelo Clarividente fica marcada no camarote e no diário", () => {
    const base = view("Mesa6");
    const reveal: LogEntry = {
      key: "REVEALED_CARD|ana|1",
      turn: 4,
      at: NOW,
      actorId: "ana",
      actorName: "Ana",
      pred: "revelou uma carta de Bruno: Kamikaze como Clarividente.",
      reveal: { targetId: "bruno", cardIndex: 1, card: "KAMIKAZE" },
    };
    const v = { ...base, log: [...base.log, reveal] };
    const seat = selectSeatSheet(v, "bruno")!;
    expect(seat.cards[1]).toEqual({ state: "revealed", role: "KAMIKAZE", index: 1 });
    expect(selectSeats(v).find((s) => s.id === "bruno")!.cards[1]).toEqual({ state: "revealed", role: "KAMIKAZE", index: 1 });
    expect(seat.log.map((l) => l.text)).toContain("Ana revelou uma carta de Bruno: Kamikaze como Clarividente.");
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
