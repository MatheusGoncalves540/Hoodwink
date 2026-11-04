import type { Layout } from "../../app/useLayout";
import type { SeatVM } from "../../game/selectors";
import type { SeatBox } from "../../game/seats";
import { CardMini } from "../../ui/Card/Card";
import { Coin } from "../../ui/Coin/Coin";
import { cx } from "../../ui/cx";
import s from "./Seat.module.css";

type Props = {
  seat: SeatVM;
  box: SeatBox;
  big: boolean;
  layout: Layout;
  onOpen?(id: string): void;
};

/** Camarote: arco, primeiro nome, 2 miniaturas + moeda e "···" quando há algo para ver. */
export function Seat({ seat, box, big, layout, onOpen }: Props) {
  const mini = layout === "desktop" ? (big ? "desk-l" : "desk-s") : big ? "l" : "s";
  return (
    <button
      type="button"
      className={cx("m-item", s.seat, !big && s.small, seat.inScene && s.active, seat.out && s.out, layout === "desktop" && s.desk)}
      style={{ left: box.x, top: box.y, width: box.w, height: box.h }}
      aria-label={seat.label}
      onClick={() => onOpen?.(seat.id)}
    >
      <span className={s.name}>{seat.short}</span>
      {!seat.out && (
        <span className={s.body}>
          <span className={s.cards}>
            {seat.cards.slice(0, 2).map((c, i) => (
              <CardMini key={i} state={c.state} role={c.role} size={mini} />
            ))}
          </span>
          <Coin size={big ? 26 : 22} value={seat.coins} />
        </span>
      )}
      <span aria-hidden="true" className={cx(s.dots, !seat.hasMore && s.noDots)}>
        ···
      </span>
    </button>
  );
}
