import type { HTMLAttributes, ReactNode } from "react";
import { cx } from "../cx";

export type Lugar = "lobby" | "elenco" | "fim" | "aviso" | "pergunta";

type Props = HTMLAttributes<HTMLElement> & { lugar: Lugar; as?: "div" | "section"; children: ReactNode };

/**
 * Moldura trocável por skin (plans/03-skins.md): três quadros de contorno (.hw-moldura-f) + o conteúdo.
 * O desenho mora no CSS da peça (padrão em styles/components.css); o componente só marca o lugar.
 * Lobby = letreiro de lâmpadas na borda (trava 3). Elenco, fim e aviso = filete fervendo (opção B, provisória).
 */
export function Moldura({ lugar, as: Tag = "div", className, children, ...rest }: Props) {
  return (
    <Tag {...rest} data-lugar={lugar} className={cx("hw-moldura", "m-contorno", className)}>
      <i aria-hidden="true" className="hw-moldura-f m-f" />
      <i aria-hidden="true" className="hw-moldura-f m-f" />
      <i aria-hidden="true" className="hw-moldura-f m-f" />
      {children}
    </Tag>
  );
}
