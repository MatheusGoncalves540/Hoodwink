import type { ReactNode } from "react";
import { copy } from "../../game/copy";
import { cx } from "../cx";
import s from "./Timer.module.css";
import { useCountdown } from "./timer";

type Props = {
  /** Fim da janela, em ms (gameEvent.expiresAt). */
  expiresAt: number;
  /** Duração cheia da janela, em segundos (10, ou 5 na morte súbita). */
  total: number;
  /** aria-label da barra: "Tempo da jogada de Carla". */
  label: string;
  tone?: "tinta" | "alerta";
  /** Janela de decisão: mostra "restam 7?" no lugar de "7 s". */
  decision?: boolean;
  /** Texto à esquerda do número ("Vez de Carla. Se declarar, você reage."). */
  caption?: ReactNode;
};

/** Número e barra saem do mesmo relógio (useCountdown): os dois zeram juntos, no prazo. */
export function Timer({ expiresAt, total, label, tone = "tinta", decision, caption }: Props) {
  const { left, progress } = useCountdown(expiresAt, total);

  return (
    <div className={cx("m-deixa", s.timer)}>
      <div className={cx(s.row, !caption && s.end)}>
        {caption && <span className={s.caption}>{caption}</span>}
        {decision ? (
          <span className={cx("hw-rotulo", s.decision)}>
            {copy.left}{" "}
            <b className="hw-num">
              {left}
              {/* o "?" treme (tira rápida, 3 desenhos); o número não treme */}
              <span className={cx("m-tira", s.ask)}>
                <span>
                  <i>?</i>
                  <i>?</i>
                  <i>?</i>
                </span>
              </span>
            </b>
          </span>
        ) : (
          <span className={s.secs}>
            <b className="hw-num">{left}</b> {copy.seconds}
          </span>
        )}
      </div>
      <div
        role="progressbar"
        aria-label={label}
        aria-valuemin={0}
        aria-valuemax={Math.max(total, left)}
        aria-valuenow={left}
        className={s.track}
      >
        <div className={cx(s.fill, s[tone])} style={{ transform: `scaleX(${progress})` }} />
      </div>
    </div>
  );
}
