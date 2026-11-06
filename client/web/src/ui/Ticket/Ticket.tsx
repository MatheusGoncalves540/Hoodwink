import type { ButtonHTMLAttributes, ReactNode } from "react";
import { cx } from "../cx";
import s from "./Ticket.module.css";

type Props = Omit<ButtonHTMLAttributes<HTMLButtonElement>, "disabled"> & {
  children: ReactNode;
  disabled?: boolean;
  size?: "m" | "l";
};

/**
 * Botão principal (02-design-system.md → ticket). Desabilitado usa aria-disabled, não disabled:
 * continua focável e o leitor de tela ainda lê o rótulo.
 */
export function Ticket({ children, disabled, size = "m", className, onClick, type = "button", ...rest }: Props) {
  return (
    <button
      {...rest}
      type={type}
      className={cx("ticket", s.ticket, s[size], className)}
      aria-disabled={disabled || undefined}
      onClick={disabled ? (e) => e.preventDefault() : onClick}
    >
      {children}
    </button>
  );
}
