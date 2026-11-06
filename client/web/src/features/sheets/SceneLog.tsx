import { copy } from "../../game/copy";
import type { GameView } from "../../game/selectors";
import { selectLog } from "../../game/selectors";
import { Label } from "../../ui/Label/Label";
import s from "./SceneLog.module.css";

/** Diário do trilho: as últimas linhas, da mais nova para a mais antiga. */
export function SceneLog({ view, limit = 5 }: { view: GameView; limit?: number }) {
  const rows = selectLog(view).slice(0, limit);
  const turn = view.room?.turn ?? 0;

  return (
    <section aria-label={copy.sceneLog} className={s.log}>
      <div className={s.head}>
        <h2 className={s.title}>{copy.sceneLog}</h2>
        <Label tone="ouro">{copy.roundN(turn)}</Label>
      </div>
      {rows.length > 0 ? (
        <ol className={s.rows}>
          {rows.map((r) => (
            <li key={r.key} className={s.row}>
              <span className={s.round}>{r.turn}</span>
              <span className={s.text}>{r.text}</span>
            </li>
          ))}
        </ol>
      ) : (
        <p className={s.empty}>{copy.nothingSeen}</p>
      )}
    </section>
  );
}
