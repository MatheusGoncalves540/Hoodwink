import { useEffect, useState, type CSSProperties } from "react";
import type { Layout } from "../../app/useLayout";
import { useLayout } from "../../app/useLayout";
import LoginModal from "../../componentsTSX/loginModal";
import { copy } from "../../game/copy";
import type { selectFinale } from "../../game/selectors";
import { selectFinale as selectFinaleVm } from "../../game/selectors";
import {
  ReplayApiError,
  downloadReplay,
  isAbortError,
  listMyMatches,
  saveReplayFile,
} from "../../net/replayApi";
import { playerIdFromToken, useSessionStore } from "../../store/sessionStore";
import { useAuthStore } from "../../store/authStore";
import { useGameStore } from "../../store/gameStore";
import { BugCard } from "../bugReport/BugCard";
import { draftFor } from "../bugReport/drafts";
import { Button } from "../../ui/Button/Button";
import { Ticket } from "../../ui/Ticket/Ticket";
import { Moldura } from "../../ui/Moldura/Moldura";
import { cx } from "../../ui/cx";
import s from "./Finale.module.css";
import { SupportDev } from "../support/SupportDev";

const REPLAY_POLL_MS = 3000;
const REPLAY_POLL_TIMEOUT_MS = 30000;

type ReplayStatus = "idle" | "checking" | "ready" | "pending" | "unauthorized";

export type FinaleReplay = {
  ready: boolean;
  checking: boolean;
  pending: boolean;
  unauthorized: boolean;
  downloading: boolean;
  error: string;
  onWatch(): void;
  onDownload(): void;
  onRetry(): void;
  onLogin(): void;
};

type Props = {
  vm: ReturnType<typeof selectFinale>;
  layout: Layout;
  onNewRoom(): void;
  onLobby(): void;
  replay?: FinaleReplay;
};

/** Fim da partida (tela 23): cartaz com o vencedor e a ordem de saída. */
export function Finale({ vm, layout, onNewRoom, onLobby, replay }: Props) {
  const desktop = layout === "desktop";
  const size = vm.winnerName.length > 9 ? (desktop ? 96 : 56) : desktop ? 132 : 92;

  return (
    <main className={cx(s.screen, desktop && s.desk)}>
      <Moldura lugar="fim" as="section" aria-label={copy.finaleCart} className={cx("m-hang", s.cartaz)}>
        <p className={s.label}>{copy.finale}</p>
        <h1 className="m-bow" style={{ fontSize: size }}>
          {vm.winnerName}
        </h1>
        <p className={s.won}>{copy.wonAtRound(vm.round, vm.coins)}</p>
        {vm.fala && (
          <p className={s.fala}>
            “{vm.fala}” <span>{copy.proso}</span>
          </p>
        )}
        <h2 className={cx("hw-rotulo", s.orderLabel)}>{copy.exitOrder}</h2>
        <ol className={s.exits} style={{ "--m-base": ".9s" } as CSSProperties}>
          {vm.exits.map((e) => (
            <li key={`${e.place}-${e.name}`} className={cx("m-item", s.exit, e.me && s.me)}>
              <span className={s.place}>{copy.place(e.place)}</span>
              <span className={s.exitName}>{e.name}</span>
              <span className={s.exitRound}>{copy.roundN(e.round)}</span>
            </li>
          ))}
        </ol>
      </Moldura>

      <section aria-label={copy.afterShow} className={cx("m-rise", s.actions)}>
        <Ticket size="l" onClick={onNewRoom}>
          {copy.playAgain}
        </Ticket>
        {replay && (
          <section aria-label={copy.perfil.replayTitle} className={s.replay}>
            <h2 className={cx("hw-rotulo", s.replayTitle)}>{copy.perfil.replayTitle}</h2>
            <div className={s.replayActions}>
              <Button
                variant="outline"
                className={s.replayAction}
                disabled={!replay.ready}
                onClick={replay.onWatch}
              >
                {copy.perfil.watch}
              </Button>
              <Button
                variant="outline"
                className={s.replayAction}
                disabled={!replay.ready || replay.downloading}
                onClick={replay.onDownload}
              >
                {replay.downloading ? copy.perfil.downloading : copy.perfil.download}
              </Button>
            </div>
            {replay.checking && (
              <p role="status" className={s.replayNote}>
                {copy.perfil.replayChecking}
              </p>
            )}
            {replay.pending && (
              <div className={s.replayPending}>
                <p role="status" className={s.replayNote}>
                  {copy.perfil.replayPending}
                </p>
                <p className={s.replayHint}>{copy.perfil.replayPendingHint}</p>
                <Button variant="link" onClick={replay.onRetry}>
                  {copy.perfil.retry}
                </Button>
              </div>
            )}
            {replay.unauthorized && (
              <div className={s.replayPending}>
                <p role="status" className={s.replayNote}>
                  {copy.perfil.needLogin}
                </p>
                <Button variant="link" onClick={replay.onLogin}>
                  {copy.perfil.login}
                </Button>
              </div>
            )}
            {replay.error && (
              <p role="alert" className={s.replayError}>
                {replay.error}
              </p>
            )}
          </section>
        )}
        <Button variant="link" onClick={onLobby}>
          {copy.backToEntrance}
        </Button>
        <SupportDev layout={layout} />
      </section>
    </main>
  );
}

/**
 * Procura o replay recém-gravado em /me/matches por até 30 s (uma consulta a
 * cada 3 s). Não bloqueia o cartaz: a tela aparece inteira e os botões só
 * habilitam quando a partida entra no histórico.
 */
