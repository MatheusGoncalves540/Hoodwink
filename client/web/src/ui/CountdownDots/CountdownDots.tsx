import { cx } from "../cx";
import s from "./CountdownDots.module.css";

/**
 * Os pontos da contestação (ContestaBlefe.dc.html): cheios os segundos que restam, o último cheio pisca.
 * Decorativo: o texto ao lado ("6 segundos para escolher") é quem informa o leitor de tela.
 */
export function CountdownDots({ seconds, total }: { seconds: number; total: number }) {
  const left = Math.max(0, Math.min(total, seconds));
  return (
    <span aria-hidden="true" className={s.dots}>
      {Array.from({ length: total }, (_, i) => (
        <span key={i} className={cx(s.dot, i < left ? s.full : s.empty, i === left - 1 && "m-tick")} />
      ))}
    </span>
  );
}
