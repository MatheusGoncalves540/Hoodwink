import type { CSSProperties } from "react";
import type { SeatBox } from "../../game/seats";
import { cx } from "../../ui/cx";
import s from "./TargetSpotlight.module.css";

type Props = {
  /** Camarotes acesos (etapa "quem": todos os elegíveis; etapa "qual carta": só o escolhido). */
  boxes: SeatBox[];
  /** Centro do palco, onde fica o canhão. */
  cx: number;
  desktop: boolean;
};

/**
 * Escolha de alvo: a casa apaga (cabeçalho, faixa Você e trilho inclusive) e um canhão só, no alto do palco,
 * abre o feixe por cima de todos os camarotes elegíveis. Escolhido o alvo, o feixe fecha nele (clip-path anima).
 * Um feixe só: nada se cruza, nem com 9 adversários. Os camarotes acesos sobem acima do escuro pelo .targetable
 * (Seat.module.css); cada um ganha a sua poça no chão. Cores só de tokens.
 */
export function TargetSpotlight({ boxes, cx: center, desktop }: Props) {
  if (boxes.length === 0) return null;
  const lens = desktop ? 38 : 30;
  const left = Math.min(...boxes.map((b) => b.x)) - 10;
  const right = Math.max(...boxes.map((b) => b.x + b.w)) + 10;
  const bottom = Math.max(...boxes.map((b) => b.y + b.h));
  const beam = { clipPath: `polygon(${center - 12}px ${lens}px, ${center + 12}px ${lens}px, ${right}px ${bottom}px, ${left}px ${bottom}px)` };
  return (
    <div aria-hidden="true" className={cx(s.layer, desktop && s.desk)} data-slot="target-light">
      <div className={cx("m-scrim", s.scrim)} />
      <span className={cx("m-cone", s.beam)} style={beam} />
      {boxes.map((b) => (
        <span
          key={`${b.x}:${b.y}`}
          className={cx("m-cone", s.pool)}
          style={{ "--x": `${b.x}px`, "--y": `${b.y}px`, "--w": `${b.w}px`, "--h": `${b.h}px` } as CSSProperties}
        />
      ))}
      <svg className={s.lamp} style={{ left: center }} viewBox="0 0 96 64">
        <path d="M10 4H86" className={s.truss} />
        <path d="M34 4v14M62 4v14" className={s.yoke} />
        <path d="M26 18h44l-7 26H33z" className={s.can} />
        <ellipse cx="48" cy="45" rx="16" ry="4.5" className={s.lens} />
      </svg>
    </div>
  );
}
