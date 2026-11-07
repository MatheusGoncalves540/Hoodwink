import { useId, type InputHTMLAttributes, type Ref } from "react";
import { cx } from "../cx";
import s from "./TextField.module.css";

type Props = Omit<InputHTMLAttributes<HTMLInputElement>, "onChange" | "onSubmit"> & {
  label: string;
  value: string;
  onChange(value: string): void;
  /** Texto do botão colado ("Entrar", "Falar", "Enviar"). Enter no campo também dispara. */
  action?: string;
  /** Desabilita só o botão colado, sem tirar o campo do fluxo (Lobby: menos de 6 letras). */
  actionDisabled?: boolean;
  /** "dark" no trilho do desktop: o "Enviar" vira pílula de tinta. */
  tone?: "light" | "dark";
  inputRef?: Ref<HTMLInputElement>;
  onSubmit?(): void;
  upper?: boolean;
};

/** Campo em pílula com rótulo só para leitor de tela (Lobby.dc.html). */
export function TextField({
  label,
  value,
  onChange,
  action,
  actionDisabled,
  tone = "light",
  inputRef,
  onSubmit,
  upper,
  className,
  ...rest
}: Props) {
  const id = useId();
  return (
    <form
      className={cx(s.field, className)}
      onSubmit={(e) => {
        e.preventDefault();
        if (!actionDisabled) onSubmit?.();
      }}
    >
      <label htmlFor={id} className={s.hidden}>
        {label}
      </label>
      <input
        autoComplete="off"
        {...rest}
        ref={inputRef}
        id={id}
        className={cx(s.input, upper && s.upper)}
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
      {action && (
        <button
          type="submit"
          className={cx(s.action, tone === "dark" && s.actionDark)}
          aria-disabled={actionDisabled || undefined}
          onClick={actionDisabled ? (e) => e.preventDefault() : undefined}
        >
          {action}
        </button>
      )}
    </form>
  );
}
