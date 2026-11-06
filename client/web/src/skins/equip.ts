/** Funções puras de skin (sem DOM). `resolveEquip` é a fronteira de confiança do que vem salvo. */
import { DEFAULT_EQUIP, SLOTS, itemOf, type Equip, type Item, type Pack } from "./catalog";

export function resolveEquip(saved: unknown): Equip {
  let data: unknown = saved;
  if (typeof saved === "string") {
    try {
      data = JSON.parse(saved);
    } catch {
      data = null;
    }
  }
  const obj = data && typeof data === "object" ? (data as Record<string, unknown>) : {};
  const out = { ...DEFAULT_EQUIP };
  for (const slot of SLOTS) {
    const id = obj[slot];
    if (typeof id === "string" && itemOf(slot, id)) out[slot] = id;
  }
  return out;
}

export const equipPack = (pack: Pack): Equip => ({ ...pack.equip });
export const equipItem = (equip: Equip, item: Item): Equip => ({ ...equip, [item.slot]: item.id });
