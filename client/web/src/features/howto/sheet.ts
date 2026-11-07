/**
 * Ficha de cada personagem na tela Como se joga (canvas "Como se joga", Refino v2).
 * Os números vêm de pricing.ts com a taxa de teste; os textos fixos seguem o canvas.
 */
import { cardValue, guardianPrice, priceOf, taxed, type PriceRoom } from "../../game/pricing";
import type { RoleId } from "../../game/protocol";
import { ROLES } from "../../game/roles";
import { CLASSIC_RULES } from "../../game/rules.classic";

export type Sheet = {
  id: RoleId;
  group: number;
  name: string;
  /** o que vai na etiqueta e na moeda: "+3" para ganho, "5" para custo */
  price: string;
  efeito: string;
  quando: string;
  alvo: string;
  faixa: string;
  barra: string;
  contesta: string;
  nota: string;
};

/** Ordem do canvas: cinco grupos, na ordem em que aparecem na lista. */
export const GROUPS: RoleId[][] = [
  ["TRILLIONAIRE", "TRICKSTER", "INVESTOR"],
  ["ASSASSIN", "CLAIRVOYANT", "KAMIKAZE"],
  ["GUARDIAN", "GRAVEDIGGER", "CROUPIER"],
  ["POLITICAL", "REBEL"],
  ["SELFISH"],
];

const T = "No seu turno";
const SIM = "Sim, como qualquer personagem";
const DOBRA = "O preço dobra para todos. Após 3 turnos sem uso, cai pela metade.";

export function sheetOf(id: RoleId, tax: number): Sheet {
  const room: PriceRoom = { tax, doubledCardValues: {} };
  const v = cardValue(id, room);
  const p = priceOf(id, room);
  const sel = priceOf("SELFISH", room);
  const gate = sel.kind === "reaction" ? Math.max(1, sel.threshold) : 0;
  const group = GROUPS.findIndex((g) => g.includes(id));
  const base = { id, group, name: ROLES[id].name, price: String(v), quando: T, contesta: SIM, nota: "" };
  switch (id) {
    case "TRILLIONAIRE":
      return { ...base, price: `+${v}`, efeito: `Recebe ${v} moedas.`, alvo: "Ninguém", faixa: "3 + taxa, no mínimo 1", barra: `Egoísta, se o ganho for ${gate} ou mais` };
    case "TRICKSTER": {
      const keep = p.coins;
      return { ...base, price: `+${keep}`, efeito: `Tira ${v} moedas de outro jogador e fica com ${keep}.`, alvo: "Outro jogador", faixa: "Tira 2 + taxa, de 2 a 5", barra: "Egoísta, se o ganho passar de 1", nota: "Em número ímpar, o Trapaceiro fica com a metade maior." };
    }
    case "INVESTOR": {
      const cap = taxed(CLASSIC_RULES.investorCap, tax);
      return { ...base, price: "1", efeito: "Investe 1 moeda. Em todo turno seu, rende metade do que está investido, em moedas inteiras.", quando: `${T}, 1 moeda por turno`, alvo: "Ninguém", faixa: `Até ${cap} moedas investidas (4 + taxa, de 2 a 6)`, barra: "Egoísta barra o investimento, nunca o rendimento", contesta: "Só quando investe", nota: "Cada moeda volta para a sua mão após 6 turnos. Se a taxa cair, o que passar do teto é descartado." };
    }
    case "ASSASSIN":
      return { ...base, efeito: "Mata uma carta de outro jogador.", alvo: "Uma carta de outro jogador", faixa: "4 + taxa, no mínimo 3", barra: "Guardiã" };
    case "CLAIRVOYANT":
      return { ...base, efeito: "Revela uma carta de outro jogador, para todos ou só para você.", alvo: "Uma carta de outro jogador", faixa: "4 + taxa, no mínimo 2", barra: "Guardiã", nota: "Depois a carta volta a se esconder, no mesmo lugar." };
    case "KAMIKAZE":
      return { ...base, price: "0", efeito: "Quando morre, leva junto uma carta de outro jogador.", quando: "Quando morre, ou de propósito no seu turno", alvo: "Uma carta de outro jogador", faixa: "De graça. A taxa não muda o preço", barra: "Ninguém. Não tem defesa" };
    case "GUARDIAN":
      return { ...base, efeito: "Impede um ataque contra uma carta sua.", quando: "Como reação, ou no seu turno para o próximo ataque", alvo: "Uma carta sua", faixa: `Reação: ${v} · contra Clarividente: ${guardianPrice("clairvoyant", room)} · no seu turno: ${guardianPrice("ownTurn", room)}`, barra: "Ninguém", nota: "Uma carta protegida por vez. Não segura Kamikaze nem contestação." };
    case "GRAVEDIGGER":
      return { ...base, efeito: "Revive uma carta morta sua, a mesma que você tinha.", alvo: "Uma carta morta sua", faixa: "5 + taxa, no mínimo 3", barra: "Egoísta", nota: "Não funciona com a mão cheia." };
    case "CROUPIER":
      return { ...base, efeito: "Descarta uma carta sua, viva ou morta, e puxa uma nova do baralho.", alvo: "Uma carta sua", faixa: "6 + taxa, no mínimo 4", barra: "Egoísta" };
    case "POLITICAL":
      return { ...base, efeito: "Sobe a taxa em 1.", alvo: "A mesa toda", faixa: "2, dobra a cada uso: 4, depois 8", barra: "Ninguém", nota: DOBRA };
    case "REBEL":
      return { ...base, efeito: "Baixa a taxa em 1.", alvo: "A mesa toda", faixa: "2, dobra a cada uso: 4, depois 8", barra: "Ninguém", nota: DOBRA };
    case "SELFISH": {
      const card = sel.kind === "reaction" ? sel.cardChange : v;
      return { ...base, price: "1", efeito: `Por 1, impede alguém de receber ${gate} ou mais moedas. Por ${card}, impede de receber uma carta.`, quando: "Só fora do seu turno", alvo: "Quem está recebendo", faixa: "Moedas: 1 fixo · carta: 5 + taxa, de 3 a 8", barra: "—", nota: "Também barra o investimento, o Coveiro e o Crupiê." };
    }
  }
}
