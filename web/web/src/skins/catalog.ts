/**
 * Catálogo de skins (plans/03-skins.md). Dados puros: categorias, peças, pacotes e peça padrão.
 * A peça padrão de cada categoria é a skin Holofote, que já está em tokens.css/components.css (sem `load`).
 * Sem preço nem cadeado: tudo liberado. Quando houver posse, só `isUnlocked` muda.
 */
export const SLOTS = ["cores", "letras", "molduras", "botoes", "cartas"] as const;
export type Slot = (typeof SLOTS)[number];
export type Equip = Record<Slot, string>;

export type Item = {
  id: string;
  slot: Slot;
  name: string;
  /** Carrega o CSS da peça (chunk do Vite). Ausente na peça padrão. */
  load?: () => Promise<unknown>;
};
export type Pack = { id: string; name: string; equip: Equip };

export const DEFAULT_EQUIP: Equip = { cores: "holofote", letras: "cartaz", molduras: "holofote", botoes: "ingresso", cartas: "veludo" };

export const ITEMS: Item[] = [
  { id: "holofote", slot: "cores", name: "Holofote" },
  { id: "bastidor", slot: "cores", name: "Bastidor", load: () => import("./cores/bastidor.css") },
  { id: "noir", slot: "cores", name: "Noir", load: () => import("./cores/noir.css") },
  { id: "ensaio", slot: "cores", name: "Ensaio", load: () => import("./cores/ensaio.css") },

  { id: "cartaz", slot: "letras", name: "Cartaz" },
  { id: "pincel", slot: "letras", name: "Pincel", load: () => import("./letras/pincel.css") },
  { id: "gala", slot: "letras", name: "Gala", load: () => import("./letras/gala.css") },

  { id: "holofote", slot: "molduras", name: "Holofote" },
  { id: "canhoto", slot: "molduras", name: "Canhoto", load: () => import("./molduras/canhoto.css") },
  { id: "bambinela", slot: "molduras", name: "Bambinela", load: () => import("./molduras/bambinela.css") },
  { id: "cantoneiras", slot: "molduras", name: "Cantoneiras", load: () => import("./molduras/cantoneiras.css") },
  { id: "papelao", slot: "molduras", name: "Papelão", load: () => import("./molduras/papelao.css") },

  { id: "ingresso", slot: "botoes", name: "Ingresso" },
  { id: "placa", slot: "botoes", name: "Placa", load: () => import("./botoes/placa.css") },
  { id: "carimbo", slot: "botoes", name: "Carimbo", load: () => import("./botoes/carimbo.css") },
  { id: "veludo", slot: "botoes", name: "Veludo", load: () => import("./botoes/veludo.css") },

  { id: "veludo", slot: "cartas", name: "Veludo" },
  { id: "papelao", slot: "cartas", name: "Papelão", load: () => import("./cartas/papelao.css") },
  { id: "xadrez", slot: "cartas", name: "Xadrez", load: () => import("./cartas/xadrez.css") },
  { id: "azulejo", slot: "cartas", name: "Azulejo", load: () => import("./cartas/azulejo.css") },
];

export const PACKS: Pack[] = [
  { id: "holofote", name: "Holofote", equip: DEFAULT_EQUIP },
  { id: "bastidor", name: "Bastidor", equip: { cores: "bastidor", letras: "pincel", molduras: "papelao", botoes: "placa", cartas: "papelao" } },
  { id: "noir", name: "Noir", equip: { cores: "noir", letras: "gala", molduras: "cantoneiras", botoes: "carimbo", cartas: "xadrez" } },
  { id: "ensaio", name: "Ensaio", equip: { cores: "ensaio", letras: "cartaz", molduras: "canhoto", botoes: "veludo", cartas: "azulejo" } },
];

/** Tokens de cor que toda peça de `cores` redefine, no claro e no escuro. As cores da carta são da categoria `cartas`. */
export const COLOR_TOKENS = [
  "plateia", "luz", "tinta", "tinta-suave", "titulo", "veludo", "sobre-veludo", "alerta", "ouro", "ouro-texto",
  "lampada", "filete", "poca", "cone", "scrim",
].map((t) => `--hw-${t}`);

export const itemOf = (slot: Slot, id: string) => ITEMS.find((i) => i.slot === slot && i.id === id);
export const itemsOf = (slot: Slot) => ITEMS.filter((i) => i.slot === slot);

/** Hoje tudo é liberado. Quando houver posse, só esta função muda. */
export const isUnlocked = (_item: Item) => true;
