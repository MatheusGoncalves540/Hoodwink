import { useState } from "react";
import type { Layout } from "../../app/useLayout";
import { FEATURES } from "../../config/features";
import { copy } from "../../game/copy";
import { Button } from "../../ui/Button/Button";
import { Sheet } from "../../ui/Sheet/Sheet";
import { Ticket } from "../../ui/Ticket/Ticket";
import { cx } from "../../ui/cx";
import { watchSupportVideo } from "./supportVideo";
import s from "./SupportDev.module.css";

type Step = "invite" | "watching" | "thanks" | "unavailable";

/**
 * "Apoie o dev": link discreto (entrada, sala de espera e fim da partida; nunca na mesa, onde um vídeo
 * faria a pessoa perder a vez) que abre a gaveta do convite. O vídeo não dá vantagem no jogo.
 * Atrás de FEATURES.supportVideo: desligado, não renderiza nada.
 */
export function SupportDev({ layout, enabled = FEATURES.supportVideo, className }: { layout: Layout; enabled?: boolean; className?: string }) {
  const [step, setStep] = useState<Step | null>(null);
  if (!enabled) return null;

  const watch = async () => {
    setStep("watching");
    setStep((await watchSupportVideo()) ? "thanks" : "unavailable");
  };

  return (
    <>
      <Button variant="link" className={cx(s.open, className)} onClick={() => setStep("invite")}>
        <svg aria-hidden="true" width="16" height="14" viewBox="0 0 16 14" className={s.heart}>
          <path d="M8 13S1 8.6 1 4.4A3.4 3.4 0 0 1 8 3a3.4 3.4 0 0 1 7 1.4C15 8.6 8 13 8 13z" />
        </svg>
        {copy.support.open}
      </Button>
      {step && (
        <Sheet open onClose={() => setStep(null)} label={copy.support.kicker} mode={layout} backLabel={copy.support.back}>
          <div className={s.body}>
            <p className={cx("hw-rotulo", s.kicker)}>{copy.support.kicker}</p>
            {step === "thanks" ? (
              <>
                <h2 className={s.title}>{copy.support.thanksTitle}</h2>
                <p>{copy.support.thanks}</p>
              </>
            ) : step === "unavailable" ? (
              <p role="status">{copy.support.unavailable}</p>
            ) : (
              <>
                <h2 className={s.title}>{copy.support.title}</h2>
                <p>{copy.support.body}</p>
                <p className={s.note}>{copy.support.note}</p>
                <Ticket size="l" disabled={step === "watching"} onClick={watch}>
                  {step === "watching" ? copy.support.watching : copy.support.watch}
                </Ticket>
              </>
            )}
          </div>
        </Sheet>
      )}
    </>
  );
}
