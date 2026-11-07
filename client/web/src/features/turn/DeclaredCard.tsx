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

/** Palco da reação/contestação (Main, ContestaBlefe e ContestaVerdade) e da carta revelada (Clarividente). */
export function DeclaredCard({ view, layout }: { view: GameView; layout: Layout }) {
  const stage = selectStage(view);
  const desktop = layout === "desktop";
  const size = { width: desktop ? 132 : 116, height: desktop ? 186 : 162 };

  if (stage.revealed) {
    const { actorName, targetId, targetName, card, mine, toAll } = stage.revealed;
    const revealedName = firstName(actorName);
    const role = card ? roleName(card) : null;
    const cardPhrase = copy.possessive(targetName, targetId === view.me);
    return (
      <section aria-label={copy.stage} aria-live="assertive" className={cx(s.stage, desktop && s.desk)}>
        <figure className={cx(s.figure, s.revealFigure)}>
          {card ? (
            <CardFace as="div" size="xl" role={card} className={cx(s.face, s.reveal)} style={size} />
          ) : (
            <div className={s.dashed} style={size}>
              <span className={s.art} role="img" aria-label={copy.cardArt(role ?? "carta")}>
                ?
              </span>
              <span className={s.dashedName}>{targetName}</span>
            </div>
          )}
        </figure>
        <div className={cx("m-rise", s.text, s.textLate)}>
          <p className={s.lead}>{copy.revealedLead(revealedName, cardPhrase)}</p>
          <h1 className={s.role}>{role ?? "?"}</h1>
          <p className={s.effect}>{toAll ? copy.revealedToAll : mine ? copy.revealedPrivateMe : copy.revealedHidden}</p>
        </div>
      </section>
    );
  }

  const declared = stage.declared;
  if (!declared) return null;

  const { actorId, actorName, role, phase, targetId, targetName, contesterId, contesterName, chooserName } = declared;
  const name = firstName(actorName);
  const roleLabel = roleName(role);
  const price = view.room ? priceOf(role, view.room) : ({ kind: "cost", coins: 0 } as const);
  const value = CLASSIC_RULES.cards[role].value ?? 0;
  const truth = phase === "truth";
  const contested = phase === "contested" || phase === "bluff" || truth;
  const effect = truth
    ? copy.truthEffect(name, copy.possessive(contesterName ?? "", contesterId === view.me))
    : phase === "bluff"
      ? copy.bluffEffect(name, roleLabel, chooserName ?? "", copy.possessive(actorName, actorId === view.me))
      : phase === "contested"
        ? copy.doubtedBy(contesterName ?? "")
        : copy.declaredEffect(name, role, price, targetId ? { name: targetName ?? "", mine: targetId === view.me } : null);

  return (
    <section aria-label={copy.stage} aria-live="assertive" className={cx(s.stage, desktop && s.desk)}>
      <figure className={cx(s.figure, contested && "m-hit")}>
        {truth ? (
          <CardFace as="div" size="xl" role={role} className={cx("m-flip", s.face)} style={size} />
        ) : (
          <div className={cx(s.dashed, phase === "declared" && "m-drop")} style={size}>
            {phase === "declared" && (
              <span className={s.value} aria-label={copy.baseValue(value)}>
                {value}
              </span>
            )}
            <span className={s.art} role="img" aria-label={copy.cardArt(roleLabel)}>
              {copy.cardArt(roleLabel)}
            </span>
            <span className={s.dashedName}>{roleLabel}</span>
          </div>
        )}
        {contested && (
          <span className={cx("m-stamp", s.stamp, truth && s.truthStamp)}>{truth ? copy.stampTruth : copy.stampBluff}</span>
        )}
        {phase === "declared" && <figcaption className={s.caption}>{copy.declaredNotRevealed}</figcaption>}
      </figure>

      <div className={cx("m-rise", s.text, contested && s.textLate)}>
        <p className={s.lead}>{phase === "declared" ? copy.declaredLead(name) : truth ? copy.truthLead(name) : copy.contestedLead(name)}</p>
        <h1 className={s.role}>{roleLabel}</h1>
        {effect && <p className={s.effect}>{effect}</p>}
      </div>
    </section>
  );
}
