/** Único lugar que mexe no DOM por causa de skin: carrega o CSS das peças e escreve data-sk-* no elemento. */
import { SLOTS, itemOf, type Equip, type Slot } from "./catalog";

const attr = (s: Slot) => `sk${s.charAt(0).toUpperCase()}${s.slice(1)}`;

export async function applySkin(el: HTMLElement, equip: Equip) {
  // carrega antes de trocar: sem piscar. Chunk que falhar segue com o padrão daquela categoria.
  await Promise.all(SLOTS.map((s) => itemOf(s, equip[s])?.load?.().catch(() => undefined)));
  for (const s of SLOTS) el.dataset[attr(s)] = equip[s];
}

/** Atributos para um contêiner de prévia (Camarim): mesmos nomes, sem tocar no documento. */
export const skinAttrs = (equip: Equip) => Object.fromEntries(SLOTS.map((s) => [`data-sk-${s}`, equip[s]]));
