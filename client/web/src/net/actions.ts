/**
 * Mensagens do cliente para o gameServer: { type, payload }, como no painel de botões antigo.
 * O servidor valida com DecodeStrictJSON (structs/validatePlayerPlayPayload.go): campo a mais é recusado.
 */
import type { ActionId } from "../game/roles";

export type ClientMessage = { type: string; payload: unknown };
export type Target = { playerId?: string; cardIndex?: number; allyCardIndex?: number; showToAll?: boolean; both?: boolean };

export const startGame = (): ClientMessage => ({ type: "COMMAND_START_GAME", payload: {} });
export const pass = (): ClientMessage => ({ type: "PASS", payload: {} });
export const contest = (): ClientMessage => ({ type: "CONTEST", payload: {} });
export const contestPenalty = (cardIndex: number): ClientMessage => ({ type: "CONTEST_PENALTY", payload: { TargetCardIndex: cardIndex } });
export const chat = (msg: string): ClientMessage => ({ type: "CHAT_MESSAGE", payload: { msg } });

/** Papéis que exigem alvo (jogador e/ou carta) antes de enviar; o servidor fecha o socket se faltar. */
const NEEDS_PLAYER: ActionId[] = ["ASSASSIN", "KAMIKAZE", "CLAIRVOYANT", "TRICKSTER", "KILL_CARD"];
const NEEDS_CARD: ActionId[] = ["ASSASSIN", "KAMIKAZE", "CLAIRVOYANT", "GUARDIAN", "GRAVEDIGGER", "CROUPIER", "KILL_CARD", "SWAP_CARDS"];

/**
 * Jogada de personagem (também as reações Guardiã e Egoísta) e ações do turno (02 §6: TAKE_COINS, KILL_CARD, SWAP_CARDS).
 * Papéis de carta própria (Guardiã, Coveiro, Crupiê) sempre miram você.
 * Falta de alvo é erro de programação do fluxo de alvo: lança antes de enviar.
 */
export function play(role: ActionId, me: string, t: Target = {}): ClientMessage {
  if (NEEDS_PLAYER.includes(role) && !t.playerId) throw new Error(`${role} exige targetPlayer`);
  if (NEEDS_CARD.includes(role) && t.cardIndex == null) throw new Error(`${role} exige targetCardIndex`);
  const payload = (() => {
    switch (role) {
      case "ASSASSIN":
      case "KILL_CARD":
        return { targetPlayer: t.playerId, targetCardIndex: t.cardIndex };
      case "TAKE_COINS":
        return {};
      case "SWAP_CARDS":
        return t.both ? { targetCardIndex: t.cardIndex, useOnTwoCards: true } : { targetCardIndex: t.cardIndex };
      case "KAMIKAZE":
        return { targetPlayer: t.playerId, targetCardIndex: t.cardIndex, targetAllyCardIndex: t.allyCardIndex };
      case "CLAIRVOYANT":
        return { targetPlayer: t.playerId, targetCardIndex: t.cardIndex, showToAllPlayers: !!t.showToAll };
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
