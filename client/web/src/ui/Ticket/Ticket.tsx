import { useId, type ButtonHTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";
import s from "./Ticket.module.css";

type Props = Omit<ButtonHTMLAttributes<HTMLButtonElement>, "disabled"> & {
  children: ReactNode;
  disabled?: boolean;
  size?: "m" | "l";
  /** Motivo anunciado quando o botão está desabilitado. */
  reason?: string;
};

/**
 * Botão principal (02-design-system.md → ticket). Desabilitado usa aria-disabled, não disabled:
 * continua focável e o leitor de tela ainda lê o rótulo.
 */
export function Ticket({ children, disabled, reason, size = "m", className, onClick, type = "button", ...rest }: Props) {
  const reasonId = useId();
  const describedBy = disabled && reason ? [rest["aria-describedby"], reasonId].filter(Boolean).join(" ") : rest["aria-describedby"];
  return (
    <button
      {...rest}
      type={type}
      className={cx("ticket", s.ticket, s[size], className)}
      aria-disabled={disabled || undefined}
      aria-describedby={describedBy}
      onClick={disabled ? (e) => e.preventDefault() : onClick}
    >
      {children}
      {disabled && reason && (
        <span id={reasonId} role="status" hidden>
          {reason}
        </span>
      )}
    </button>
  );
}
