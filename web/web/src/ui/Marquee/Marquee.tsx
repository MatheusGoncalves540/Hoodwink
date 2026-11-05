import type { HTMLAttributes, ReactNode } from "react";
import { cx } from "../cx";
import s from "./Marquee.module.css";

/** Letreiro: moldura de veludo com lâmpadas (.hw-letreiro). */
export function Marquee({ children, className, ...rest }: HTMLAttributes<HTMLDivElement> & { children: ReactNode }) {
  return (
    <div {...rest} className={cx("hw-letreiro", s.marquee, className)}>
      {children}
    </div>
  );
}
