import { copy } from "../../game/copy";
import { roleName, type ActionId } from "../../game/roles";
import { Button } from "../../ui/Button/Button";
import { cx } from "../../ui/cx";
import s from "./TargetBar.module.css";

type Props = {
  role: ActionId;
  onBack(): void;
  /** Trocar cartas: troca as duas de uma vez (custa todas as moedas). */
  onSwapBoth?: () => void;
  /** Clarividente: mostrar a carta para todos ou só para você. Ausente nos outros. */
  showToAll?: boolean;
  onShowToAll?: (v: boolean) => void;
};

/** Fluxo de alvo inferido (05-protocolo.md): escolhida a ação, a mesa espera o lugar ou a carta. */
export function TargetBar({ role, onBack, onSwapBoth, showToAll, onShowToAll }: Props) {
  return (
    <section aria-label={copy.pickTarget} className={s.bar}>
      <p className={s.title}>{copy.pickTarget}</p>
      <p className={s.hint}>{role === "KILL_CARD" ? copy.killHint : role === "SWAP_CARDS" ? copy.swapHint : copy.pickTargetHint(roleName(role))}</p>
      {showToAll != null && onShowToAll && (
        <div className={s.choice} role="group" aria-label={copy.showToAll}>
          <span className={s.choiceLabel}>{copy.showToAll}</span>
          <div className={s.pair}>
            <button
              type="button"
              className={cx(s.option, showToAll && s.on)}
              aria-pressed={showToAll}
              onClick={() => onShowToAll(true)}
            >
              {copy.showToAllYes}
            </button>
            <button
              type="button"
              className={cx(s.option, !showToAll && s.on)}
              aria-pressed={!showToAll}
              onClick={() => onShowToAll(false)}
            >
              {copy.showToAllNo}
            </button>
          </div>
          <p className={s.hint}>{showToAll ? copy.showToAllHint : copy.showToMe}</p>
        </div>
      )}
      {onSwapBoth && (
        <>
          <Button variant="outline" className={s.back} onClick={onSwapBoth}>
            {copy.swapBoth}
          </Button>
          <p className={s.hint}>{copy.swapBothWarn}</p>
        </>
      )}
      <Button variant="link" className={s.back} onClick={onBack}>
        {copy.backToActions}
      </Button>
    </section>
  );
}
