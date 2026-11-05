/**
 * Elenco da sala (Sala.dc.html → seção "Elenco").
 * De 7 a 10 lugares as linhas encolhem (34 px, nome 19 px) para caber sem rolar.
 */
import type { Layout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import { CLASSIC_RULES } from "../../game/rules.classic";
import type { selectWaitingRoom } from "../../game/selectors";
import { cx } from "../../ui/cx";
import s from "./Cast.module.css";

export type WaitingRoomVM = ReturnType<typeof selectWaitingRoom>;

type Props = { vm: WaitingRoomVM; layout: Layout; className?: string };

export function Cast({ vm, layout, className }: Props) {
  return (
    <section
      aria-label={copy.castRegion}
      className={cx("m-hang", s.cast, layout === "desktop" && s.desk, className)}
    >
      <h1 className={s.title}>{vm.name}</h1>
      <p className={s.castLine}>
        {copy.castLineLabel}
        {vm.castLine}
      </p>
      <ul className={cx(s.list, vm.tight && s.tight)}>
        {vm.seats.map((seat) => (
          <li key={seat.key} className={cx("m-item", s.row, seat.vacant && s.vacant)}>
            <span className={s.name}>{seat.name}</span>
            <span aria-hidden="true" className={s.leader} />
            <span className={s.tag}>{seat.tag}</span>
          </li>
        ))}
      </ul>
      <p className={s.rule}>
        {copy.houseRuleInitial(CLASSIC_RULES.general.initialCoins)}
        <br />
        {copy.houseRuleLimit(CLASSIC_RULES.general.maxCoins)}
      </p>
    </section>
  );
}
