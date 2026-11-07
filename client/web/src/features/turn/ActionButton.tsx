import { memo, useId, type CSSProperties } from "react";
import { copy } from "../../game/copy";
import type { Price } from "../../game/pricing";
import { actionDef, BASIC_GROUP, isBasic, ROLE_GROUPS, type ActionId } from "../../game/roles";
import { cx } from "../../ui/cx";
import { RoleEyes } from "./RoleEyes";
import s from "./ActionButton.module.css";

/** Tom da faixa de arte: tokens --hw-tom-<tom> e --hw-tom-<tom>-faixa (tokens.css). */
const toneOf = (id: ActionId) => (isBasic(id) ? BASIC_GROUP.tone : ROLE_GROUPS.find((g) => g.roles.includes(id))?.tone ?? "vinho");

type Props = {
  role: ActionId;
  price: Price;
  held: boolean;
  offTurn: boolean;
  /** Moedas que faltam; null quando dá para pagar. */
  missing: number | null;
  /** Motivo de estar desabilitado além de moeda (ex.: "Precisa de 6 moedas"). */
  blocked?: string;
  onClick(role: ActionId): void;
};

/** Botão da tela 6 (personagem ou ação do turno): faixa de arte, nome, "☞ na mão", descrição e preço. */
export const ActionButton = memo(function ActionButton({ role, price, held, offTurn, missing, blocked, onClick }: Props) {
  const descId = useId();
  const def = actionDef(role);
  const tone = toneOf(role);
  const poor = missing != null;
  const disabled = offTurn || poor || !!blocked;
  const priceText = price.kind === "gain" ? `+${price.coins}` : price.kind === "reaction" ? `${price.coins}+` : String(price.coins);
  const label = offTurn
    ? copy.actionOffTurn(def.name)
    : blocked
      ? `${def.name}, ${blocked}`
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
      aria-describedby={disabled ? descId : undefined}
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
        <span id={descId} aria-live="polite" className={s.desc}>
          {blocked ?? (poor ? copy.missing(missing) : copy.roleDesc(role, price))}
        </span>
      </span>
      <span className={cx(s.price, price.kind === "gain" && s.gain, disabled && s.priceOff)}>{priceText}</span>
    </button>
  );
});
