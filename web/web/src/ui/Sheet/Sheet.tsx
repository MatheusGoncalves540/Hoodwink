/**
 * Gaveta (02-design-system.md → Gaveta; Camarote/CamaroteDesk.dc.html). Radix Dialog cuida de Esc, foco e aria.
 * - mobile: sobe de baixo, canhoto (.stub), véu e foco preso (modal);
 * - desktop: painel de 420 px no trilho da direita, sem véu; a mesa continua clicável (não modal).
 * Fechar devolve o foco a quem abriu (FocusScope do Radix).
 */
import { Dialog } from "radix-ui";
import { useRef, type ReactNode } from "react";
import { copy } from "../../game/copy";
import { Button } from "../Button/Button";
import { cx } from "../cx";
import s from "./Sheet.module.css";

type Props = {
  open: boolean;
  onClose(): void;
  /** Nome acessível da gaveta: "Camarote de Bruno", "Menu da sala". */
  label: string;
  mode: "mobile" | "desktop";
  children: ReactNode;
  /** Pé fixo acima do "☜ Voltar" (campo do chat, por exemplo). */
  footer?: ReactNode;
  /** Texto do botão de voltar; padrão "☜ Voltar à mesa". */
  backLabel?: string;
};

// ponytail: sem animação de saída (o Radix desmonta na hora); entra com m-sheet/m-rail.
export function Sheet({ open, onClose, label, mode, children, footer, backLabel = copy.backToTable }: Props) {
  const mobile = mode === "mobile";
  // Aberta por estado (sem Dialog.Trigger), o Radix não sabe a quem devolver o foco: guardamos quem abriu.
  const opener = useRef<HTMLElement | null>(null);
  return (
    <Dialog.Root open={open} onOpenChange={(o) => !o && onClose()} modal={mobile}>
      <Dialog.Portal>
        {mobile && <Dialog.Overlay className={cx("m-scrim", s.scrim)} />}
        <Dialog.Content
          aria-describedby={undefined}
          className={cx("stub", mobile ? cx("m-sheet", s.sheet) : s.rail)}
          onInteractOutside={mobile ? undefined : (e) => e.preventDefault()}
          onOpenAutoFocus={() => {
            opener.current = document.activeElement as HTMLElement | null;
          }}
          onCloseAutoFocus={(e) => {
            e.preventDefault();
            opener.current?.focus();
          }}
        >
          <Dialog.Title className={s.hidden}>{label}</Dialog.Title>
          <div className={s.body}>{children}</div>
          {footer}
          <Dialog.Close asChild>
            <Button variant="link" className={s.back}>
              {backLabel}
            </Button>
          </Dialog.Close>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
