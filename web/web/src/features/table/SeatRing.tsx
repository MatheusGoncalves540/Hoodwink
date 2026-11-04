import type { Layout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import type { SeatVM } from "../../game/selectors";
import type { SeatLayout } from "../../game/seats";
import { Seat } from "./Seat";
import s from "./SeatRing.module.css";

type Props = {
  seats: SeatVM[];
  layoutSeats: SeatLayout;
  layout: Layout;
  onSeat?(id: string): void;
};

/** Ferradura de camarotes; posições de game/seats.ts (tabela mobile e parábola desktop). */
export function SeatRing({ seats, layoutSeats, layout, onSeat }: Props) {
  return (
    <section aria-label={copy.seatsRegion} className={s.seats}>
      {seats.map((seat, i) => {
        const box = layoutSeats.boxes[i];
        return box ? <Seat key={seat.id} seat={seat} box={box} big={layoutSeats.big} layout={layout} onOpen={onSeat} /> : null;
      })}
    </section>
  );
}
