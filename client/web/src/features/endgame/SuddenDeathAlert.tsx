import { Fragment } from "react";
import type { Layout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import { cx } from "../../ui/cx";
import { CURTAIN_DESK_HTML, CURTAIN_HTML } from "./sudden";
import s from "./SuddenDeathAlert.module.css";

type Props = {
  layout: Layout;
  /** Segundos de cada deixa na morte súbita (5). */
  seconds: number;
};

/** Aviso de morte súbita (tela 21): véu escuro, holofote e a cortina em chamas. */
export function SuddenDeathAlert({ layout, seconds }: Props) {
  const desktop = layout === "desktop";
  return (
    <div role="alert" data-theme="dark" className={cx("m-scrim", s.alert, desktop && s.desk)} style={{ animationDelay: "2s" }}>
      <div aria-hidden="true" className="ms-clarao brilho" />
      <div aria-hidden="true" className={cx("ms-holofote", desktop && s.deskHolofote)} />

      <div className={s.text}>
        <p className={s.deckOver}>{copy.deckOver}</p>
        <h1 className={cx("m-slam", s.title)} style={{ animationDelay: "2.45s" }}>
          {copy.suddenDeathLines.map((line, i) => (
            <Fragment key={line}>
              {i > 0 && <br />}
              {line}
            </Fragment>
          ))}
        </h1>
        <p className={s.each}>{copy.eachCue(seconds)}</p>
      </div>

      <p className={cx("m-rise", s.resume)} style={{ animationDelay: "2.9s" }}>{copy.playResumes(4)}</p>

      <div aria-hidden="true" className={s.curtain} dangerouslySetInnerHTML={{ __html: desktop ? CURTAIN_DESK_HTML : CURTAIN_HTML }} />
    </div>
  );
}
