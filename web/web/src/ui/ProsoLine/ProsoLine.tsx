import { copy } from "../../game/copy";
import { cx } from "../cx";
import s from "./ProsoLine.module.css";

/** Fala do narrador entre aspas curvas, seguida de "Proso". Some quando a fala está vazia. */
export function ProsoLine({ text, className }: { text?: string; className?: string }) {
  if (!text) return null;
  return (
    <p className={cx(s.line, className)}>
      “{text}” <span className={s.name}>{copy.proso}</span>
    </p>
  );
}
