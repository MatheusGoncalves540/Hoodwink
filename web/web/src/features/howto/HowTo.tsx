/**
 * Como se joga: bastidor em papelão (canvas "Como se joga", Refino v2, desktop e celular).
 * Placas de papelão nos títulos, folhas presas com durex, rabiscos em Permanent Marker.
 * Movimento: só o papel das folhas (.m-q3), o picote do canhoto e as setas (.m-tira).
 * A chave "Movimento" grava a preferência Animações de espera (a mesma do menu), não uma própria.
 */
import { useEffect, useState } from "react";
import type { Layout } from "../../app/useLayout";
import { useLayout } from "../../app/useLayout";
import { useMotionPref } from "../../app/useMotionPref";
import { copy, signed } from "../../game/copy";
import { killCost, basicPrice } from "../../game/pricing";
import type { RoleId } from "../../game/protocol";
import { CLASSIC_RULES } from "../../game/rules.classic";
import { cx } from "../../ui/cx";
import { GROUPS, sheetOf } from "./sheet";
import s from "./HowTo.module.css";

const h = copy.howto;

/** Três quadros da mesma seta (traço fervendo). */
function Arrow() {
  const paths = [
    "M4 8 C20 4 40 10 48 30 M40 24 L48 31 L53 21",
    "M5 9 C21 3 41 11 47 29 M39 23 L47 30 L54 22",
    "M3 7 C19 5 39 9 49 31 M41 25 L49 32 L52 20",
  ];
  return (
    <span aria-hidden="true" className={cx("m-tira", s.arrow)}>
      <span>
        {paths.map((d) => (
          <svg key={d} viewBox="0 0 60 40">
            <path d={d} />
          </svg>
        ))}
      </span>
    </span>
  );
}

/** "?" do placeholder da arte, em três quadros. */
function Ask() {
  const q = [
    ["M20 50 C14 22 34 8 52 8 C72 8 82 24 80 42 C78 62 52 66 52 90 L52 104", 14, 52, 150, 34, 8],
    ["M22 48 C15 24 33 10 53 9 C71 7 83 25 79 43 C76 61 51 67 53 91 L51 103", 13, 51, 151, 33, 7.5],
    ["M19 51 C16 21 36 7 51 8 C73 9 81 23 81 41 C79 63 53 65 51 89 L53 105", 15, 53, 149, 35, 8.5],
  ] as const;
  return (
    <span aria-hidden="true" className={cx("m-tira", s.ask)}>
      <span>
        {q.map(([d, w, cx_, cy, rx, ry]) => (
          <svg key={d} viewBox="0 0 100 160">
            <path d={d} strokeWidth={w} />
            <ellipse cx={cx_} cy={cy} rx={rx} ry={ry} />
          </svg>
        ))}
      </span>
    </span>
  );
}

/** Picote do canhoto: contorno tracejado em três quadros. */
function Picote() {
  const d = [
    "M1.5 2 L50 1 L98.5 2.2 L98 50 L98.8 98 L50 99 L1.2 98.2 L2 50 Z",
    "M2 1.2 L50 2 L98 1.5 L98.8 50 L98 98.6 L50 98.2 L1.8 99 L1.2 50 Z",
    "M1.2 1.6 L50 1.4 L98.8 1.2 L98.2 50 L98.4 98.2 L50 99.2 L2 98.6 L1.6 50 Z",
  ];
  return (
    <span aria-hidden="true" className={cx("m-contorno", s.picote)}>
      {d.map((p) => (
        <svg key={p} className="m-f" viewBox="0 0 100 100" preserveAspectRatio="none">
          <path d={p} vectorEffect="non-scaling-stroke" />
        </svg>
      ))}
    </span>
  );
}

/** Âncora sem mexer no hash (o hash abre esta tela). */
const jump = (id: string) => (e: { preventDefault(): void }) => {
  e.preventDefault();
  document.getElementById(id)?.scrollIntoView({ behavior: "smooth" });
};

