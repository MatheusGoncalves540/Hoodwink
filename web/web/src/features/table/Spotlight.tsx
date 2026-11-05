import type { SeatBox } from "../../game/seats";
import { conePolygon } from "../../game/seats";
import { cx } from "../../ui/cx";
import s from "./Spotlight.module.css";

/** Cone de luz: do alto do palco até a base do camarote em cena, com 6 px de folga.
 *  TRAVA: cone e poça são um desenho só, presos à base do camarote. Ver tech-docs/travas-de-design.md §2. */
export function Spotlight({ box, width, top = 0 }: { box?: SeatBox; width: number; top?: number }) {
  if (!box) return null;
  return <div aria-hidden="true" className={cx("m-cone", s.cone)} style={{ clipPath: conePolygon(box, width / 2, top) }} />;
}
