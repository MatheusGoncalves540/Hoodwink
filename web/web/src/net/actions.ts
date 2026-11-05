/**
 * Mensagens do cliente para o gameServer: { type, payload }, como no painel de botões antigo.
 * O servidor valida com DecodeStrictJSON (structs/validatePlayerPlayPayload.go): campo a mais é recusado.
 */
import type { RoleId } from "../game/protocol";

export type ClientMessage = { type: string; payload: unknown };
export type Target = { playerId?: string; cardIndex?: number; allyCardIndex?: number };

export const startGame = (): ClientMessage => ({ type: "COMMAND_START_GAME", payload: {} });
export const contest = (): ClientMessage => ({ type: "CONTEST", payload: {} });
export const contestPenalty = (cardIndex: number): ClientMessage => ({ type: "CONTEST_PENALTY", payload: { TargetCardIndex: cardIndex } });
export const chat = (msg: string): ClientMessage => ({ type: "CHAT_MESSAGE", payload: { msg } });

/** Papéis que exigem alvo (jogador e/ou carta) antes de enviar; o servidor fecha o socket se faltar. */
const NEEDS_PLAYER: RoleId[] = ["ASSASSIN", "KAMIKAZE", "CLAIRVOYANT", "TRICKSTER"];
const NEEDS_CARD: RoleId[] = ["ASSASSIN", "KAMIKAZE", "CLAIRVOYANT", "GUARDIAN", "GRAVEDIGGER", "CROUPIER"];

/**
 * Jogada de papel (também as reações Guardiã e Egoísta).
 * Papéis de carta própria (Guardiã, Coveiro, Crupiê) sempre miram você.
 * Falta de alvo é erro de programação do fluxo de alvo: lança antes de enviar.
 */
export function play(role: RoleId, me: string, t: Target = {}): ClientMessage {
  if (NEEDS_PLAYER.includes(role) && !t.playerId) throw new Error(`${role} exige targetPlayer`);
  if (NEEDS_CARD.includes(role) && t.cardIndex == null) throw new Error(`${role} exige targetCardIndex`);
  const payload = (() => {
    switch (role) {
      case "ASSASSIN":
        return { targetPlayer: t.playerId, targetCardIndex: t.cardIndex };
      case "KAMIKAZE":
        return { targetPlayer: t.playerId, targetCardIndex: t.cardIndex, targetAllyCardIndex: t.allyCardIndex };
      case "CLAIRVOYANT":
        return { targetPlayer: t.playerId, targetCardIndex: t.cardIndex, showToAllPlayers: false };
      case "TRICKSTER":
        return { targetPlayer: t.playerId };
      case "GUARDIAN":
      case "GRAVEDIGGER":
        return { targetPlayer: me, targetCardIndex: t.cardIndex };
      case "CROUPIER":
        return { targetPlayer: me, targetCardIndex: t.cardIndex, useOnTwoCards: false };
      default:
        return null;
    }
  })();
  return { type: role, payload };
}
