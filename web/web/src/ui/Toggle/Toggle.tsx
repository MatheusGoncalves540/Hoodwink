import { copy } from "../../game/copy";
import { cx } from "../cx";
import s from "./Toggle.module.css";

type Props = { label: string; hint?: string; value: boolean; disabled?: boolean; onChange(value: boolean): void };

/**
 * Par "sim / não" do menu. role="switch" como na prancha (Menu.dc.html), no lugar do radiogroup
 * do prompt 02: um botão só, um Tab, e o leitor anuncia "ligado/desligado".
 */
export function Toggle({ label, hint, value, disabled, onChange }: Props) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={value}
      aria-disabled={disabled || undefined}
      className={cx(s.toggle, disabled && s.disabled)}
      onClick={disabled ? undefined : () => onChange(!value)}
    >
      <span className={s.text}>
        <span className={s.label}>{label}</span>
        {hint && <span className={s.hint}>{hint}</span>}
      </span>
      <span aria-hidden="true" className={s.leader} />
      <span aria-hidden="true" className={s.options}>
        <span className={cx(s.option, value && s.on)}>{copy.yes}</span>
        <span className={cx(s.option, !value && s.on)}>{copy.no}</span>
      </span>
    </button>
  );
}
