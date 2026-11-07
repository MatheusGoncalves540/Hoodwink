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
  /** Linha do que já foi escolhido (Clarividente: revelar para todos ou só para você). */
  note?: string;
};

/** Fluxo de alvo inferido (05-protocolo.md): escolhida a ação, a mesa espera o lugar ou a carta. */
export function TargetBar({ role, onBack, onSwapBoth, note }: Props) {
  return (
    <section aria-label={copy.pickTarget} className={s.bar}>
      <p className={cx("m-respira-forte", s.title)}>{copy.pickTarget}</p>
      <p className={s.hint}>
        {role === "KILL_CARD" ? copy.killHint : role === "SWAP_CARDS" ? copy.swapHint : copy.pickTargetHint(roleName(role))}
      </p>
      {note && <p className={s.note}>{note}</p>}
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
