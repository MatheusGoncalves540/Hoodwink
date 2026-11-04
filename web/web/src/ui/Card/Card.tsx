/**
 * Cartas (02-design-system.md → Cartas: seis estados; Cartas.dc.html, Mesa6, Voce, ContestaBlefe).
 * Tamanhos: s 78×110 (faixa Você), m 112 de altura e largura do pai (gaveta Você), l 150 de altura (escolha
 * na contestação), xl 132×186 (catálogo). A prancha Mesa6 usa 78×110 na faixa Você, não os 64 px do doc.
 */
import type { CSSProperties, HTMLAttributes, ReactNode } from "react";
import type { RoleId } from "../../game/protocol";
import { roleName } from "../../game/roles";
import { Stamp, type StampText } from "../Stamp/Stamp";
import { cx } from "../cx";
import s from "./Card.module.css";

export type CardSize = "s" | "m" | "l" | "xl" | "d";

type Base = HTMLAttributes<HTMLElement> & { size: CardSize; as?: "div" | "button" };

type FaceProps = Base & {
  role: RoleId;
  /** Abaixo do nome, na faixa de cima ("a mesa vê: carta protegida"). */
  top?: ReactNode;
  /** Faixa de baixo ("protegida", descrição do catálogo). */
  bottom?: ReactNode;
  /** URL da arte do papel; sem ela, o listrado fixo da .hw-carta. */
  art?: string;
  /** Moeda no alto da faixa de baixo; null = papel sem custo (losango). */
  cost?: number | null;
  protected?: boolean;
  stamp?: StampText;
};

/**
 * Carta Cortina. A .hw-carta usa column-reverse: o último filho é a faixa de cima e o primeiro, a de baixo.
 * A ordem dos filhos abaixo depende disso.
 */
export function CardFace({ role, top, bottom, art, cost, protected: prot, stamp, size, as: Tag = "div", className, style, ...rest }: FaceProps) {
  const hasBottom = bottom != null || cost !== undefined;
  // ponytail: a arte entra por variável inline (URL dinâmica); o resto do visual está no módulo.
  const artStyle = art ? ({ ...style, "--arte": `url(${JSON.stringify(art)})` } as CSSProperties) : style;
  return (
    <Tag {...rest} style={artStyle} className={cx("hw-carta", s.face, s[size], prot && s.protected, art && s.art, className)}>
      {hasBottom && (
        <div className={cx(s.bottom, cost !== undefined && s.withCost)}>
          {cost === null ? <span className={s.diamond} /> : cost !== undefined && <span className={cx("hw-num", s.cost)}>{cost}</span>}
          {bottom}
        </div>
      )}
      {stamp && <Stamp text={stamp} />}
      <div className={s.top}>
        <span className={s.name}>{roleName(role)}</span>
        {top && <span className={s.note}>{top}</span>}
      </div>
    </Tag>
  );
}

/** Verso: baralho e mão alheia. */
export function CardBack({ size, as: Tag = "div", className, ...rest }: Base) {
  return <Tag {...rest} className={cx(s.card, s.back, s[size], className)} />;
}

/** Carta de outro, vista aberta: tracejada, poça de luz e "?". */
export function CardHidden({ size, label, protected: prot, as: Tag = "div", className, ...rest }: Base & { label?: ReactNode; protected?: boolean }) {
  return (
    <Tag {...rest} className={cx(s.card, s.hidden, s[size], prot && s.protected, className)}>
      {label && <span className={s.hiddenLabel}>{label}</span>}
      <span aria-hidden="true" className={s.ask}>
        ?
      </span>
    </Tag>
  );
}

/** Morta: filete cortado pela faixa diagonal de alerta. O papel, se conhecido, vai no pé. */
export function CardDead({ size, role, as: Tag = "div", className, ...rest }: Base & { role?: RoleId }) {
  return (
    <Tag {...rest} className={cx(s.card, s.dead, s[size], className)}>
      {role && <span className={s.deadName}>{roleName(role)}</span>}
    </Tag>
  );
}

export type MiniState = "back" | "revealed" | "protected" | "dead";

/** Miniatura do camarote. Decorativa: o aria-label do camarote descreve tudo. */
export function CardMini({ state, size }: { state: MiniState; role?: RoleId; size: "l" | "s" | "desk-l" | "desk-s" }) {
  return <span aria-hidden="true" className={cx(s.mini, s[`mini-${size}`], s[`mini-${state}`])} />;
}
