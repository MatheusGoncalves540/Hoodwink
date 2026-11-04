import { useId, type InputHTMLAttributes } from "react";
import { cx } from "../cx";
import s from "./TextField.module.css";

type Props = Omit<InputHTMLAttributes<HTMLInputElement>, "onChange" | "onSubmit"> & {
  label: string;
  value: string;
  onChange(value: string): void;
  /** Texto do botão colado ("Entrar", "Falar", "Enviar"). Enter no campo também dispara. */
  action?: string;
  /** Desabilita só o botão colado, sem tirar o campo do fluxo (Lobby: menos de 4 letras). */
  actionDisabled?: boolean;
  onSubmit?(): void;
  upper?: boolean;
};

/** Campo em pílula com rótulo só para leitor de tela (Lobby.dc.html). */
export function TextField({ label, value, onChange, action, actionDisabled, onSubmit, upper, className, ...rest }: Props) {
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
        id={id}
        className={cx(s.input, upper && s.upper)}
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
      {action && (
        <button
          type="submit"
          className={s.action}
          aria-disabled={actionDisabled || undefined}
          onClick={actionDisabled ? (e) => e.preventDefault() : undefined}
        >
          {action}
        </button>
      )}
    </form>
  );
}
