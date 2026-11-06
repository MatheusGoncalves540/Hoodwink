import type { Layout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import { roleName } from "../../game/roles";
import type { Decision, GameView } from "../../game/selectors";
import { selectYou } from "../../game/selectors";
import { Button } from "../../ui/Button/Button";
import { Coin } from "../../ui/Coin/Coin";
import { Ticket } from "../../ui/Ticket/Ticket";
import { Timer } from "../../ui/Timer/Timer";
import { cx } from "../../ui/cx";
import s from "./ReactionPanel.module.css";

type ReactDecision = Extract<Decision, { kind: "react" }>;

type Props = {
  view: GameView;
  decision: ReactDecision;
  layout: Layout;
  onContest(): void;
  onBlock(): void;
  onPass(): void;
  /** Sua carta Kamikaze está morrendo: escolher quem perde uma carta junto. */
  onKamikaze?(): void;
};

/** Sua deixa (tela 5): bloquear, contestar ou deixar passar, com o tempo correndo. */
export function ReactionPanel({ view, decision, layout, onContest, onBlock, onPass, onKamikaze }: Props) {
  const you = selectYou(view);
  const hand = you.cards.filter((c) => !c.dead && c.role).map((c) => roleName(c.role!));
  const block = decision.block;
  const desktop = layout === "desktop";

  return (
    <section aria-label={copy.reactionLabel} className={cx(s.panel, desktop ? s.deskPanel : s.sheet)}>
      <div className={s.head}>
        <div className={s.titles}>
          <h2 className={s.title}>{copy.yourCue}</h2>
          <span className={s.hand}>{copy.hand(hand)}</span>
        </div>
        <Coin size={52} value={you.coins} label={copy.coins(you.coins)} />
      </div>

      {you.timer && <Timer expiresAt={you.timer.expiresAt} total={you.timer.total} label={copy.reactionTimer} decision />}

      {block && (
        <Ticket className={s.block} disabled={you.coins < block.cost} onClick={onBlock}>
          {copy.block(roleName(block.role))} <span className={s.blockCost}>{copy.blockCost(block.cost)}</span>
        </Ticket>
      )}

      {decision.kamikaze && (
        <Ticket className={s.block} onClick={onKamikaze}>
          {copy.kamikazeStrike}
        </Ticket>
      )}

      <div className={s.actions}>
        {decision.canContest && (
          <Button variant="danger" onClick={onContest}>
            {copy.contest}
          </Button>
        )}
        <Button variant="outline" onClick={onPass}>
          {copy.letPass}
        </Button>
      </div>
    </section>
  );
}
