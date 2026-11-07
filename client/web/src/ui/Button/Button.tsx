import { useId, type ButtonHTMLAttributes, type ReactNode } from "react";
import { cx } from "../cx";
import s from "./Button.module.css";

type Props = Omit<ButtonHTMLAttributes<HTMLButtonElement>, "disabled"> & {
  children: ReactNode;
  variant: "outline" | "danger" | "link";
  disabled?: boolean;
  /** Motivo anunciado quando o botão está desabilitado. */
  reason?: string;
};

/** Secundários: "Deixar passar" (outline), "Contestar" (danger), "☜ Voltar" e "Sair" (link). Medidas: Main, Camarote, ForaDeCena. */
export function Button({ children, variant, disabled, reason, className, onClick, type = "button", ...rest }: Props) {
  const reasonId = useId();
  const describedBy = disabled && reason ? [rest["aria-describedby"], reasonId].filter(Boolean).join(" ") : rest["aria-describedby"];
  return (
    <button
      {...rest}
      type={type}
      className={cx(s.button, s[variant], variant === "outline" && "hw-botao2", className)}
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
