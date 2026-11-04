import { cx } from "../cx";
import s from "./Coin.module.css";

export type CoinSize = 22 | 26 | 28 | 34 | 46 | 50 | 56 | 60;

type Props = { value?: number; size: CoinSize; label?: string; className?: string };

/** Moeda de ouro; sem valor vira o losango do papel sem custo (Kamikaze). */
export function Coin({ value, size, label, className }: Props) {
  const a11y = label ? { role: "img", "aria-label": label } : { "aria-hidden": true };
  if (value === undefined) return <span className={cx(s.diamond, className)} {...a11y} />;
  return (
    <span className={cx("hw-num", s.coin, s[`s${size}`], className)} {...a11y}>
      {value}
    </span>
  );
}
