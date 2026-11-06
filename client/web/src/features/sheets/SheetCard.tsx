import type { ReactNode } from "react";
import { copy } from "../../game/copy";
import type { RoleId } from "../../game/protocol";
import { roleName } from "../../game/roles";
import { CardFace, type MiniState } from "../../ui/Card/Card";
import { cx } from "../../ui/cx";
import s from "./SheetCard.module.css";

type Props = {
  state: MiniState;
  role?: RoleId;
  protected?: boolean;
  top?: ReactNode;
  bottom?: ReactNode;
};

/**
 * Carta de gaveta (96 px). Escondida e morta seguem a prancha da gaveta: rótulo no alto,
 * nome/"?" no pé, alinhados à esquerda — diferente da miniatura centrada do CardHidden.
 */
export function SheetCard({ state, role, protected: prot, top, bottom }: Props) {
  if (state === "dead") {
    return (
      <div className={s.deadCard}>
        <span className={s.state}>{copy.cardState.dead}</span>
        {role && <span className={s.deadName}>{roleName(role)}</span>}
      </div>
    );
  }
  if (state === "revealed" && role) {
    return <CardFace size="m" role={role} protected={prot} top={top} bottom={bottom ?? copy.cardState.revealed} className={s.revealed} />;
  }
  if (state === "protected") {
    return (
      <div className={cx(s.hiddenCard, s.protected)}>
        <span className={s.hiddenLabel}>{copy.cardState.protected}</span>
        <span className={s.ask}>?</span>
      </div>
    );
  }
  return (
    <div className={s.hiddenCard}>
      <span className={s.hiddenLabel}>{copy.cardState.hidden}</span>
      <span className={s.ask}>?</span>
    </div>
  );
}
