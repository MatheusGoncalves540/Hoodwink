/**
 * Cartão de detalhes do relato (plano 06, T6): 6 categorias, texto opcional (obrigatório
 * no Lobby), chave dos dados do aparelho e Enviar/Pular. Com rascunho faz PATCH; sem
 * rascunho (Lobby) faz POST. 404/403 apaga e avisa; 429 e rede mantêm texto e rascunho.
 */
import { useState } from "react";
import type { Layout } from "../../app/useLayout";
import { copy, type BugCategory } from "../../game/copy";
import { BugApiError, patchBugReport, sendBugReport } from "../../net/bugApi";
import { collectBugContext } from "../../net/bugContext";
import { Button } from "../../ui/Button/Button";
import { Sheet } from "../../ui/Sheet/Sheet";
import { TextField } from "../../ui/TextField/TextField";
import { Ticket } from "../../ui/Ticket/Ticket";
import { Toggle } from "../../ui/Toggle/Toggle";
import { cx } from "../../ui/cx";
import { clear, type BugDraft } from "./drafts";
import s from "./BugReport.module.css";

type Props = {
  open: boolean;
  onClose(): void;
  mode: Layout;
  /** Presente = detalha um relato de partida (fim ou pendência no Lobby). */
  draft?: BugDraft;
  /** Token do backend; vazio = anônimo (só o relato novo do Lobby). */
  token: string;
};

export function BugCard({ open, onClose, mode, draft, token }: Props) {
  const lobby = !draft;
  const [category, setCategory] = useState<BugCategory | null>(null);
  const [description, setDescription] = useState("");
  const [sendDevice, setSendDevice] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [gone, setGone] = useState(false);

  const submit = async () => {
    const text = description.trim();
    if (lobby) {
      if (!category) return setError(copy.bugReportChooseCategory);
      if (text.length < 10) return setError(copy.bugReportShortDescription);
    }

    setBusy(true);
    setError(null);
    const context = collectBugContext(sendDevice);

    try {
      if (draft) {
        await patchBugReport(token, draft.reportId, { category, description: text || null, context });
        clear(draft.reportId);
      } else {
        await sendBugReport(token, { category: category!, description: text, context });
      }
      onClose();
    } catch (e) {
      if (e instanceof BugApiError) {
        if (e.status === 403 || e.status === 404) {
          if (draft) clear(draft.reportId);
          setGone(true);
        } else if (e.result === "limit") {
          setError(copy.bugReportLimit);
        } else if (e.result === "unauthorized") {
          onClose();
        } else {
          setError(copy.bugReportError);
        }
      } else {
        setError(copy.bugReportError);
      }
    } finally {
      setBusy(false);
    }
  };

  const skip = () => {
    if (draft) clear(draft.reportId);
    onClose();
  };

  return (
    <Sheet open={open} onClose={onClose} label={copy.bugReportLobby} mode={mode}>
      {gone ? (
        <p role="alert" className={s.error}>
          {copy.bugReportGone}
        </p>
      ) : (
        <div className={s.card}>
          <h2 className={s.title}>{draft ? copy.bugReportQuestion(draft.turns) : copy.bugReportLobbyTitle}</h2>

          <fieldset className={s.group}>
            <legend className={s.legend}>{copy.bugReportCategoriesLabel}</legend>
            <div className={s.categories}>
              {copy.bugReportCategories.map((cat) => (
                <button
                  key={cat}
                  type="button"
                  className={cx(s.category, category === cat && s.categoryOn)}
                  aria-pressed={category === cat}
                  onClick={() => setCategory(cat)}
                >
                  {copy.bugReportCategoryLabels[cat]}
                </button>
              ))}
            </div>
          </fieldset>

          <TextField
            label={copy.bugReportDescriptionLabel}
            placeholder={copy.bugReportDescriptionPlaceholder}
            value={description}
            onChange={setDescription}
            onSubmit={submit}
            maxLength={2000}
          />

          <Toggle label={copy.bugReportDevice} hint={copy.bugReportDeviceHint} value={sendDevice} onChange={setSendDevice} />

          {error && (
            <p role="alert" className={s.error}>
              {error}
            </p>
          )}

          <div className={s.actions}>
            <Ticket disabled={busy} onClick={submit}>
              {busy ? copy.bugReportSending : copy.bugReportSend}
            </Ticket>
            {draft && (
              <Button variant="outline" disabled={busy} onClick={skip}>
                {copy.bugReportSkip}
              </Button>
            )}
          </div>
        </div>
      )}
    </Sheet>
  );
}
