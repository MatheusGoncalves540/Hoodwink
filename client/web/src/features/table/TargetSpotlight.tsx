import type { CSSProperties } from "react";
import type { SeatBox } from "../../game/seats";
import { cx } from "../../ui/cx";
import s from "./TargetSpotlight.module.css";

/**
 * Escolha de alvo: a casa apaga (cabeçalho, faixa Você e trilho inclusive) e cada camarote elegível ganha o seu
 * refletor, com feixe até o pé e poça no chão: as mesmas peças do holofote da pergunta global (GlobalQuestion).
 * Os camarotes acesos sobem acima do escuro pelo .targetable (Seat.module.css); este layer só desenha a luz.
 * A geometria vai em variáveis CSS por camarote; cores só de tokens.
 */
export function TargetSpotlight({ boxes, desktop }: { boxes: SeatBox[]; desktop: boolean }) {
  return (
    <div aria-hidden="true" className={cx(s.layer, desktop && s.desk)} data-slot="target-light">
      <div className={cx("m-scrim", s.scrim)} />
      {boxes.map((b) => (
        <span key={`${b.x}:${b.y}`} className={s.light} style={{ "--x": `${b.x}px`, "--y": `${b.y}px`, "--w": `${b.w}px`, "--h": `${b.h}px` } as CSSProperties}>
          <span className={cx("m-cone", s.beam)} />
          <span className={cx("m-cone", s.pool)} />
          <svg className={s.lamp} viewBox="0 0 96 64">
            <path d="M10 4H86" className={s.truss} />
            <path d="M34 4v14M62 4v14" className={s.yoke} />
            <path d="M26 18h44l-7 26H33z" className={s.can} />
            <ellipse cx="48" cy="45" rx="16" ry="4.5" className={s.lens} />
          </svg>
        </span>
      ))}
    </div>
  );
}
