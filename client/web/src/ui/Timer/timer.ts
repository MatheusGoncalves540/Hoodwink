import { useEffect, useRef, useState } from "react";

export type Countdown = {
  /** Segundos restantes; vira 0 no mesmo instante em que o prazo acaba. */
  left: number;
  /** Fração da janela que ainda falta, de 0 a 1 (a barra usa esta mesma leitura). */
  progress: number;
};

/** Número e barra do mesmo `now`: o número arredonda para cima e a barra usa a fração exata. */
export function countdown(expiresAt: number | undefined, total: number, now: number, windowMs: number): Countdown {
  if (expiresAt === undefined) return { left: total, progress: 1 };
  const remaining = Math.max(0, expiresAt - now);
  return {
    left: remaining <= 0 ? 0 : Math.ceil(remaining / 1000),
    progress: windowMs > 0 ? Math.min(1, remaining / windowMs) : 0,
  };
}

/**
 * Número e barra saem da mesma leitura de `Date.now()`, a cada 250 ms.
 *
 * O intervalo de 1 s antigo não caía nos segundos cheios: o número ficava parado no
 * total por até 2 s e, no fim, ainda mostrava 1 depois de a barra zerar. O `windowMs`
 * é a janela real (nunca menor que o que falta no primeiro render), então relógio
 * adiantado ou evento que chega atrasado não congelam a barra no cheio.
 */
export function useCountdown(expiresAt: number | undefined, total: number): Countdown {
  const windowRef = useRef({ key: Number.NaN, ms: Math.max(0, total) * 1000 });
  const key = expiresAt ?? Number.NaN;

  const read = (): Countdown => {
    if (expiresAt === undefined) return countdown(undefined, total, Date.now(), windowRef.current.ms);
    if (windowRef.current.key !== key) {
      windowRef.current = { key, ms: Math.max(total * 1000, expiresAt - Date.now()) };
    }
    return countdown(expiresAt, total, Date.now(), windowRef.current.ms);
  };

  const [state, setState] = useState<Countdown>(read);

  useEffect(() => {
    const tick = () => setState(read());
    tick();
    if (expiresAt === undefined) return;

    const id = setInterval(tick, 250);
    // Voltar de outra aba atrasa o intervalo; reaproveita o mesmo cálculo absoluto.
    document.addEventListener("visibilitychange", tick);
    window.addEventListener("focus", tick);
    return () => {
      clearInterval(id);
      document.removeEventListener("visibilitychange", tick);
      window.removeEventListener("focus", tick);
    };
  }, [expiresAt, total]);

  return state;
}
