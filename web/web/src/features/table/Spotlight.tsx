import { useId } from "react";
import type { SeatBox } from "../../game/seats";
import { spotShape } from "../../game/seats";
import { cx } from "../../ui/cx";
import s from "./Spotlight.module.css";

/** Deslocamento da ponta de cima em cada quadro do contorno (só ela mexe). */
const TIP = [0, 3, -3];

/**
 * Holofote: cone do alto do palco até a base do camarote em cena, com a poça de luz na mesma base.
 * TRAVA: cone e poça são um desenho só, presos à base do camarote. Ver tech-docs/travas-de-design.md §2.
 * Fica por cima do camarote, translúcido e sem receber clique. Idle: os 3 quadros trocam juntos (.m-contorno).
 */
export function Spotlight({ box, width, top = 0 }: { box?: SeatBox; width: number; top?: number }) {
  const id = useId();
  if (!box) return null;
  return (
    <div aria-hidden="true" className={cx("m-cone", "m-contorno", s.spot)}>
      {TIP.map((dx, i) => {
        const { cone, pool } = spotShape(box, width / 2, top, dx);
        return (
          <svg key={dx} className={cx("m-f", s.frame)}>
            <defs>
              <linearGradient id={`${id}-${i}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0" className={s.fade0} />
                <stop offset="1" className={s.fade1} />
              </linearGradient>
            </defs>
            <polygon points={cone} fill={`url(#${id}-${i})`} />
            <ellipse {...pool} className={s.pool} />
          </svg>
        );
      })}
    </div>
  );
}
