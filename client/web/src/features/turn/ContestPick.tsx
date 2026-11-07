import type { Layout } from "../../app/useLayout";
import { copy, firstName } from "../../game/copy";
import type { RoleId } from "../../game/protocol";
import { roleName } from "../../game/roles";
import type { Decision, GameView, MiniCard } from "../../game/selectors";
import { selectSeatSheet, selectYou, selectYouSheet } from "../../game/selectors";
import { CountdownDots } from "../../ui/CountdownDots/CountdownDots";
import { useCountdown } from "../../ui/Timer/timer";
import { cx } from "../../ui/cx";
import s from "./ContestPick.module.css";

type PickDecision = Extract<Decision, { kind: "pickCardToKill" } | { kind: "waitPick" }>;
type PickCard = { index: number; state: MiniCard["state"]; role?: RoleId };

type Props = {
  view: GameView;
  decision: PickDecision;
  layout: Layout;
  onPick(index: number): void;
};

/** Contestação resolvida (telas 18 e 19): escolher a carta que cai ou ver a escolha do outro. */
export function ContestPick({ view, decision, layout, onPick }: Props) {
  const ev = view.room?.gameEvent;
  const total = selectYou(view).timer?.total ?? 6;
  // mesma contagem do Timer: os pontos e o texto mudam no segundo cheio e zeram no prazo
  const { left } = useCountdown(ev?.expiresAt, total);

  const picking = decision.kind === "pickCardToKill";
  const target = picking && decision.targetId ? selectSeatSheet(view, decision.targetId) : null;
  const you = !picking ? selectYouSheet(view) : null;
  const chooser = !picking && decision.kind === "waitPick" ? view.room?.players[decision.chooserId]?.name : "";
  const chooserName = firstName(chooser || "");
  const name = firstName(target?.name ?? "");
  const cards: PickCard[] = picking
    ? (target?.cards ?? []).map((c) => ({ index: c.index, state: c.state, role: c.role }))
    : (you?.cards ?? []).map((c) => ({
        index: c.index,
        state: c.dead ? "dead" : c.protected ? "protected" : "back",
        role: c.role,
      }));

  if (picking ? !target : !you) return null;

  return (
    <section
      aria-label={picking ? copy.chooseFallingCard : copy.choosesYourCard(chooserName)}
      className={cx(s.panel, layout === "desktop" ? s.deskPanel : s.sheet)}
    >
      <h2 className={s.title}>{picking ? copy.dropCard(name) : copy.choosesYourCard(chooserName)}</h2>
      {!picking && <p className={s.hint}>{copy.waitPickHint}</p>}

      <div role="group" aria-label={picking ? copy.cardsOf(name) : copy.yourCards} className={s.cards}>
        {cards.map((c) => {
          if (picking && c.state === "dead") return null;
          // na espera, as suas cartas são conhecidas: mostram a face mesmo sem estarem reveladas à mesa
          const face = c.state === "revealed" || c.state === "protected" || (!picking && !!c.role);
          const note = !picking && c.state === "protected" ? copy.protectedWarn : undefined;
          const state = picking
            ? copy.cardState[c.state === "back" ? "hidden" : c.state]
            : c.state === "dead"
              ? copy.cardState.dead
              : c.state === "protected"
                ? copy.cardState.protected
                : c.state === "revealed"
                  ? copy.cardState.revealed
                  : copy.inHandState;
          const body = face ? (
            <>
              <span className={s.state}>{state}</span>
              <span className={s.topBand}>
                <span className={s.cardName}>{c.role ? roleName(c.role) : "?"}</span>
                {note && <span className={s.note}>{note}</span>}
              </span>
            </>
          ) : (
            <>
              <span className={s.state}>{c.state === "dead" ? copy.cardState.dead : state}</span>
              <span className={s.cardName}>{c.role ? roleName(c.role) : "?"}</span>
            </>
          );

          if (!picking) {
            return (
              <div key={c.index} className={cx(face && "hw-carta", face && s.faceCard, s.card, !face && c.state === "dead" && s.deadCard, !face && c.state !== "dead" && s.hiddenCard)}>
                {body}
              </div>
            );
          }

          const label = face && c.role ? copy.dropRevealed(name, roleName(c.role)) : copy.dropHidden(name);
          return (
            <button
              key={c.index}
              type="button"
              className={cx(face && "hw-carta", face && s.faceCard, s.card, s.pickable, !face && s.hiddenCard)}
              aria-label={label}
              onClick={() => onPick(c.index)}
            >
              {body}
            </button>
          );
        })}
      </div>

      <div className={s.foot}>
        <CountdownDots seconds={left} total={total} />
        <span className={s.footText}>{picking ? copy.sorteEscolhe(left) : copy.sorteEscolheOther(chooserName, left)}</span>
      </div>
    </section>
  );
}
