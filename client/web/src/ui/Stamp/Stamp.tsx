import { cx } from "../cx";
import s from "./Stamp.module.css";

export type StampText = "Blefe?" | "Verdade" | "revelada";

/**
 * Carimbo girado. "revelada" é o selo pequeno sobre a carta (Cartas.dc.html);
 * "Blefe?" e "Verdade" são a faixa larga da contestação (ContestaBlefe/Verdade.dc.html).
 * Posicionado em absoluto: o pai precisa de position: relative.
 */
export function Stamp({ text, animate, className }: { text: StampText; animate?: boolean; className?: string }) {
  const kind = text === "revelada" ? s.card : text === "Verdade" ? s.truth : s.bluff;
  return (
    <span data-carimbo className={cx(s.stamp, kind, animate && "m-stamp", className)}>
      {text}
    </span>
  );
}
