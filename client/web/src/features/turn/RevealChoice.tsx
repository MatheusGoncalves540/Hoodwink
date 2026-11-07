import { copy } from "../../game/copy";
import { Button } from "../../ui/Button/Button";
import { cx } from "../../ui/cx";
import s from "./RevealChoice.module.css";

type Props = {
  onPick(toAll: boolean): void;
  onBack(): void;
};

/**
 * Primeira etapa do Clarividente: dentro do holofote, escolher se a carta aparece
 * para todos ou só para você. Só depois a mesa mostra quem escolher.
 */
export function RevealChoice({ onPick, onBack }: Props) {
  return (
    <section aria-label={copy.revealFor} className={s.panel}>
      <p className={cx("m-respira-forte", s.label)}>{copy.revealFor}</p>
      <div className={s.options}>
        <button type="button" className={cx("m-respira-forte", s.option)} onClick={() => onPick(true)}>
          <span className={s.title}>{copy.revealEveryone}</span>
          <span className={s.hint}>{copy.revealEveryoneHint}</span>
        </button>
        <button type="button" className={cx("m-respira-forte", s.option)} onClick={() => onPick(false)}>
          <span className={s.title}>{copy.revealOnlyMe}</span>
          <span className={s.hint}>{copy.revealOnlyMeHint}</span>
        </button>
      </div>
      <Button variant="link" className={s.back} onClick={onBack}>
        {copy.backToActions}
      </Button>
    </section>
  );
}
