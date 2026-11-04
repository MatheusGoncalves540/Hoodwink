/**
 * Contrato com o servidor de jogo (Hoodwink-server/gameServer).
 *
 * O servidor manda, a cada mudança, o JSON de `PublicRoomForUpdates` (game/structs/rooms/room.go),
 * sem envelope. `parseRoom(raw)` valida com Zod e devolve um `PublicRoom` normalizado em camelCase,
 * ou `null` (com console.warn) se a mensagem não servir. Nunca lança.
 *
 * Normalizações:
 * - `gameEvent` não tem tags JSON no Go: chega como `PlayerID`, `Type`, `ExpiresAt`, `Payload`.
 *   Aceitamos as duas grafias e devolvemos `{ playerId, type, expiresAt (ms), payload }`.
 * - `deck` é número no estado público e lista no modo de depuração do servidor: vira o tamanho.
 * - listas `null` do Go (`investments`, `chat`) viram `[]`.
 * - a ordem de jogo é `moveTime` crescente (rooms/roomPlayers.go → GetNextRoundPlayer).
 */
import { z } from "zod";

export const ROLE_IDS = [
  "TRILLIONAIRE",
  "TRICKSTER",
  "INVESTOR",
  "ASSASSIN",
  "CLAIRVOYANT",
  "KAMIKAZE",
  "GUARDIAN",
  "GRAVEDIGGER",
  "CROUPIER",
  "POLITICAL",
  "REBEL",
  "SELFISH",
] as const;
export type RoleId = (typeof ROLE_IDS)[number];

const nullList = <T extends z.ZodType>(item: T) =>
  z
    .array(item)
    .nullish()
    .transform((v) => v ?? []);

export const publicCardSchema = z.object({
  name: z.enum(ROLE_IDS).optional(),
  index: z.number(),
  protected: z.boolean(),
  dead: z.boolean(),
});

export const publicPlayerSchema = z.object({
  id: z.string(),
  name: z.string(),
  moveTime: z.number().default(0),
  cards: z.array(publicCardSchema),
  coins: z.number(),
  investments: nullList(z.number()),
  pendingInvestmentCoin: z.boolean().default(false),
  alive: z.boolean(),
});

export const doubledSchema = z.object({
  timesValueDoubled: z.number(),
  roundsUntilDecrease: z.number(),
  usedThisTurn: z.boolean().default(false),
});

const pick = (o: Record<string, unknown>, a: string, b: string) => (a in o ? o[a] : o[b]);

export const gameEventSchema = z.preprocess(
  (raw) => {
    if (!raw || typeof raw !== "object") return raw;
    const o = raw as Record<string, unknown>;
    return {
      playerId: pick(o, "playerId", "PlayerID"),
      type: pick(o, "type", "Type"),
      expiresAt: pick(o, "expiresAt", "ExpiresAt"),
      payload: pick(o, "payload", "Payload"),
    };
  },
  z.object({
    playerId: z.string(),
    type: z.string(),
    expiresAt: z.string().transform((s, ctx) => {
      const ms = Date.parse(s);
      if (Number.isNaN(ms)) ctx.addIssue({ code: "custom", message: "ExpiresAt inválido" });
      return ms;
    }),
    payload: z.unknown(),
  }),
);

export const chatMessageSchema = z.object({
  playerName: z.string(),
  timestamp: z.string(),
  message: z.string(),
});

export const publicRoomSchema = z.object({
  id: z.string(),
  name: z.string().default(""),
  /** ainda não vem em PublicRoomForUpdates (pedido ao servidor); quem cria a sala sabe o número */
  maxPlayers: z.number().optional(),
  turn: z.number(),
  tax: z.number(),
  doubledCardValues: z
    .record(z.string(), doubledSchema)
    .nullish()
    .transform((v) => v ?? {}),
  players: z.record(z.string(), publicPlayerSchema),
  deck: z.union([z.number(), z.array(z.unknown()).transform((a) => a.length)]),
  currentPlayer: z.string().default(""),
  gameEvent: gameEventSchema.nullish().transform((v) => v ?? null),
  gameOver: z.boolean().default(false),
  chat: nullList(chatMessageSchema),
});

export type PublicCard = z.infer<typeof publicCardSchema>;
export type PublicPlayer = z.infer<typeof publicPlayerSchema>;
export type Doubled = z.infer<typeof doubledSchema>;
export type GameEvent = z.infer<typeof gameEventSchema>;
export type ChatMessage = z.infer<typeof chatMessageSchema>;
export type PublicRoom = z.infer<typeof publicRoomSchema>;

/** Valida uma mensagem do servidor. Inválida: console.warn e null. */
export function parseRoom(raw: unknown): PublicRoom | null {
  const r = publicRoomSchema.safeParse(raw);
  if (r.success) return r.data;
  console.warn("Mensagem do servidor ignorada:", r.error.issues);
  return null;
}

/** Jogadores na ordem de jogo (moveTime crescente; empate pelo id, para ser estável). */
export function playOrder(room: PublicRoom): PublicPlayer[] {
  return Object.values(room.players).sort((a, b) => a.moveTime - b.moveTime || a.id.localeCompare(b.id));
}
