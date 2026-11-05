import type { ButtonHTMLAttributes, ReactNode } from "react";
import { cx } from "../cx";
import s from "./Button.module.css";

type Props = Omit<ButtonHTMLAttributes<HTMLButtonElement>, "disabled"> & {
  children: ReactNode;
  variant: "outline" | "danger" | "link";
  disabled?: boolean;
};

/** Secundários: "Deixar passar" (outline), "Contestar" (danger), "☜ Voltar" e "Sair" (link). Medidas: Main, Camarote, ForaDeCena. */
export function Button({ children, variant, disabled, className, onClick, type = "button", ...rest }: Props) {
  return (
    <button
      {...rest}
      type={type}
      className={cx(s.button, s[variant], variant === "outline" && "hw-botao2", className)}
      aria-disabled={disabled || undefined}
      onClick={disabled ? (e) => e.preventDefault() : onClick}
    >
      {children}
    </button>
  );
}
