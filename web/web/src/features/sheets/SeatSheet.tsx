import { copy } from "../../game/copy";
import type { GameView, SeatSheetVM } from "../../game/selectors";
import { selectSeatSheet } from "../../game/selectors";
import { Coin } from "../../ui/Coin/Coin";
import { Label } from "../../ui/Label/Label";
import { Sheet } from "../../ui/Sheet/Sheet";
import { SheetCard } from "./SheetCard";
import s from "./SeatSheet.module.css";

type Props = {
  view: GameView;
  seatId: string;
  mode: "mobile" | "desktop";
  onClose(): void;
  /** Fluxo de alvo: cada carta viva vira botão. */
  onPickCard?(index: number): void;
  backLabel?: string;
};

/** Camarote aberto (telas 4 e 14). */
export function SeatSheet({ view, seatId, mode, onClose, onPickCard, backLabel }: Props) {
  const vm = selectSeatSheet(view, seatId);
  if (!vm) return null;
  return (
    <Sheet open onClose={onClose} label={copy.seatSheet(vm.name)} mode={mode} backLabel={backLabel}>
      <SeatSheetBody vm={vm} onPickCard={onPickCard} />
    </Sheet>
  );
}

export function SeatSheetBody({ vm, onPickCard }: { vm: SeatSheetVM; onPickCard?(index: number): void }) {
  return (
    <>
      <div className={s.header}>
        <h2 className={s.name}>{vm.name}</h2>
        <Coin size={56} value={vm.coins} label={copy.coins(vm.coins)} />
      </div>

      <div className={s.cards}>
        {vm.cards.map((c, i) => {
          if (!onPickCard) return <SheetCard key={i} state={c.state} role={c.role} />;
          const disabled = c.state === "dead";
          const state = copy.cardState[c.state === "back" ? "hidden" : c.state];
          return (
            <button
              key={i}
              type="button"
              className={s.pick}
              aria-label={copy.pickCard(vm.name, state)}
              aria-disabled={disabled || undefined}
              onClick={disabled ? undefined : () => onPickCard(c.index)}
            >
              <SheetCard state={c.state} role={c.role} />
            </button>
          );
        })}
      </div>

      <Label>{copy.tableSaw}</Label>
      {vm.log.length > 0 ? (
        <ol className={s.log}>
          {vm.log.map((l) => (
            <li key={l.key} className={s.logRow}>
              <span className={s.round}>{l.turn}</span>
              <span className={s.text}>{l.text}</span>
            </li>
          ))}
        </ol>
      ) : (
        <p className={s.empty}>{copy.nothingSeen}</p>
      )}

      {vm.investments.length > 0 && (
        <div role="group" aria-label={copy.investmentsGroup} className={s.investments}>
          <Label>{copy.investmentsGroup}</Label>
          <div className={s.boxes}>
            {vm.investments.map((rounds, i) => (
              <div key={i} className={s.box} aria-label={copy.endsInLabel(rounds)}>
                <span className={s.endsIn}>{copy.endsIn}</span>
                <span aria-hidden="true" className={s.pips}>
                  {Array.from({ length: rounds }, (_, j) => (
                    <span key={j} className={s.pip} />
                  ))}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </>
  );
}
