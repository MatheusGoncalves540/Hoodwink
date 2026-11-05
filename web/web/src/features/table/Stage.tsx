import type { Layout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import type { selectStage } from "../../game/selectors";
import { ProsoLine } from "../../ui/ProsoLine/ProsoLine";
import { cx } from "../../ui/cx";
import s from "./Stage.module.css";

/** Palco: nome de quem está em cena (Limelight grande) e a fala do Proso, quando há. */
export function Stage({ vm, layout }: { vm: ReturnType<typeof selectStage>; layout: Layout }) {
  return (
    <section aria-label={copy.stage} aria-live="polite" className={cx("m-rise", s.stage, layout === "desktop" && s.desk)}>
      <h1 className={s.name}>{vm.activeName}</h1>
      <ProsoLine text={vm.fala} className={s.fala} />
    </section>
  );
}
