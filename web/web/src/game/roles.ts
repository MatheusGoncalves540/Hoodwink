/**
 * Os doze papéis (02-design-system.md → "Os doze papéis"; textos e olhares de Acoes.dc.html).
 * - `target`: 'other' escolhe um ator e (às vezes) uma carta dele; 'own' escolhe uma carta sua; null, nada.
 * - `look`: o par de olhos provisório da faixa de arte das ações, até chegar a arte final.
 * As descrições com número (preço) são montadas em copy.ts a partir de pricing.ts.
 */
import type { RoleId } from "./protocol";

export type Tone = "vinho" | "ouro" | "cedula";
export type Look = { h: number; tilt: number; side?: number; wink?: boolean; monocle?: boolean; raise?: boolean; dollar?: boolean; star?: boolean; small?: boolean };
export type Role = { id: RoleId; name: string; target: "other" | "own" | null; look: Look };
export type RoleGroup = { title: string; tone: Tone; offTurn?: boolean; roles: RoleId[] };

export const ROLES: Record<RoleId, Role> = {
  TRILLIONAIRE: { id: "TRILLIONAIRE", name: "Trilionário", target: null, look: { h: 9, tilt: 2, monocle: true, raise: true } },
  TRICKSTER: { id: "TRICKSTER", name: "Trapaceiro", target: "other", look: { h: 11, wink: true, tilt: 3 } },
  INVESTOR: { id: "INVESTOR", name: "Investidor", target: null, look: { h: 13, tilt: -2, dollar: true } },
  ASSASSIN: { id: "ASSASSIN", name: "Assassino", target: "other", look: { h: 5, tilt: 9 } },
  CLAIRVOYANT: { id: "CLAIRVOYANT", name: "Clarividente", target: "other", look: { h: 15, star: true, tilt: -2 } },
  KAMIKAZE: { id: "KAMIKAZE", name: "Kamikaze", target: "other", look: { h: 20, tilt: -8, small: true } },
  GUARDIAN: { id: "GUARDIAN", name: "Guardiã", target: "own", look: { h: 8, tilt: 0 } },
  GRAVEDIGGER: { id: "GRAVEDIGGER", name: "Coveiro", target: "own", look: { h: 1, tilt: 2 } },
  CROUPIER: { id: "CROUPIER", name: "Crupiê", target: "own", look: { h: 11, side: 8, tilt: 0 } },
  POLITICAL: { id: "POLITICAL", name: "Político", target: null, look: { h: 9, tilt: -7 } },
  REBEL: { id: "REBEL", name: "Rebelde", target: null, look: { h: 11, tilt: 7, side: -6 } },
  SELFISH: { id: "SELFISH", name: "Egoísta", target: null, look: { h: 6, tilt: 4, side: -9 } },
};

export const ROLE_GROUPS: RoleGroup[] = [
  { title: "Para ganhar moedas", tone: "ouro", roles: ["TRILLIONAIRE", "TRICKSTER", "INVESTOR"] },
  { title: "Para atacar", tone: "vinho", roles: ["ASSASSIN", "CLAIRVOYANT", "KAMIKAZE"] },
  { title: "Para mexer nas cartas", tone: "vinho", roles: ["GUARDIAN", "GRAVEDIGGER", "CROUPIER"] },
  { title: "Para mudar a taxa", tone: "cedula", roles: ["POLITICAL", "REBEL"] },
  { title: "Fora do seu turno", tone: "vinho", offTurn: true, roles: ["SELFISH"] },
];

export const roleName = (id: RoleId | undefined) => (id ? ROLES[id].name : "?");

/** Papel declarado por um evento `CARD_PLAYED_<PAPEL>`, ou null. */
export function declaredRole(eventType: string | undefined): RoleId | null {
  const m = eventType?.match(/^CARD_PLAYED_([A-Z]+)$/);
  return m && m[1] in ROLES ? (m[1] as RoleId) : null;
}
