import { useEffect, useId, useRef, type ReactNode } from "react";
import { copy } from "../../game/copy";
import { Button } from "../../ui/Button/Button";
import { cx } from "../../ui/cx";
import { Moldura } from "../../ui/Moldura/Moldura";
import s from "./GlobalQuestion.module.css";

export type GlobalOption<T> = { value: T; title: string; hint: string; tag?: string; icon?: ReactNode };

type Props<T> = {
  /** Sobretítulo: de quem é a pergunta e em que etapa ("Clarividente · etapa 1 de 2"). */
  step: string;
  /** A pergunta; o "?" final ganha a cor da marca. */
  question: string;
  context?: string;
  options: GlobalOption<T>[];
  onPick(value: T): void;
  onBack(): void;
  /** Timer da jogada, no pé do cartaz. */
  timer?: ReactNode;
};

/**
 * Holofote de pergunta global: pergunta que não depende de ator nem de carta (ex.: Clarividente,
 * "Revelar para quem?"). A casa apaga, um refletor no teto do palco lança o feixe até a poça no chão
 * e o cartaz da pergunta pisa nela. Cores só de tokens (--hw-scrim/cone/poca/lampada), cartaz em
 * .hw-moldura e respostas em .hw-botao2: cada skin veste o holofote sem CSS próprio.
 * Pergunta sobre um ator ou uma carta continua nas poças de cada camarote (Table, targetLight).
 */
export function GlobalQuestion<T>({ step, question, context, options, onPick, onBack, timer }: Props<T>) {
  const titleId = useId();
  const firstRef = useRef<HTMLButtonElement>(null);
  const mark = question.endsWith("?");

  // diálogo: o foco vai para a primeira resposta e Esc volta às ações
  // shortcut: sem prender o Tab no cartaz; prender se a mesa por baixo virar problema de foco
  useEffect(() => {
    firstRef.current?.focus();
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onBack();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onBack]);

  return (
    <div className={s.stage} data-slot="global-question">
      <div aria-hidden="true" className={cx("m-scrim", s.scrim)} />
      <div aria-hidden="true" className={cx("m-cone", s.halo)} />
      <div aria-hidden="true" className={cx("m-cone", s.beam)} />
      <div aria-hidden="true" className={cx("m-cone", s.pool)} />
      <i aria-hidden="true" className={cx("m-cone", s.motes)} />
      <svg aria-hidden="true" className={s.lamp} viewBox="0 0 96 64">
        <path d="M10 4H86" className={s.truss} />
        <path d="M34 4v14M62 4v14" className={s.yoke} />
        <path d="M26 18h44l-7 26H33z" className={s.can} />
        <ellipse cx="48" cy="45" rx="16" ry="4.5" className={s.lens} />
      </svg>

      <Moldura lugar="pergunta" as="section" role="dialog" aria-modal="true" aria-labelledby={titleId} className={cx("m-rise", s.card)}>
        <div className={s.inner}>
          <p className={s.step}>{step}</p>
          <h2 id={titleId} className={s.question}>
            {mark ? question.slice(0, -1) : question}
            {mark && <span className={s.mark}>?</span>}
          </h2>
          {context && <p className={s.context}>{context}</p>}
          <div role="group" aria-labelledby={titleId} className={s.options}>
            {options.map((o, i) => (
              <button
                key={String(o.value)}
                ref={i === 0 ? firstRef : undefined}
                type="button"
                className={cx("hw-botao2", s.option)}
                onClick={() => onPick(o.value)}
              >
                {(o.icon || o.tag) && (
                  <span className={s.optionHead}>
                    {o.icon}
                    {o.tag && <span className={s.tag}>{o.tag}</span>}
                  </span>
                )}
                <span className={s.optionText}>
                  <span className={s.title}>{o.title}</span>
                  <span className={s.hint}>{o.hint}</span>
                </span>
              </button>
            ))}
          </div>
          <div className={s.foot}>
            <Button variant="link" className={s.back} onClick={onBack}>
              {copy.backToActions}
            </Button>
            {timer && <div className={s.timer}>{timer}</div>}
          </div>
        </div>
      </Moldura>
    </div>
  );
}