function useReplayOffer(matchId: string, token: string) {
  const [status, setStatus] = useState<ReplayStatus>("idle");
  const [attempt, setAttempt] = useState(0);
  const [downloading, setDownloading] = useState(false);
  const [downloadError, setDownloadError] = useState("");

  useEffect(() => {
    setDownloadError("");
    if (!matchId) {
      setStatus("idle");
      return;
    }
    if (!token) {
      setStatus("unauthorized");
      return;
    }

    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | undefined;
    const controller = new AbortController();
    const deadline = Date.now() + REPLAY_POLL_TIMEOUT_MS;
    // Se a rede pendurar, o prazo ainda fecha em 30 s e libera o "tentar de novo".
    const deadlineTimer = setTimeout(() => {
      cancelled = true;
      controller.abort();
      setStatus("pending");
    }, REPLAY_POLL_TIMEOUT_MS);
    setStatus("checking");

    const check = async () => {
      try {
        const matches = await listMyMatches(token, 30, controller.signal);
        if (cancelled) return;
        if (matches.some((match) => match.match_id === matchId)) {
          clearTimeout(deadlineTimer);
          setStatus("ready");
          return;
        }
      } catch (failure) {
        if (cancelled || isAbortError(failure)) return;
        if (failure instanceof ReplayApiError && (failure.status === 401 || failure.status === 403)) {
          clearTimeout(deadlineTimer);
          if (failure.status === 401) useAuthStore.getState().clearSession();
          setStatus("unauthorized");
          return;
        }
        // Rede/5xx: segue tentando até o prazo, sem travar a tela.
      }

      if (cancelled) return;
      if (Date.now() >= deadline) {
        clearTimeout(deadlineTimer);
        setStatus("pending");
        return;
      }
      timer = setTimeout(() => void check(), REPLAY_POLL_MS);
    };

    void check();
    return () => {
      cancelled = true;
      controller.abort();
      clearTimeout(timer);
      clearTimeout(deadlineTimer);
    };
  }, [matchId, token, attempt]);

  const download = async () => {
    if (!matchId || !token || downloading) return;
    setDownloading(true);
    setDownloadError("");
    try {
      saveReplayFile(await downloadReplay(token, matchId));
    } catch (failure) {
      if (failure instanceof ReplayApiError && failure.status === 401) {
        useAuthStore.getState().clearSession();
      }
      setDownloadError(failure instanceof Error ? failure.message : copy.perfil.downloadError);
    } finally {
      setDownloading(false);
    }
  };

  return {
    status,
    downloading,
    error: downloadError,
    retry: () => setAttempt((n) => n + 1),
    download: () => void download(),
    watch: () => {
      if (typeof window === "undefined") return;
      window.location.hash = `#replay?match=${encodeURIComponent(matchId)}`;
    },
  };
}

const clearRoomUrl = () => {
  if (typeof window === "undefined") return;
  const url = new URL(window.location.href);
  url.searchParams.delete("sala");
  url.searchParams.delete("cena");
  url.searchParams.delete("fixture");
  window.history.replaceState(null, "", url);
};

/** Liga o cartaz ao estado: jogar novamente usa o login da sessão; voltar limpa tudo. */
export function FinaleScene() {
  const me = useGameStore((st) => st.me);
  const room = useGameStore((st) => st.room);
  const log = useGameStore((st) => st.log);
  const jwtToken = useAuthStore((st) => st.jwtToken);
  const openLogin = useAuthStore((st) => st.openLogin);
  const rematch = useSessionStore((st) => st.rematch);
  const layout = useLayout();
  const vm = selectFinaleVm({ me, room, log });

  const onNewRoom = async () => {
    if (!jwtToken || !room?.id) {
      onLobby();
      return;
    }
    const code = await rematch(jwtToken, room.id);
    if (!code) return;
    const url = new URL(window.location.href);
    url.searchParams.delete("cena");
    url.searchParams.delete("fixture");
    url.searchParams.set("sala", code);
    window.history.replaceState(null, "", url);
    useGameStore.getState().reset();
  };

  const onLobby = () => {
    useGameStore.getState().reset();
    useSessionStore.getState().reset();
    clearRoomUrl();
  };

  const userId = playerIdFromToken(jwtToken);
  const matchId = room?.matchId ?? "";
  const offer = useReplayOffer(matchId, jwtToken);
  const replay: FinaleReplay | undefined = matchId
    ? {
        ready: offer.status === "ready",
        checking: offer.status === "checking",
        pending: offer.status === "pending",
        unauthorized: offer.status === "unauthorized",
        downloading: offer.downloading,
        error: offer.error,
        onWatch: offer.watch,
        onDownload: offer.download,
        onRetry: offer.retry,
        onLogin: openLogin,
      }
    : undefined;
  const draft = matchId ? draftFor(matchId, userId) : undefined;
  const [dismissed, setDismissed] = useState<string | null>(null);
  const bugOpen = !!draft && draft.reportId !== dismissed;
  const closeBugCard = () => {
    if (draft) setDismissed(draft.reportId);
  };

  return (
    <>
      <Finale vm={vm} layout={layout} onNewRoom={onNewRoom} onLobby={onLobby} replay={replay} />
      <LoginModal />
      {draft && <BugCard open={bugOpen} onClose={closeBugCard} mode={layout} draft={draft} token={jwtToken} />}
    </>
  );
}
