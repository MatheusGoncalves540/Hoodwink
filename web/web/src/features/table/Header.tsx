import type { Layout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import { cx } from "../../ui/cx";
import s from "./Header.module.css";

type Props = {
  turn: number;
  tax: number;
  unread: number;
  layout: Layout;
  sudden?: boolean;
  onMenu?(): void;
  onChat?(): void;
};

/** Cabeçalho (52 px mobile, 64 px desktop): menu, rodada/taxa e conversa com o contador de novas. */
export function Header({ turn, tax, unread, layout, sudden, onMenu, onChat }: Props) {
  return (
    <header className={cx(s.header, layout === "desktop" && s.desk)}>
      <button type="button" className={s.icon} aria-label={copy.menu} onClick={onMenu}>
        <svg aria-hidden="true" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
          <path d="M4 7h16M4 12h16M4 17h16" />
        </svg>
      </button>

      {sudden ? (
        <span className={s.suddenHeader}>
          {copy.roundSudden(turn)} <span className={s.sudden}>{copy.suddenDeathLower}</span>
        </span>
      ) : (
        <span className={cx("hw-rotulo", s.round)}>{copy.roundTax(turn, tax)}</span>
      )}

      <button type="button" className={s.icon} aria-label={copy.chatButton(unread)} onClick={onChat}>
        <svg aria-hidden="true" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinejoin="round">
          <path d="M4 5h16v11H9l-5 4z" />
        </svg>
        {unread > 0 && <span className={s.badge}>{unread}</span>}
      </button>
    </header>
  );
}
