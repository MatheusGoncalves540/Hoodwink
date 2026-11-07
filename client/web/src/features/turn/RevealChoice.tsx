import type { ReactNode } from "react";
import { copy } from "../../game/copy";
import { GlobalQuestion, type GlobalOption } from "./GlobalQuestion";

type Props = {
  onPick(toAll: boolean): void;
  onBack(): void;
  timer?: ReactNode;
};

const audience = (
  <svg width="48" height="34" viewBox="0 0 48 34" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" aria-hidden="true">
    <circle cx="11" cy="13" r="4.5" />
    <circle cx="24" cy="10" r="5" />
    <circle cx="37" cy="13" r="4.5" />
    <path d="M3 32c0-6 3.5-10 8-10s8 4 8 10M15 32c0-7 4-11 9-11s9 4 9 11M29 32c0-6 3.5-10 8-10s8 4 8 10" />
  </svg>
);

const wings = (
  <svg width="48" height="34" viewBox="0 0 48 34" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" aria-hidden="true">
    <path d="M24 2 L37 32 H11 Z" fill="currentColor" fillOpacity="0.16" stroke="none" />
    <path d="M19 2h10" />
    <circle cx="24" cy="15" r="5" />
    <path d="M15 32c0-7 4-11 9-11s9 4 9 11" />
  </svg>
);

const OPTIONS: GlobalOption<boolean>[] = [
  { value: true, title: copy.revealEveryone, hint: copy.revealEveryoneHint, tag: copy.revealEveryoneTag, icon: audience },
  { value: false, title: copy.revealOnlyMe, hint: copy.revealOnlyMeHint, tag: copy.revealOnlyMeTag, icon: wings },
];

/**
 * Primeira etapa do Clarividente, no holofote de pergunta global: a carta aparece
 * para todos ou só para você. Só depois a mesa mostra quem escolher.
 */
export function RevealChoice({ onPick, onBack, timer }: Props) {
  return (
    <GlobalQuestion
      step={copy.revealStep}
      question={copy.revealQuestion}
      context={copy.revealContext}
      options={OPTIONS}
      onPick={onPick}
      onBack={onBack}
      timer={timer}
    />
  );
}
