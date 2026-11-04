/**
 * Mensagens do cliente para o gameServer: { type, payload }, como no ButtonPanel.jsx legado.
 * O servidor valida com DecodeStrictJSON (structs/validatePlayerPlayPayload.go): campo a mais é recusado.
 */
import type { RoleId } from "../game/protocol";

export type ClientMessage = { type: string; payload: unknown };
export type Target = { playerId?: string; cardIndex?: number; allyCardIndex?: number };

export const startGame = (): ClientMessage => ({ type: "COMMAND_START_GAME", payload: {} });
export const contest = (): ClientMessage => ({ type: "CONTEST", payload: {} });
export const contestPenalty = (cardIndex: number): ClientMessage => ({ type: "CONTEST_PENALTY", payload: { TargetCardIndex: cardIndex } });
export const chat = (msg: string): ClientMessage => ({ type: "CHAT_MESSAGE", payload: { msg } });

/**
 * Jogada de papel (também as reações Guardiã e Egoísta).
 * Papéis de carta própria (Guardiã, Coveiro, Crupiê) sempre miram você.
 */
export function play(role: RoleId, me: string, t: Target = {}): ClientMessage {
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
