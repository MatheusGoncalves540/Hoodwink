/**
 * Como se joga (fase 7): tela de bastidor em papelão. Um componente; layout por useLayout.
 * Movimento: só o papel dos cartões (.m-q3); texto, placas e rabiscos parados. Sem respiração aqui (trava 5).
 * A chave "Movimento" grava a preferência Animações de espera (a mesma do menu), não uma própria.
 */
import { useState } from "react";
import type { Layout } from "../../app/useLayout";
import { useLayout } from "../../app/useLayout";
import { useMotionPref } from "../../app/useMotionPref";
import { copy } from "../../game/copy";
import { basicPrice, guardianPrice, priceOf, type PriceRoom } from "../../game/pricing";
import { ROLE_IDS, type RoleId } from "../../game/protocol";
import { ROLES } from "../../game/roles";
import { CLASSIC_RULES } from "../../game/rules.classic";
import { cx } from "../../ui/cx";
import s from "./HowTo.module.css";

const h = copy.howto;

/** Três desenhos da mesma seta, para a tira (rabisco tremendo). */
function Arrow() {
  const paths = ["M4 20 C 18 6, 32 6, 44 16 M36 9 L45 16 L35 21", "M4 21 C 17 8, 33 5, 44 15 M35 8 L44 15 L36 22", "M5 19 C 19 7, 31 7, 43 17 M36 10 L44 17 L34 20"];
  return (
    <span aria-hidden="true" className={cx("m-tira", s.arrow)}>
      <span>
        {paths.map((d) => (
          <svg key={d} viewBox="0 0 48 28">
            <path d={d} />
          </svg>
        ))}
      </span>
    </span>
  );
}

function fichaOf(role: RoleId, room: PriceRoom) {
  const p = priceOf(role, room);
  const sel = priceOf("SELFISH", room);
  const inv = priceOf("INVESTOR", room);
  return {
    label: copy.priceLabel(p),
    text: h.ficha(role, {
      cost: p.coins,
      gain: p.coins,
      steals: p.kind === "gain" ? (p.steals ?? p.coins) : 0,
      cap: inv.kind === "invest" ? inv.cap : 0,
      threshold: sel.kind === "reaction" ? Math.max(1, sel.threshold) : 0,
      cardChange: sel.kind === "reaction" ? sel.cardChange : 0,
      guard: [guardianPrice("reaction", room), guardianPrice("clairvoyant", room), guardianPrice("ownTurn", room)],
    }),
  };
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
  const room: PriceRoom = { tax, doubledCardValues: {} };
  const ficha = fichaOf(role, room);
  const { general } = CLASSIC_RULES;
  return (
    <main className={cx(s.screen, layout === "desktop" && s.desk, !idle && s.pausa)}>
      <header className={s.barra}>
        <a className={s.back} href="#">
          {h.back}
        </a>
        <h1 className={s.placa}>{h.title}</h1>
        <button type="button" role="switch" aria-checked={idle} className={s.chave} onClick={() => onIdle(!idle)}>
          <span>{h.motion}</span>
          <span aria-hidden="true" className={s.knob} />
        </button>
      </header>

      <div className={s.grid}>
        <section aria-labelledby="ht-goal" className={cx("m-q3", s.folha)}>
          <span className={s.fita}>{h.mode}</span>
          <h2 id="ht-goal" className={s.cardTitle}>{h.goalTitle}</h2>
          {h.goal.map((t) => (
            <p key={t}>{t}</p>
          ))}
          <p>{h.bluff}</p>
        </section>

        <section aria-labelledby="ht-turn" className={cx("m-q3", s.folha)}>
          <h2 id="ht-turn" className={s.cardTitle}>{h.turnTitle}</h2>
          <p>{h.turnLead}</p>
          <ul className={s.list}>
            <li>{h.take(basicPrice("TAKE_COINS", room, 0).coins)}</li>
            <li>{h.kill(basicPrice("KILL_CARD", room, 0).coins)}</li>
            <li>{h.swap}</li>
          </ul>
        </section>

        <section aria-labelledby="ht-use" className={cx("m-q3", "m-canhoto", "m-contorno", s.folha, s.canhoto)}>
          {[0, 1, 2].map((i) => (
            <svg key={i} aria-hidden="true" className={cx("m-f", s.picote)} preserveAspectRatio="none" viewBox="0 0 100 100">
              <rect x="3" y="3" width="94" height="94" rx={2 + i} vectorEffect="non-scaling-stroke" />
            </svg>
          ))}
          <h2 id="ht-use" className={s.cardTitle}>{h.useTitle}</h2>
          {h.use.map((t) => (
            <p key={t}>{t}</p>
          ))}
        </section>

        <section aria-labelledby="ht-doubt" className={cx("m-q3", s.folha)}>
          <h2 id="ht-doubt" className={s.cardTitle}>{h.doubtTitle}</h2>
          <p>{h.doubtHad}</p>
          <p>{h.doubtNot}</p>
        </section>

        <section aria-labelledby="ht-tax" className={cx("m-q3", s.folha)}>
          <h2 id="ht-tax" className={s.cardTitle}>{h.taxTitle}</h2>
          <p>{h.tax}</p>
          <p>{h.limit}</p>
          <label className={s.taxa}>
            <span>{h.taxTry}</span>
            <input
              type="range"
              min={general.minTax}
              max={general.maxTax}
              step={1}
              value={tax}
              onChange={(e) => onTax(Number(e.target.value))}
            />
            <b className={s.marcador}>{h.taxNow(tax)}</b>
          </label>
        </section>

        <p className={s.rabisco}>
          {h.scribble} <Arrow />
        </p>

        <section aria-labelledby="ht-roles" className={s.roles}>
          <h2 id="ht-roles" className={s.placa}>{h.rolesTitle}</h2>
          <div role="group" aria-label={h.rolesPick} className={s.etiquetas}>
            {ROLE_IDS.map((id) => (
              <button key={id} type="button" aria-pressed={id === role} className={s.etiqueta} onClick={() => onRole(id)}>
                {ROLES[id].name}
              </button>
            ))}
          </div>
          <article aria-label={ROLES[role].name} className={cx("m-q3", s.folha, s.ficha)}>
            <div className={s.arte} data-arte={`/arte/papeis/${role}/`}>
              <span className={s.arteNota}>{h.artSoon}</span>
            </div>
            <div>
              <h3 className={s.cardTitle}>{ROLES[role].name}</h3>
              <p className={s.marcador}>{h.price(ficha.label)}</p>
              <p>{ficha.text}</p>
            </div>
          </article>
        </section>
      </div>
    </main>
  );
}

/** Tela com estado local: a taxa de teste não mexe na partida. */
export function HowTo() {
  const layout = useLayout();
  const motion = useMotionPref();
  const [tax, setTax] = useState(0);
  const [role, setRole] = useState<RoleId>("ASSASSIN");
  return <HowToBody layout={layout} tax={tax} role={role} idle={motion.idle} onTax={setTax} onRole={setRole} onIdle={motion.setIdle} />;
}
