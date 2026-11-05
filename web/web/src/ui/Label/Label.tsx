import type { HTMLAttributes, ReactNode } from "react";
import { cx } from "../cx";
import s from "./Label.module.css";

type Props = HTMLAttributes<HTMLSpanElement> & { children: ReactNode; tone?: "ouro" | "suave" };

/** Rótulo em caixa-alta (Josefin Sans): "RODADA 4", "EM CARTAZ", "PARA GANHAR MOEDAS". */
export function Label({ children, tone, className, ...rest }: Props) {
  return (
    <span {...rest} className={cx("hw-rotulo", s.label, tone && s[tone], className)}>
      {children}
    </span>
  );
}
