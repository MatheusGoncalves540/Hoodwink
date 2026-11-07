import type { Layout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import { CLASSIC_RULES } from "../../game/rules.classic";
import type { selectYou } from "../../game/selectors";
import { CardBack, CardDead, CardFace } from "../../ui/Card/Card";
import { Coin } from "../../ui/Coin/Coin";
import { Timer } from "../../ui/Timer/Timer";
import { cx } from "../../ui/cx";
import s from "./YouBar.module.css";

/** Faixa "Você": 236 px no mobile e linha de 196 px no desktop. */
export function YouBar({ vm, suddenDeath, layout }: { vm: ReturnType<typeof selectYou>; suddenDeath: boolean; layout: Layout }) {
  const timer = vm.timer;
  const turn = timer ? copy.turnOf(timer.ownerName) : null;
  const seconds = CLASSIC_RULES.waitingAction.suddenDeath;
  const cardSize = layout === "desktop" ? "d" : "s";

  return (
    <section aria-label={copy.youRegion} className={cx(s.bar, layout === "desktop" && s.desk, suddenDeath && s.sudden)}>
      <div className={s.top}>
        <div className={s.cards}>
          {vm.cards.map((c) =>
            c.dead ? (
              <CardDead key={c.index} size={cardSize} role={c.role} className={s.card} />
            ) : c.role ? (
              <CardFace
                key={c.index}
                size={cardSize}
                role={c.role}
                protected={c.protected}
                bottom={c.protected ? copy.cardState.protected : undefined}
                className={s.card}
              />
            ) : (
              <CardBack key={c.index} size={cardSize} className={s.card} />
            ),
          )}
        </div>
        <div className={s.money}>
          <Coin size={60} value={vm.coins} label={copy.coins(vm.coins)} />
          <span className={s.meta}>
            {copy.limit(vm.limit)}
            {vm.investments.length > 0 && (
              <>
                <br />
                {copy.investments(vm.investments.length, vm.nextInvestment)}
              </>
            )}
          </span>
        </div>
      </div>

      {timer && turn && (
        <Timer
          expiresAt={timer.expiresAt}
          total={timer.total}
          label={copy.timerLabel(timer.ownerName)}
          tone={suddenDeath ? "alerta" : "tinta"}
          caption={
            <span>
              {turn.before}
              <b>{turn.name}</b>
              {suddenDeath ? copy.suddenTurnAfter(seconds) : turn.after}
            </span>
          }
        />
      )}
    </section>
  );
}
