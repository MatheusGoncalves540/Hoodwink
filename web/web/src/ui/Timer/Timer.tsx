import { useEffect, useState, type CSSProperties, type ReactNode } from "react";
import { copy } from "../../game/copy";
import { cx } from "../cx";
import s from "./Timer.module.css";

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

const secondsLeft = (expiresAt: number, total: number) => Math.max(0, Math.min(total, Math.ceil((expiresAt - Date.now()) / 1000)));

/** Número (atualiza a cada segundo, um intervalo por timer) e barra que encolhe só por CSS (.m-timer). */
export function Timer({ expiresAt, total, label, tone = "tinta", decision, caption }: Props) {
  const [left, setLeft] = useState(() => secondsLeft(expiresAt, total));

  useEffect(() => {
    setLeft(secondsLeft(expiresAt, total));
    const id = setInterval(() => setLeft(secondsLeft(expiresAt, total)), 1000);
    return () => clearInterval(id);
  }, [expiresAt, total]);

  return (
    <div className={s.timer}>
      <div className={cx(s.row, !caption && s.end)}>
        {caption && <span className={s.caption}>{caption}</span>}
        {decision ? (
          <span className={cx("hw-rotulo", s.decision)}>
            {copy.left}{" "}
            <b className="hw-num">
              {left}
              <span className={s.ask}>?</span>
            </b>
          </span>
        ) : (
          <span className={s.secs}>
            <b className="hw-num">{left}</b> {copy.seconds}
          </span>
        )}
      </div>
      <div role="progressbar" aria-label={label} aria-valuemin={0} aria-valuemax={total} aria-valuenow={left} className={s.track}>
        <Fill key={expiresAt} expiresAt={expiresAt} total={total} tone={tone} />
      </div>
    </div>
  );
}

/** Recalcula ponto de partida e duração só quando a janela muda (key = expiresAt); depois é só CSS. */
function Fill({ expiresAt, total, tone }: { expiresAt: number; total: number; tone: "tinta" | "alerta" }) {
  const [style] = useState(() => {
    const ms = Math.max(0, Math.min(total * 1000, expiresAt - Date.now()));
    // ponytail: valores dinâmicos da animação vão inline (CSS var + duração); o resto está no módulo.
    return { "--de": ms / (total * 1000), animationDuration: `${ms}ms` } as CSSProperties;
  });
  return <div className={cx("m-timer", s.fill, s[tone])} style={style} />;
}
