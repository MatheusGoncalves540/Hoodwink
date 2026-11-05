import type { CSSProperties } from "react";
import { copy } from "../../game/copy";
import type { Price } from "../../game/pricing";
import type { RoleId } from "../../game/protocol";
import { ROLES, ROLE_GROUPS } from "../../game/roles";
import { cx } from "../../ui/cx";
import { RoleEyes } from "./RoleEyes";
import s from "./ActionButton.module.css";

/** Tom da faixa de arte: tokens --hw-tom-<tom> e --hw-tom-<tom>-faixa (tokens.css). */
const toneOf = (role: RoleId) => ROLE_GROUPS.find((g) => g.roles.includes(role))?.tone ?? "vinho";

type Props = {
  role: RoleId;
  price: Price;
  held: boolean;
  offTurn: boolean;
  /** Moedas que faltam; null quando dá para pagar. */
  missing: number | null;
  onClick(role: RoleId): void;
};

/** Botão de papel da tela 6: faixa de arte, nome, "☞ na mão", descrição e preço. */
export function ActionButton({ role, price, held, offTurn, missing, onClick }: Props) {
  const def = ROLES[role];
  const tone = toneOf(role);
  const poor = missing != null;
  const disabled = offTurn || poor;
  const priceText = price.kind === "gain" ? `+${price.coins}` : price.kind === "reaction" ? `${price.coins}+` : String(price.coins);
  const label = offTurn
    ? copy.actionOffTurn(def.name)
    : poor
      ? copy.actionPoor(def.name, price.coins, missing)
      : price.kind === "gain"
        ? copy.actionGain(def.name, price.coins)
        : copy.actionCost(def.name, price.coins);

  return (
    <button
      type="button"
      className={cx("m-item", s.button, disabled && s.off)}
      style={{ "--tone": `var(--hw-tom-${tone})`, "--band": `var(--hw-tom-${tone}-faixa)`, "--m-base": ".25s" } as CSSProperties}
      aria-label={label}
      aria-disabled={disabled || undefined}
      onClick={disabled ? undefined : () => onClick(role)}
    >
      <span className={cx(s.art, disabled && s.artOff)} aria-hidden="true">
        <RoleEyes look={def.look} />
        <span className={s.fade} />
      </span>
      <span className={cx(s.text, disabled && s.textOff)}>
        <span className={s.nameRow}>
          <span className={s.name}>{def.name}</span>
          {held && <span className={s.held}>{copy.inHand}</span>}
        </span>
        <span className={s.desc}>{poor ? copy.missing(missing) : copy.roleDesc(role, price)}</span>
      </span>
      <span className={cx(s.price, price.kind === "gain" && s.gain, disabled && s.priceOff)}>{priceText}</span>
    </button>
  );
}
