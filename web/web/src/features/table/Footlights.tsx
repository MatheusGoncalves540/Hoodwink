import type { Layout } from "../../app/useLayout";
import { copy, signed } from "../../game/copy";
import { CLASSIC_RULES } from "../../game/rules.classic";
import type { selectFootlights } from "../../game/selectors";
import { cx } from "../../ui/cx";
import s from "./Footlights.module.css";

/** Ribalta: baralho, taxa e efeitos ativos (com pips de rodadas restantes). */
export function Footlights({ vm, layout }: { vm: ReturnType<typeof selectFootlights>; layout: Layout }) {
  const seconds = CLASSIC_RULES.waitingAction.suddenDeath;
  const deckName = vm.suddenDeath ? copy.deckEmpty : copy.deck;
  const deckValue = vm.suddenDeath ? `${seconds} s` : String(vm.deck);
  const deckLabel = vm.suddenDeath ? copy.deckEmptyLabel(seconds) : copy.deckLabel(vm.deck);

  return (
    <section aria-label={copy.footlights} className={cx(s.footlights, layout === "desktop" && s.desk)}>
      <div role="group" aria-label={copy.tableGroup} className={cx(s.group, s.tableGroup)}>
        <div aria-label={deckLabel} className={cx("m-item", s.item, vm.suddenDeath && s.alert)}>
          <span className={cx("hw-rotulo", s.label)}>{deckName}</span>
          <span className={s.value}>{deckValue}</span>
        </div>
        <div aria-label={copy.taxLabel(vm.tax)} className={cx("m-item", s.item)}>
          <span className={cx("hw-rotulo", s.label)}>{copy.tax}</span>
          <span className={s.value}>{signed(vm.tax)}</span>
        </div>
      </div>

      {vm.effects.length > 0 && (
        <>
          <span aria-hidden="true" className={s.sep} />
          <div role="group" aria-label={copy.effectsGroup} className={cx(s.effects, s.effectsGroup)}>
            {vm.effects.map((e) => (
              <div key={e.role} aria-label={e.label} className={cx("m-item", s.item, s.effect)}>
                <span className={s.effectHead}>
                  <span className={cx("hw-rotulo", s.label)}>{e.name}</span>
                  <span className={s.value}>×{2 ** e.times}</span>
                </span>
                <span aria-hidden="true" className={s.pips}>
                  {Array.from({ length: e.rounds }, (_, i) => (
                    <span key={i} className={s.pip} />
                  ))}
                </span>
              </div>
            ))}
          </div>
        </>
      )}
    </section>
  );
}
