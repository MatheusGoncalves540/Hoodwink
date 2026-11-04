import { copy } from "../../game/copy";
import type { selectYou } from "../../game/selectors";
import { CardBack, CardDead, CardFace } from "../../ui/Card/Card";
import { Coin } from "../../ui/Coin/Coin";
import { Timer } from "../../ui/Timer/Timer";
import s from "./YouBar.module.css";

/** Faixa "Você" (236 px): suas cartas, moeda/limite/investimentos e o timer da vez. */
export function YouBar({ vm, suddenDeath }: { vm: ReturnType<typeof selectYou>; suddenDeath: boolean }) {
  const timer = vm.timer;
  const turn = timer ? copy.turnOf(timer.ownerName) : null;

  return (
    <section aria-label={copy.youRegion} className={s.bar}>
      <div className={s.top}>
        <div className={s.cards}>
          {vm.cards.map((c, i) =>
            c.dead ? (
              <CardDead key={i} size="s" role={c.role} className={s.card} />
            ) : c.role ? (
              <CardFace
                key={i}
                size="s"
                role={c.role}
                protected={c.protected}
                bottom={c.protected ? copy.cardState.protected : undefined}
                className={s.card}
              />
            ) : (
              <CardBack key={i} size="s" className={s.card} />
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
              {turn.after}
            </span>
          }
        />
      )}
    </section>
  );
}
