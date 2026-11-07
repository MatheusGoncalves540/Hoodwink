import type { Layout } from "../../app/useLayout";
import { copy, firstName } from "../../game/copy";
import { priceOf } from "../../game/pricing";
import { CLASSIC_RULES } from "../../game/rules.classic";
import { roleName } from "../../game/roles";
import type { GameView } from "../../game/selectors";
import { selectStage } from "../../game/selectors";
import { CardFace } from "../../ui/Card/Card";
import { cx } from "../../ui/cx";
import s from "./DeclaredCard.module.css";

/** Palco da reação/contestação (Main, ContestaBlefe e ContestaVerdade): carta declarada, carimbo e texto. */
export function DeclaredCard({ view, layout }: { view: GameView; layout: Layout }) {
  const declared = selectStage(view).declared;
  if (!declared) return null;

  const desktop = layout === "desktop";
  const { actorName, role, phase } = declared;
  const name = firstName(actorName);
  const price = view.room ? priceOf(role, view.room) : ({ kind: "cost", coins: 0 } as const);
  const value = CLASSIC_RULES.cards[role].value ?? 0;
  const truth = phase === "truth";
  const contested = phase === "contested" || phase === "bluff" || truth;
  const targetsMe = !!view.me && declared.targetId === view.me;
  const effect = truth
    ? copy.truthReturn(name)
    : phase === "bluff"
      ? copy.bluffed(name, roleName(role)) + copy.bluffFall(name)
      : phase === "contested"
        ? ""
        : copy.declaredEffect(name, role, price, targetsMe);

  return (
    <section aria-label={copy.stage} aria-live="assertive" className={cx(s.stage, desktop && s.desk)}>
      <figure className={cx(s.figure, contested && "m-hit")}>
        {truth ? (
          <CardFace
            as="div"
            size="xl"
            role={role}
            className={cx("m-flip", s.face)}
            style={{ width: desktop ? 132 : 116, height: desktop ? 186 : 162 }}
          />
        ) : (
          <div className={cx(s.dashed, phase === "declared" && "m-drop")} style={{ width: desktop ? 132 : 116, height: desktop ? 186 : 162 }}>
            {phase === "declared" && (
              <span className={s.value} aria-label={copy.baseValue(value)}>
                {value}
              </span>
            )}
            <span className={s.art} role="img" aria-label={copy.cardArt(roleName(role))}>
              {copy.cardArt(roleName(role))}
            </span>
            <span className={s.dashedName}>{roleName(role)}</span>
          </div>
        )}
        {contested && (
          <span className={cx("m-stamp", s.stamp, truth && s.truthStamp)}>{truth ? copy.stampTruth : copy.stampBluff}</span>
        )}
        {phase === "declared" && <figcaption className={s.caption}>{copy.declaredNotRevealed}</figcaption>}
      </figure>

      <div className={cx("m-rise", s.text, contested && s.textLate)}>
        <p className={s.lead}>{phase === "declared" ? copy.declaredLead(name) : truth ? copy.truthLead(name) : copy.contestedLead(name)}</p>
        <h1 className={s.role}>{roleName(role)}</h1>
        {effect && <p className={s.effect}>{effect}</p>}
      </div>
    </section>
  );
}
