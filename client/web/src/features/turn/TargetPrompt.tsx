import { useEffect, useId, useRef, type ReactNode } from "react";
import { copy } from "../../game/copy";
import type { SeatSheetVM } from "../../game/selectors";
import { SheetCard } from "../sheets/SheetCard";
import { QuestionCard } from "./GlobalQuestion";
import s from "./TargetPrompt.module.css";

type Props = {
  step: string;
  question: string;
  context?: string;
  backLabel?: string;
  onBack(): void;
  timer?: ReactNode;
  /** Etapa "qual carta": o camarote escolhido; sem ele, a etapa é "quem" (toque nos camarotes acesos). */
  seat?: SeatSheetVM | null;
  onPickCard?(index: number): void;
};

/**
 * Cartaz da escolha de alvo, no pé da mesa apagada (TargetSpotlight): o mesmo cartaz da pergunta global, com a
 * etapa ("Assassino · etapa 1 de 2"). Na etapa da carta, as cartas vivas do alvo viram os botões do cartaz.
 */
export function TargetPrompt({ step, question, context, backLabel, onBack, timer, seat, onPickCard }: Props) {
  const titleId = useId();
  const firstRef = useRef<HTMLButtonElement>(null);
  const cards = seat && onPickCard ? seat.cards.filter((c) => c.state !== "dead") : null;

  useEffect(() => {
    firstRef.current?.focus();
  }, [seat?.id]);

  return (
    <QuestionCard step={step} question={question} context={context} titleId={titleId} onBack={onBack} backLabel={backLabel} timer={timer} className={s.card}>
      {cards && seat && onPickCard && (
        <div role="group" aria-labelledby={titleId} className={s.cards}>
          {cards.map((c, i) => (
            <button
              key={c.index}
              ref={i === 0 ? firstRef : undefined}
              type="button"
              className={s.pick}
              aria-label={copy.pickCard(seat.name, copy.cardState[c.state === "back" ? "hidden" : c.state])}
              onClick={() => onPickCard(c.index)}
            >
              <SheetCard state={c.state} role={c.role} />
            </button>
          ))}
        </div>
      )}
    </QuestionCard>
  );
}
