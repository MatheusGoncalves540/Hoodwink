import { copy } from "../../game/copy";
import { roleName } from "../../game/roles";
import type { GameView, YouSheetVM } from "../../game/selectors";
import { selectYouSheet } from "../../game/selectors";
import { Coin } from "../../ui/Coin/Coin";
import { Label } from "../../ui/Label/Label";
import { Sheet } from "../../ui/Sheet/Sheet";
import { SheetCard } from "./SheetCard";
import s from "./YouSheet.module.css";

type Props = {
  view: GameView;
  mode: "mobile" | "desktop";
  onClose(): void;
  /** Fluxo de alvo: suas cartas viram botões (Guardiã, Coveiro, Crupiê). */
  onPickCard?(index: number): void;
  cardDisabled?(index: number): boolean;
  backLabel?: string;
};

/** Seu camarote (telas 12 e 16). */
export function YouSheet({ view, mode, onClose, onPickCard, cardDisabled, backLabel }: Props) {
  const vm = selectYouSheet(view);
  if (!vm) return null;
  return (
    <Sheet open onClose={onClose} label={copy.yourBox} mode={mode} backLabel={backLabel}>
      <YouSheetBody vm={vm} onPickCard={onPickCard} cardDisabled={cardDisabled} />
    </Sheet>
  );
}

export function YouSheetBody({
  vm,
  onPickCard,
  cardDisabled,
}: {
  vm: YouSheetVM;
  onPickCard?(index: number): void;
  cardDisabled?(index: number): boolean;
}) {
  const stateOf = (c: YouSheetVM["cards"][number]) => (c.dead ? "dead" : c.role ? "revealed" : "back") as const;

  return (
    <>
      <div className={s.header}>
        <h2 className={s.name}>{vm.name}</h2>
        <Coin size={56} value={vm.coins} label={copy.coins(vm.coins)} />
      </div>
      <p className={s.rule}>{copy.overLimit(vm.limit)}</p>

      <div className={s.cards}>
        {vm.cards.map((c, i) => {
          const state = stateOf(c);
          const card = (
            <SheetCard
              state={state}
              role={c.role}
              protected={c.protected}
              top={c.dead ? undefined : copy.tableSees(c.protected ? "protected" : "back")}
              bottom={c.dead ? undefined : copy.cardState[c.protected ? "protected" : "hidden"]}
            />
          );
          if (!onPickCard) return <SheetCard key={i} state={state} role={c.role} protected={c.protected} top={c.dead ? undefined : copy.tableSees(c.protected ? "protected" : "back")} bottom={c.dead ? undefined : copy.cardState[c.protected ? "protected" : "hidden"]} />;
          const disabled = cardDisabled ? cardDisabled(c.index) : c.dead;
          const label = copy.pickOwnCard(c.role ? roleName(c.role) : copy.cardState.hidden);
          return (
            <button
              key={i}
              type="button"
              className={s.pick}
              aria-label={label}
              aria-disabled={disabled || undefined}
              onClick={disabled ? undefined : () => onPickCard(c.index)}
            >
              {card}
            </button>
          );
        })}
      </div>

      {vm.investments.length > 0 && (
        <div role="group" aria-label={copy.investmentsGroup} className={s.investments}>
          <p className={s.investLine}>{copy.investmentsLine(vm.investments.length)}</p>
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

      {vm.roles.length > 0 && (
        <div className={s.roles}>
          <Label>{copy.rolesSeen}</Label>
          <ul className={s.roleList}>
            {vm.roles.map((r) => (
              <li key={r.key} className={s.roleRow}>
                <div className={s.roleHead}>
                  <span className={s.roleName}>{r.name}</span>
                  <span className={s.roleRound}>{copy.roundN(r.turn)}</span>
                </div>
                <span className={s.roleText}>{r.text}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </>
  );
}
