/**
 * Camarim (plans/03-skins.md): escolher pacote ou peça por categoria. Tocar veste na hora e salva.
 * Sem preço, sem cadeado. A prévia e cada amostra recebem os próprios data-sk-* (vestem só aquele pedaço).
 * As amostras ficam paradas (data-idle="off" no contêiner) para respeitar o orçamento de contornos.
 */
import { useEffect, useState } from "react";
import type { Layout } from "../../app/useLayout";
import { useLayout } from "../../app/useLayout";
import { setPref, SKIN_KEY, usePref } from "../../app/prefs";
import { useTheme, type Theme } from "../../app/useTheme";
import { copy } from "../../game/copy";
import { applySkin, skinAttrs } from "../../skins/applySkin";
import { PACKS, SLOTS, itemOf, itemsOf, type Equip, type Slot } from "../../skins/catalog";
import { equipItem, equipPack, resolveEquip } from "../../skins/equip";
import { CardBack } from "../../ui/Card/Card";
import { Moldura } from "../../ui/Moldura/Moldura";
import { Ticket } from "../../ui/Ticket/Ticket";
import { cx } from "../../ui/cx";
import s from "./Camarim.module.css";

export type Tab = "pacotes" | Slot;
const TABS: Tab[] = ["pacotes", ...SLOTS];
const c = copy.camarim;
const same = (a: Equip, b: Equip) => SLOTS.every((k) => a[k] === b[k]);

/** Amostra: o pedaço do jogo que a categoria muda. */
function Sample({ tab, name }: { tab: Tab; name: string }) {
  if (tab === "pacotes")
    return (
      <span className={s.pack} aria-hidden="true">
        <b>{name}</b>
        <i />
        <i />
      </span>
    );
  if (tab === "cartas") return <CardBack size="s" />;
  if (tab === "botoes") return <span className={cx("ticket", s.miniTicket)}>{c.previewButton}</span>;
  if (tab === "molduras") return <Moldura lugar="elenco" className={s.miniMoldura}>{c.previewCast}</Moldura>;
  if (tab === "letras") return <span className={s.letras}>Aa</span>;
  return (
    <span className={s.cores} aria-hidden="true">
      <i />
      <i />
      <i />
      <i />
    </span>
  );
}

type BodyProps = { layout: Layout; theme: Theme; equip: Equip; tab: Tab; onTab(t: Tab): void; onEquip(e: Equip): void };

/** Corpo puro (sem estado): testável em Node. */
export function CamarimBody({ layout, theme, equip, tab, onTab, onEquip }: BodyProps) {
  const tiles =
    tab === "pacotes"
      ? PACKS.map((p) => ({ key: p.id, name: p.name, next: equipPack(p), worn: same(equip, p.equip) }))
      : itemsOf(tab).map((i) => ({ key: i.id, name: i.name, next: equipItem(equip, i), worn: equip[tab] === i.id }));
  return (
    <main className={cx(s.screen, layout === "desktop" && s.desk)}>
      <div className={s.side}>
        <header>
          <a className={s.back} href="#">
            {c.back}
          </a>
          <h1 className={s.title}>{c.title}</h1>
        </header>

        <section aria-label={c.preview} className={s.preview} data-theme={theme} {...skinAttrs(equip)}>
          <Moldura lugar="elenco" className={s.previewCast}>
            <span className={s.previewRoom}>{c.previewRoom}</span>
            <b>{c.previewCast}</b>
            <span className={s.previewCount}>{c.previewCount}</span>
          </Moldura>
          <div className={s.previewCards}>
            <CardBack size="s" />
            <CardBack size="s" />
          </div>
          <Ticket className={s.previewTicket}>{c.previewButton}</Ticket>
        </section>
      </div>

      <div className={s.main}>
        <p className={s.lead}>{c.lead}</p>
        <div role="tablist" aria-label={c.title} className={s.tabs}>
          {TABS.map((t) => (
            <button key={t} type="button" role="tab" aria-selected={t === tab} className={s.tab} onClick={() => onTab(t)}>
              {c.tabs[t]}
            </button>
          ))}
        </div>
        <p className={s.tabLead}>{c.tabLead[tab]}</p>

        <div role="tabpanel" aria-label={c.tabs[tab]} className={s.grid}>
          {tiles.map((t) => (
            <button key={t.key} type="button" aria-pressed={t.worn} className={s.tile} onClick={() => onEquip(t.next)}>
              <span className={s.sample} data-theme={theme} data-idle="off" {...skinAttrs(t.next)}>
                <Sample tab={tab} name={t.name} />
              </span>
              <span className={s.tileName}>{t.name}</span>
              {t.worn && <span className={s.worn}>{c.worn}</span>}
            </button>
          ))}
        </div>
      </div>
    </main>
  );
}

/** Vestir: salva e aplica no documento inteiro. `apply` entra como parâmetro para o teste. */
export function wear(equip: Equip, el: HTMLElement, apply = applySkin) {
  setPref(SKIN_KEY, JSON.stringify(equip));
  return apply(el, equip);
}

/** Peças a pré-carregar ao abrir uma aba, para a prévia trocar sem espera. */
export const preloadOf = (tab: Tab) =>
  tab === "pacotes" ? PACKS.flatMap((p) => SLOTS.map((k) => itemOf(k, p.equip[k])!)) : itemsOf(tab);

export function Camarim() {
  const layout = useLayout();
  const { theme } = useTheme();
  const equip = resolveEquip(usePref(SKIN_KEY));
  const [tab, setTab] = useState<Tab>("pacotes");
  useEffect(() => {
    preloadOf(tab).forEach((i) => void i.load?.().catch(() => undefined));
  }, [tab]);
  return (
    <CamarimBody layout={layout} theme={theme} equip={equip} tab={tab} onTab={setTab} onEquip={(e) => void wear(e, document.documentElement)} />
  );
}
