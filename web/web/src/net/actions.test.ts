/**
 * Paridade com o painel de botões antigo (removido na T26): o servidor usa
 * DecodeStrictJSON, então nome e forma de cada campo importam — campo a mais é recusado.
 * A tabela abaixo é a cópia literal das mensagens do painel antigo, com os mesmos alvos.
 */
import { describe, expect, it } from "vitest";
import { chat, contest, contestPenalty, play, startGame } from "./actions";

const me = "eu";
const targetPlayer = "ana";
const targetCardIndex = 2;

// No painel antigo, todas as cartas usam o mesmo índice no alvo e no aliado.
const legacyPlay = (type: string, payload: unknown) => ({ type, payload });

describe("actions", () => {
  it("monta as mensagens de comando iguais às do painel antigo", () => {
    expect(startGame()).toEqual(legacyPlay("COMMAND_START_GAME", {}));
    expect(contest()).toEqual(legacyPlay("CONTEST", {}));
    expect(contestPenalty(targetCardIndex)).toEqual(legacyPlay("CONTEST_PENALTY", { TargetCardIndex: targetCardIndex }));
    expect(chat("ablubluble")).toEqual(legacyPlay("CHAT_MESSAGE", { msg: "ablubluble" }));
  });

  it("monta as mensagens de papel iguais às do painel antigo", () => {
    expect(play("TRILLIONAIRE", me)).toEqual(legacyPlay("TRILLIONAIRE", null));
    expect(play("POLITICAL", me)).toEqual(legacyPlay("POLITICAL", null));
    expect(play("REBEL", me)).toEqual(legacyPlay("REBEL", null));
    expect(play("INVESTOR", me)).toEqual(legacyPlay("INVESTOR", null));
    expect(play("SELFISH", me)).toEqual(legacyPlay("SELFISH", null));

    expect(play("ASSASSIN", me, { playerId: targetPlayer, cardIndex: targetCardIndex })).toEqual(
      legacyPlay("ASSASSIN", { targetPlayer, targetCardIndex }),
    );
    expect(play("KAMIKAZE", me, { playerId: targetPlayer, cardIndex: targetCardIndex, allyCardIndex: targetCardIndex })).toEqual(
      legacyPlay("KAMIKAZE", { targetPlayer, targetCardIndex, targetAllyCardIndex: targetCardIndex }),
    );
    expect(play("CLAIRVOYANT", me, { playerId: targetPlayer, cardIndex: targetCardIndex })).toEqual(
      legacyPlay("CLAIRVOYANT", { targetPlayer, targetCardIndex, showToAllPlayers: false }),
    );
    expect(play("GUARDIAN", me, { cardIndex: targetCardIndex })).toEqual(
      legacyPlay("GUARDIAN", { targetPlayer: me, targetCardIndex }),
    );
    expect(play("TRICKSTER", me, { playerId: targetPlayer })).toEqual(legacyPlay("TRICKSTER", { targetPlayer }));
    expect(play("GRAVEDIGGER", me, { cardIndex: targetCardIndex })).toEqual(
      legacyPlay("GRAVEDIGGER", { targetPlayer: me, targetCardIndex }),
    );
    expect(play("CROUPIER", me, { cardIndex: targetCardIndex })).toEqual(
      legacyPlay("CROUPIER", { targetPlayer: me, targetCardIndex, useOnTwoCards: false }),
    );
  });

  it("não manda campo indefinido no JSON (o servidor recusa campo desconhecido)", () => {
    expect(JSON.parse(JSON.stringify(play("TRICKSTER", me, { playerId: targetPlayer })))).toEqual({
      type: "TRICKSTER",
      payload: { targetPlayer },
    });
  });
});
