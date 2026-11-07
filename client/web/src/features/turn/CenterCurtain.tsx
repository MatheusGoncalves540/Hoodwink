import { useCallback, useEffect, useId, useRef, type ReactNode } from "react";
import { copy } from "../../game/copy";
import { cx } from "../../ui/cx";
import s from "./CenterCurtain.module.css";

export type CurtainMove = "toggle" | "open" | "close";

/**
 * Máquina de estado da cortina (pura, testável em Node): aberta/fechada é um
 * booleano e cada gesto só a move num sentido.
 */
export function curtainState(open: boolean, move: CurtainMove): boolean {
  if (move === "open") return true;
  if (move === "close") return false;
  return !open;
}

type Props = {
  open: boolean;
  /** Sem `onToggle` a cortina não recolhe (durante o alvo, o painel fica preso aberto). */
  onToggle?(open: boolean): void;
  children: ReactNode;
};

/**
 * Painel central em cortina sobre a área dos jogadores. Sobe de baixo para cima,
 * deixa o handle focável quando fechada e prende o conteúdo com `inert` +
 * `visibility: hidden`. Com Menos movimento a troca é instantânea
 * ([data-motion="reduced"] em styles/motion.css).
 */
export function CenterCurtain({ open, onToggle, children }: Props) {
  const bodyId = useId();
  const bodyRef = useRef<HTMLDivElement>(null);
  const handleRef = useRef<HTMLButtonElement>(null);
  const collapsible = onToggle != null;

  // fechada, o conteúdo sai do foco e da árvore de acessibilidade
  useEffect(() => {
    const el = bodyRef.current;
    if (!el) return;
    if (open) el.removeAttribute("inert");
    else el.setAttribute("inert", "");
  }, [open]);

  // Esc recolhe a cortina e devolve o foco ao handle
  const move = useCallback(
    (next: CurtainMove) => {
      if (!onToggle) return;
      onToggle(curtainState(open, next));
    },
    [open, onToggle],
  );

  useEffect(() => {
    if (!open || !collapsible) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== "Escape") return;
      onToggle?.(false);
      handleRef.current?.focus();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, collapsible, onToggle]);

  return (
    <div className={cx(s.curtain, open && s.open)} data-open={open || undefined}>
      <div id={bodyId} ref={bodyRef} className={s.body} aria-hidden={open ? undefined : true}>
        {children}
      </div>
      {collapsible && (
        <button
          ref={handleRef}
          type="button"
          className={s.handle}
          aria-expanded={open}
          aria-controls={bodyId}
          onClick={() => move("toggle")}
        >
          <span aria-hidden="true" className={s.grip} />
          <span className={s.handleLabel}>{open ? copy.curtainClose : copy.curtainOpen}</span>
        </button>
      )}
    </div>
  );
}