const Durex = ({ at }: { at: "e" | "d" | "c" }) => <span aria-hidden="true" className={cx(s.durex, s[at])} />;

function Placa({ id, children, tilt = 0 }: { id: string; children: string; tilt?: number }) {
  return (
    <div className={s.sombra}>
      <h2 id={id} className={s.placa} style={{ rotate: `${tilt}deg` }}>
        {children}
      </h2>
    </div>
  );
}

type BodyProps = {
  layout: Layout;
  tax: number;
  role: RoleId;
  /** Animações de espera ligada (a chave Movimento). */
  idle: boolean;
  onTax(t: number): void;
  onRole(r: RoleId): void;
  onIdle(on: boolean): void;
};

/** Corpo puro (sem estado): testável em Node. */
export function HowToBody({ layout, tax, role, idle, onTax, onRole, onIdle }: BodyProps) {
  const room = { tax, doubledCardValues: {} };
  const taxLabel = signed(tax);
  const sel = sheetOf(role, tax);
  const { general } = CLASSIC_RULES;
  const take = basicPrice("TAKE_COINS", room, 0).coins;
  const rows: [string, string][] = [
    [h.fields.quando, sel.quando],
    [h.fields.alvo, sel.alvo],
    [h.fields.faixa, sel.faixa],
    [h.fields.barra, sel.barra],
    [h.fields.contesta, sel.contesta],
  ];

  return (
    <div className={cx(s.screen, layout === "desktop" && s.desk, !idle && s.pausa)}>
      <header className={s.barra}>
        <a className={s.back} href="#">
          {h.back}
        </a>
        <nav aria-label={h.navLabel} className={s.nav}>
          {h.nav.map(([id, label]) => (
            <a key={id} href="#como-se-joga" onClick={jump(`ht-${id}`)}>
              {label}
            </a>
          ))}
        </nav>
        <button type="button" role="switch" aria-checked={idle} className={s.chave} onClick={() => onIdle(!idle)}>
          <span aria-hidden="true" className={s.trilho}>
            <span className={s.pino} />
          </span>
          {h.motion(idle)}
        </button>
      </header>

      <main className={s.page}>
        <section aria-labelledby="ht-titulo" className={s.hero}>
          <div className={s.heroText}>
            <span className={s.fita}>{h.mode}</span>
            <div className={s.sombra}>
              <h1 id="ht-titulo" className={cx(s.placa, s.titulo)}>
                {h.title}
              </h1>
            </div>
            <p className={s.lead}>{h.intro}</p>
          </div>
          <div className={cx("m-q3", s.folha, s.resumo)}>
            <Durex at="e" />
            <Durex at="d" />
            <h2 className={s.marker}>{h.summaryTitle}</h2>
            <dl>
              {h.summary.map(([n, t]) => (
                <div key={t}>
                  <dt>{n}</dt>
                  <dd>{t}</dd>
                </div>
              ))}
            </dl>
          </div>
        </section>

        <section id="ht-vez" aria-labelledby="ht-vez-h" className={s.block}>
          <Placa id="ht-vez-h" tilt={1}>{h.roundTitle}</Placa>
          <ol className={s.cards}>
            {h.round.map(([title, text], i) => (
              <li key={title} className={cx("m-q3", s.folha, s.step)}>
                <Durex at="c" />
                <span className={s.stepN}>{i + 1}</span>
                <h3>{title}</h3>
                <p>{text}</p>
              </li>
            ))}
          </ol>
        </section>

        <section id="ht-acoes" aria-labelledby="ht-acoes-h" className={s.block}>
          <div className={s.headRow}>
            <Placa id="ht-acoes-h" tilt={-1}>{h.actionsTitle}</Placa>
            <p className={s.rabisco}>
              {h.actionsScribble(taxLabel)}
              <Arrow />
            </p>
          </div>
          <div className={s.cards}>
            <article className={cx("m-q3", s.folha, s.action)}>
              <Durex at="e" />
              <div className={s.actionHead}>
                <h3>{h.take}</h3>
                <span className={s.num}>+{take}</span>
              </div>
              <p>{h.takeText}</p>
            </article>
            <article className={cx("m-q3", s.folha, s.action)}>
              <Durex at="d" />
              <div className={s.actionHead}>
                <h3>{h.kill}</h3>
                <span className={s.num}>{killCost(tax)}</span>
              </div>
              <p>{h.killText}</p>
              <ul aria-label={h.killTaxLabel} className={s.killTax}>
                {h.killTax.map((t) => (
                  <li key={t}>{t}</li>
                ))}
              </ul>
            </article>
            <article className={cx("m-q3", s.folha, s.action)}>
              <Durex at="e" />
              <div className={s.actionHead}>
                <h3>{h.swap}</h3>
                <span className={s.num}>½</span>
              </div>
              <p>{h.swapText}</p>
            </article>
            <article className={cx("m-q3", "m-canhoto", s.folha, s.action, s.canhoto)}>
              <Picote />
              <div className={s.actionHead}>
                <h3>{h.use}</h3>
                <span className={s.num}>?</span>
              </div>
              <p>
                {h.useText[0]}
                <a href="#como-se-joga" onClick={jump("ht-elenco")}>
                  {h.useText[1]}
                </a>
                {h.useText[2]}
              </p>
            </article>
          </div>
        </section>

        <section id="ht-duvidar" aria-labelledby="ht-duvidar-h" className={s.split}>
          <div className={s.splitText}>
            <Placa id="ht-duvidar-h" tilt={1.5}>{h.doubtTitle}</Placa>
            <p className={s.lead2}>{h.doubtLead}</p>
            <p>{h.doubtBlock}</p>
          </div>
          <div className={cx(s.cards, s.doubtCards)}>
            <article className={cx("m-q3", s.folha, s.doubt)}>
              <Durex at="e" />
              <Durex at="d" />
              <span className={cx(s.carimbo, s.blefe)}>{h.bluffStamp}</span>
              <h3>{h.bluffTitle}</h3>
              <p>{h.bluffText}</p>
            </article>
            <article className={cx("m-q3", s.folha, s.doubt)}>
              <Durex at="e" />
              <Durex at="d" />
              <span className={cx(s.carimbo, s.verdade)}>{h.truthStamp}</span>
              <h3>{h.truthTitle}</h3>
              <p>{h.truthText}</p>
            </article>
          </div>
        </section>

        <section id="ht-taxa" aria-labelledby="ht-taxa-h" className={cx(s.split, s.center)}>
          <div className={s.splitText}>
            <Placa id="ht-taxa-h" tilt={-1}>{h.taxTitle}</Placa>
            <p className={s.lead2}>{h.taxLead}</p>
            <p>{h.taxNote}</p>
          </div>
          <div className={cx("m-q3", s.folha, s.taxa)}>
            <Durex at="c" />
            <label htmlFor="ht-taxa-range" className={s.marker}>
              {h.taxTry}
            </label>
            <output htmlFor="ht-taxa-range" className={s.taxaNum}>
              {taxLabel}
            </output>
            <input
              id="ht-taxa-range"
              type="range"
              min={general.minTax}
              max={general.maxTax}
              step={1}
              value={tax}
              onChange={(e) => onTax(Number(e.target.value))}
            />
            <span className={s.hint}>{h.taxHint}</span>
          </div>
        </section>

        <section id="ht-elenco" aria-labelledby="ht-elenco-h" className={s.block}>
          <div className={s.headRow}>
            <Placa id="ht-elenco-h" tilt={1}>{h.rolesTitle}</Placa>
            <p className={s.rabisco}>
              {h.rolesScribble(taxLabel)}
              <Arrow />
            </p>
          </div>
          <div className={s.elenco}>
            <nav aria-label={h.rolesPick} className={s.lista}>
              {GROUPS.map((ids, g) => (
                <div key={g} className={s.grupo}>
                  <span className={s.grupoTitle}>{h.groups[g]}</span>
                  {ids.map((id) => {
                    const r = sheetOf(id, tax);
                    return (
                      <button key={id} type="button" aria-pressed={id === role} className={s.etiqueta} onClick={() => onRole(id)}>
                        <span aria-hidden="true" className={s.furo} />
                        <span className={s.etiquetaNome}>{r.name}</span>
                        <span className={s.etiquetaPreco}>{r.price}</span>
                      </button>
                    );
                  })}
                </div>
              ))}
            </nav>

            <article aria-live="polite" aria-label={sel.name} className={cx("m-q3", s.folha, s.ficha)}>
              <Durex at="e" />
              <Durex at="d" />
              <figure className={s.figura}>
                <div role="img" aria-label={h.artLabel(sel.name)} className={s.quadro} data-arte={`/arte/papeis/${role}/`}>
                  <span className={s.tacha} style={{ top: 8, left: 8 }} />
                  <span className={s.tacha} style={{ top: 8, right: 8 }} />
                  <span className={s.tacha} style={{ bottom: 8, left: 8 }} />
                  <span className={s.tacha} style={{ bottom: 8, right: 8 }} />
                  <div className={s.tela}>
                    <Ask />
                    <span className={s.arteNota}>{h.artSoon}</span>
                  </div>
                </div>
                <figcaption>
                  <span>papeis/{role.toLowerCase()}/</span>
                  <span>{h.artFrames}</span>
                </figcaption>
              </figure>
              <div className={s.fichaText}>
                <span className={s.fita}>{h.groups[sel.group]}</span>
                <div className={s.fichaHead}>
                  <h3>{sel.name}</h3>
                  <span aria-label={h.priceAria(sel.price)} className={s.moeda}>
                    {sel.price}
                  </span>
                </div>
                <p className={s.efeito}>{sel.efeito}</p>
                <dl className={s.campos}>
                  {rows.map(([k, v]) => (
                    <div key={k}>
                      <dt>{k}</dt>
                      <dd>{v}</dd>
                    </div>
                  ))}
                </dl>
                {sel.nota && <p className={s.nota}>☞ {sel.nota}</p>}
              </div>
            </article>
          </div>
        </section>

        <section id="ht-casa" aria-labelledby="ht-casa-h" className={s.split}>
          <div className={cx("m-q3", s.folha, s.casa)}>
            <Durex at="e" />
            <Durex at="d" />
            <h2 id="ht-casa-h" className={s.marker}>
              {h.houseTitle}
            </h2>
            <ul>
              {h.house.map(([a, b, c]) => (
                <li key={b}>
                  {a}
                  <b>{b}</b>
                  {c}
                </li>
              ))}
            </ul>
          </div>
          <div className={cx(s.sombra, s.fimWrap)}>
            <div className={cx(s.placa, s.fim)}>
              <span className={s.fimLabel}>{h.endLabel}</span>
              <h3>{h.endText}</h3>
            </div>
          </div>
        </section>

        <footer className={s.footer}>
          <p className={s.pronto}>{h.ready}</p>
          <a className={s.criar} href="#">
            {h.create}
          </a>
        </footer>
      </main>
    </div>
  );
}

/** Tela com estado local: a taxa de teste não mexe na partida. */
export function HowTo() {
  const layout = useLayout();
  const motion = useMotionPref();
  const [tax, setTax] = useState(0);
  const [role, setRole] = useState<RoleId>("GUARDIAN");
  // no celular as etiquetas rolam de lado: mantém a escolhida à vista
  useEffect(() => {
    const on = document.querySelector<HTMLElement>("#ht-elenco nav [aria-pressed='true']");
    const bar = on?.parentElement?.parentElement;
    if (on && bar && bar.scrollWidth > bar.clientWidth) bar.scrollTo({ left: on.offsetLeft - bar.clientWidth / 2 + on.offsetWidth / 2, behavior: "smooth" });
  }, [role]);
  return <HowToBody layout={layout} tax={tax} role={role} idle={motion.idle} onTax={setTax} onRole={setRole} onIdle={motion.setIdle} />;
}
