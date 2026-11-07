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
  /** Fluxo de alvo: ids acesos recebem o anel e sobem acima do escuro; fora de cena ficam apagados. */
  targets?: string[];
};

/** Ferradura de camarotes; posições de game/seats.ts (tabela mobile e parábola desktop). */
export function SeatRing({ seats, layoutSeats, layout, onSeat, targets }: Props) {
  return (
    <section aria-label={copy.seatsRegion} className={s.seats}>
      {seats.map((seat, i) => {
        const box = layoutSeats.boxes[i];
        return box ? (
          <Seat
            key={seat.id}
            seat={seat}
            box={box}
            big={layoutSeats.big}
            wide={layoutSeats.wide}
            layout={layout}
            onOpen={onSeat}
            targetable={!!targets?.includes(seat.id)}
            dimmed={!!targets && seat.out}
          />
        ) : null;
      })}
    </section>
  );
}
