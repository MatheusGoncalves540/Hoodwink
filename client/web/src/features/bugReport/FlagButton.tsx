/**
 * Botão "Algo deu errado" da mesa (plano 06, T6): mora no menu da sala e no trilho do
 * desktop, nunca no lugar de uma ação de jogo. O primeiro toque só abre a confirmação;
 * "Marcar" chama POST /bugReports/flag e só então grava o rascunho.
 */
import { useState } from "react";
import type { Layout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import { BugApiError, flagBug } from "../../net/bugApi";
import { useAuthStore } from "../../store/authStore";
import { playerIdFromToken } from "../../store/sessionStore";
import { Button } from "../../ui/Button/Button";
import { Sheet } from "../../ui/Sheet/Sheet";
import { addFlag } from "./drafts";
import s from "./BugReport.module.css";

type Props = {
  mode: Layout;
  matchId: string;
  turn: number;
};

export function FlagButton({ mode, matchId, turn }: Props) {
  const jwtToken = useAuthStore((st) => st.jwtToken);
  const openLogin = useAuthStore((st) => st.openLogin);
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [marked, setMarked] = useState(false);

  // Sem matchId no estado da sala (plano 05) o botão não existe.
  if (!matchId) return null;

  const confirm = async () => {
    if (!jwtToken) {
      setOpen(false);
      openLogin();
      return;
    }
    setBusy(true);
    setMessage(null);
    try {
      const { reportId, turns } = await flagBug(jwtToken, { matchId, turn });
      addFlag({ reportId, matchId, turns, at: Date.now(), userId: playerIdFromToken(jwtToken) });
      setOpen(false);
      setMarked(true);
    } catch (e) {
      setMessage(e instanceof BugApiError && e.result === "limit" ? copy.bugReportLimit : copy.bugReportError);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className={s.flag}>
      <Button variant="outline" onClick={() => setOpen(true)}>
        {copy.bugReportButton}
      </Button>

      {marked && (
        <p role="status" className={s.flagMarked}>
          {copy.bugReportMarked}
        </p>
      )}

      <Sheet
        open={open}
        onClose={() => setOpen(false)}
        label={copy.bugReportButton}
        mode={mode}
        backLabel={copy.bugReportConfirmNo}
      >
        <div className={s.confirm}>
          <p className={s.confirmText}>{copy.bugReportConfirm}</p>
          <div className={s.confirmRow}>
            <Button variant="outline" disabled={busy} onClick={confirm}>
              {copy.bugReportConfirmYes}
            </Button>
          </div>
          {message && (
            <p role="alert" className={s.error}>
              {message}
            </p>
          )}
        </div>
      </Sheet>
    </div>
  );
}
