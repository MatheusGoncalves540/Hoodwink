import type { Layout } from "../../app/useLayout";
import { cx } from "../../ui/cx";
import { STAGE_DESK_HTML, STAGE_HTML } from "./sudden";
import s from "./SuddenDeathStage.module.css";

/** Mesa em morte súbita (tela 22): cenário da prancha MesaSubita, sempre escuro. */
export function SuddenDeathStage({ layout }: { layout: Layout }) {
  const html = layout === "desktop" ? STAGE_DESK_HTML : STAGE_HTML;
  return <div aria-hidden="true" className={cx(s.stage, layout === "desktop" && s.desk)} dangerouslySetInnerHTML={{ __html: html }} />;
}
