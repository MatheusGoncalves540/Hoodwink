import type { Layout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import { roleName } from "../../game/roles";
import type { GameView } from "../../game/selectors";
import { selectYou } from "../../game/selectors";
import { Button } from "../../ui/Button/Button";
import { Ticket } from "../../ui/Ticket/Ticket";
import { cx } from "../../ui/cx";
import s from "./OutOfScene.module.css";

type Props = {
  view: GameView;
  layout: Layout;
  onChat(): void;
  onLeave(): void;
};

/** Você saiu de cena (tela 20): a mesa continua acima; o painel substitui a faixa Você. */
export function OutOfScene({ view, layout, onChat, onLeave }: Props) {
  const you = selectYou(view);
  const exit = view.log.find((e) => e.key === `EXIT|${view.me}`);
  const names = you.cards.filter((c) => c.role).map((c) => roleName(c.role!));

  return (
    <section aria-label={copy.outOfSceneRegion} className={cx(s.panel, layout === "desktop" ? s.deskPanel : s.sheet)}>
      <div className={s.head}>
        <h2 className={s.title}>{copy.outOfScene}</h2>
        <span role="img" aria-label={copy.deadCardsLabel(names)} className={s.deadCards}>
          {you.cards.map((c, i) => (
            <span key={i} className={s.deadCard}>
              {c.role ? roleName(c.role) : "?"}
            </span>
          ))}
        </span>
      </div>
      <p className={s.text}>{copy.outOfSceneAt(exit?.turn ?? view.room?.turn ?? 0)}</p>
      <div className={s.actions}>
        <Ticket onClick={onChat}>{copy.openChat}</Ticket>
        <Button variant="link" onClick={onLeave}>
          {copy.leaveRoom}
        </Button>
      </div>
    </section>
  );
}
