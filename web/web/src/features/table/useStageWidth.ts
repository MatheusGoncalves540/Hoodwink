import { useEffect, useRef, useState } from "react";

/** Largura real do palco (o mobile é 390; o desktop desconta o trilho). SSR começa no valor de referência. */
export function useStageWidth(initial = 390) {
  const ref = useRef<HTMLDivElement>(null);
  const [width, setWidth] = useState(initial);

  useEffect(() => {
    const el = ref.current;
    if (!el || typeof ResizeObserver === "undefined") return;
    const observer = new ResizeObserver(([entry]) => setWidth(entry.contentRect.width));
    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  return [ref, width] as const;
}
