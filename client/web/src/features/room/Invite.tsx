/**
 * Convite da sala (Sala.dc.html → seção "Convite").
 * O ingresso é escuro (classe global `ticket`); o botão usa currentColor para aparecer sobre ele.
 */
import { useEffect, useState, type CSSProperties } from "react";
import { copy } from "../../game/copy";
import { Button } from "../../ui/Button/Button";
import { cx } from "../../ui/cx";
import type { Layout } from "../../app/useLayout";
import s from "./Invite.module.css";

type Props = { code: string; layout: Layout; className?: string };

export function Invite({ code, layout, className }: Props) {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!copied) return;
    const timer = setTimeout(() => setCopied(false), 2000);
    return () => clearTimeout(timer);
  }, [copied]);

  const copyInvite = () => {
    if (typeof window === "undefined") return;
    const url = new URL(window.location.href);
    url.searchParams.delete("cena");
    url.searchParams.delete("fixture");
    url.searchParams.set("sala", code);
    navigator.clipboard?.writeText(url.toString()).then(
      () => setCopied(true),
      () => setCopied(false),
    );
  };

  return (
    <section
      aria-label={copy.inviteRegion}
      className={cx("ticket", "m-rise", s.invite, layout === "desktop" && s.desk, className)}
    >
      <div className={s.room}>
        <span className={s.label}>{copy.roomNumber}</span>
        <span className={s.code} style={{ "--len": Math.max(code.length, 4) } as CSSProperties}>{code}</span>
      </div>
      <Button variant="outline" className={s.copy} onClick={copyInvite}>
        {copied ? copy.inviteCopied : copy.copyInvite}
      </Button>
    </section>
  );
}
