import { useEffect, useRef, useState } from "react";

export type Countdown = {
  /** Segundos restantes; vira 0 no mesmo instante em que o prazo acaba. */
  left: number;
  /** Fração da janela que ainda falta, de 0 a 1 (a barra usa esta mesma leitura). */
  progress: number;
};

/**
 * Número e barra saem da mesma leitura de `Date.now()`, a cada 100 ms.
 *
 * O intervalo de 1 s antigo não caía nos segundos cheios: o número ficava parado no
 * total por até 2 s e, no fim, ainda mostrava 1 depois de a barra zerar. O `span` é a
 * janela real (nunca menor que o que falta no primeiro render), então relógio adiantado
 * ou evento que chega atrasado não congelam a barra no cheio.
 */
export function useCountdown(expiresAt: number | undefined, total: number): Countdown {
  const span = useRef({ key: Number.NaN, ms: Math.max(0, total) * 1000 });
  const key = expiresAt ?? Number.NaN;

  const read = (): Countdown => {
    if (expiresAt === undefined) return { left: total, progress: 1 };
    if (span.current.key !== key) {
      span.current = { key, ms: Math.max(total * 1000, expiresAt - Date.now()) };
    }
    const remaining = Math.max(0, expiresAt - Date.now());
    return {
      left: remaining <= 0 ? 0 : Math.ceil(remaining / 1000),
      progress: span.current.ms > 0 ? Math.min(1, remaining / span.current.ms) : 0,
    };
  };

  const [state, setState] = useState<Countdown>(read);

  useEffect(() => {
    setState(read());
    if (expiresAt === undefined) return;
    const id = setInterval(() => setState(read()), 100);
    return () => clearInterval(id);
  }, [expiresAt, total]);

  return state;
}
