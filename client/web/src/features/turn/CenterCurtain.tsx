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

/** Arraste vertical no handle: para cima abre, para baixo fecha; menos de 30 px é toque. */
export function swipeMove(dy: number): CurtainMove | null {
  if (dy <= -30) return "open";
  if (dy >= 30) return "close";
  return null;
}

type Props = {
  open: boolean;
  /** Sem `onToggle` a cortina não recolhe (durante o alvo, o painel fica preso aberto). */
  onToggle?(open: boolean): void;
  /** Conteúdo do handle (título, moedas, tempo): fica à vista com a gaveta fechada. */
  peek?: ReactNode;
  children: ReactNode;
};

/**
 * Gaveta da mesa, puxada de baixo para cima: o handle é a borda de cima e sobe junto com o painel
 * (fechada, só ele fica à vista no pé, acima da faixa Você). Toque ou arraste abre e fecha;
 * aberta, para em --curtain-top e deixa o palco de cima à vista. O conteúdo fechado sai do foco com
 * `inert` + `visibility`. Com Menos movimento a troca é instantânea
 * ([data-motion="reduced"] em styles/motion.css).
 */
export function CenterCurtain({ open, onToggle, peek, children }: Props) {
  const bodyId = useId();
  const bodyRef = useRef<HTMLDivElement>(null);
  const handleRef = useRef<HTMLButtonElement>(null);
  const collapsible = onToggle != null;
  const startY = useRef<number | null>(null);
  const swiped = useRef(false);

  // fechada, o conteúdo sai do foco e da árvore de acessibilidade
  useEffect(() => {
    const el = bodyRef.current;
    if (!el) return;
    if (open) el.removeAttribute("inert");
    else el.setAttribute("inert", "");
  }, [open]);

  const move = useCallback(
    (next: CurtainMove) => {
      if (!onToggle) return;
      onToggle(curtainState(open, next));
    },
    [open, onToggle],
  );

  // Esc recolhe a gaveta e devolve o foco ao handle
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
    <div className={cx(s.curtain, open && s.open, collapsible && s.hasHandle, peek != null && s.hasPeek)} data-open={open || undefined}>
      <div className={s.drawer}>
        {collapsible && (
          <button
            ref={handleRef}
            type="button"
            className={s.handle}
            aria-expanded={open}
            aria-controls={bodyId}
            aria-label={peek != null ? (open ? copy.curtainClose : copy.curtainOpen) : undefined}
            onPointerDown={(e) => {
              // o soltar cai fora do handle quando o dedo desce: captura para receber o pointerup
              e.currentTarget.setPointerCapture?.(e.pointerId);
              startY.current = e.clientY;
              swiped.current = false;
            }}
            onPointerUp={(e) => {
              if (startY.current == null) return;
              const next = swipeMove(e.clientY - startY.current);
              startY.current = null;
              if (next) {
                swiped.current = true;
                move(next);
              }
            }}
            onClick={() => {
              // o arraste já moveu a gaveta; o click que vem depois não desfaz
              if (swiped.current) swiped.current = false;
              else move("toggle");
            }}
          >
            <span aria-hidden="true" className={s.grip} />
            {peek ?? <span className={s.handleLabel}>{open ? copy.curtainClose : copy.curtainOpen}</span>}
          </button>
        )}
        <div id={bodyId} ref={bodyRef} className={s.body} aria-hidden={open ? undefined : true}>
          {children}
        </div>
      </div>
    </div>
  );
}
