/**
 * Tela `#perfil` (plano 05, T15): histórico das partidas do usuário.
 *
 * O avatar do Lobby abre esta tela. A lista vem de `GET /me/matches` com o JWT
 * no Authorization; cada partida ganha "Assistir" (`#replay?match=<id>`) e
 * "Baixar" (o `.hwr` assinado). Datas são curtas, em PT-BR, e nomes entram como
 * texto do React, nunca como HTML.
 *
 * Estados: logado-fora (CTA de login), carregando, vazio, erro e lista com
 * "Carregar mais" (30 por vez, até o teto de 100 do backend).
 */
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useLayout } from "../../app/useLayout";
import LoginModal from "../../componentsTSX/loginModal";
import { copy } from "../../game/copy";
import {
  MATCHES_MAX_LIMIT,
  MATCHES_PAGE_SIZE,
  ReplayApiError,
  downloadReplay,
  isAbortError,
  listMyMatches,
  saveReplayFile,
  type MyMatch,
} from "../../net/replayApi";
import { useAuthStore } from "../../store/authStore";
import { playerIdFromToken, usernameFromToken } from "../../store/sessionStore";
import { Button } from "../../ui/Button/Button";
import { Moldura } from "../../ui/Moldura/Moldura";
import { Ticket } from "../../ui/Ticket/Ticket";
import { cx } from "../../ui/cx";
import s from "./Profile.module.css";

const dateFormatter = new Intl.DateTimeFormat("pt-BR", {
  day: "2-digit",
  month: "2-digit",
  year: "numeric",
});

/** Data curta PT-BR (`06/11/2026`); valor inválido vira o texto de apoio. */
export function formatMatchDate(iso: string): string {
  const date = new Date(iso);
  return Number.isNaN(date.getTime()) ? copy.perfil.unknownDate : dateFormatter.format(date);
}

/** Nome exibido para um id: "Você" para o usuário atual, "Jogador xxxx" para os demais. */
export function playerLabel(id: string, meId: string): string {
  if (id && meId && id === meId) return copy.perfil.you;
  const short = id.replace(/[^0-9a-zA-Z]/g, "").slice(0, 4);
  return short ? copy.perfil.player(short) : copy.perfil.player("?");
}

export function describePlayers(ids: readonly string[], meId: string): string {
  const names = ids.map((id) => playerLabel(id, meId));
  if (names.length === 0) return copy.perfil.noPlayers;
  if (names.length === 1) return names[0]!;
  return `${names.slice(0, -1).join(", ")} e ${names[names.length - 1]}`;
}

export function describeWinner(winnerId: string | null | undefined, meId: string): string {
  if (!winnerId) return copy.perfil.noWinner;
  return winnerId === meId ? copy.perfil.youWon : copy.perfil.winner(playerLabel(winnerId, meId));
}

export type ProfileViewProps = {
  signedIn: boolean;
  name: string;
  userId: string;
  matches: readonly MyMatch[];
  loading: boolean;
  loadingMore: boolean;
  hasMore: boolean;
  error: string;
  notice: string;
  busyId: string;
  onLogin(): void;
  onWatch(matchId: string): void;
  onDownload(match: MyMatch): void;
  onLoadMore(): void;
  onRetry(): void;
};

function MatchRow({
  match,
  meId,
  busy,
  onWatch,
  onDownload,
}: {
  match: MyMatch;
  meId: string;
  busy: boolean;
  onWatch(matchId: string): void;
  onDownload(match: MyMatch): void;
}) {
  const date = formatMatchDate(match.ended_at);
  return (
    <li className={s.match}>
      <div className={s.matchInfo}>
        <div className={s.matchHead}>
          <time className={s.date} dateTime={match.ended_at}>
            {date}
          </time>
          <span className={s.result}>{describeWinner(match.winner_id, meId)}</span>
        </div>
        <p className={s.players}>{copy.perfil.players(describePlayers(match.player_ids, meId))}</p>
      </div>
      <div className={s.actions}>
        <Button
          variant="outline"
          className={s.action}
          aria-label={copy.perfil.watchLabel(date)}
          onClick={() => onWatch(match.match_id)}
        >
          {copy.perfil.watch}
        </Button>
        <Button
          variant="outline"
          className={s.action}
          aria-label={copy.perfil.downloadLabel(date)}
          disabled={busy}
          onClick={() => onDownload(match)}
        >
          {busy ? copy.perfil.downloading : copy.perfil.download}
        </Button>
      </div>
    </li>
  );
}

/** Parte visual do perfil, sem rede: facilita os testes de estado. */
export function ProfileView({
  signedIn,
  name,
  userId,
  matches,
  loading,
  loadingMore,
  hasMore,
  error,
  notice,
  busyId,
  onLogin,
  onWatch,
  onDownload,
  onLoadMore,
  onRetry,
}: ProfileViewProps) {
  const layout = useLayout();
  const display = name.trim() || copy.you;
  const initial = (display.charAt(0) || "?").toUpperCase();

  return (
    <main className={cx(s.screen, layout === "desktop" && s.desk)}>
      <div className={s.content}>
        <header className={s.header}>
          <a className={s.back} href="#" aria-label={copy.perfil.back}>
            <span aria-hidden="true">←</span>
          </a>
          <div className={s.heading}>
            <h1 className={s.title}>{copy.perfil.title}</h1>
            <p className={s.lead}>{copy.perfil.historyHint}</p>
          </div>
        </header>

        {!signedIn ? (
          <Moldura lugar="aviso" as="section" aria-label={copy.perfil.signedOutTitle} className={s.state}>
            <h2 className={s.stateTitle}>{copy.perfil.signedOutTitle}</h2>
            <p className={s.stateText}>{copy.perfil.signedOutText}</p>
            <Ticket size="l" className={s.stateAction} onClick={onLogin}>
              {copy.login}
            </Ticket>
          </Moldura>
        ) : (
          <>
            <section aria-label={copy.perfil.title} className={s.user}>
              <span aria-hidden="true" className={s.avatar}>
                {initial}
              </span>
              <div className={s.userText}>
                <h2 className={s.userName}>{display}</h2>
                <p className={s.userNote}>{copy.perfil.historyHint}</p>
              </div>
            </section>

            <section aria-label={copy.perfil.history} aria-busy={loading || loadingMore} className={s.section}>
              <div className={s.sectionHead}>
                <h2 className={cx("hw-rotulo", s.historyTitle)}>{copy.perfil.history}</h2>
              </div>

              {notice && (
                <p role="status" className={s.notice}>
                  {notice}
                </p>
              )}
              {error && matches.length > 0 && (
                <p role="alert" className={s.error}>
                  {error}
                </p>
              )}

              {loading ? (
                <p role="status" className={s.status}>
                  {copy.perfil.loading}
                </p>
              ) : matches.length === 0 ? (
                error ? (
                  <section role="alert" className={s.state}>
                    <h3 className={s.stateTitle}>{copy.perfil.errorTitle}</h3>
                    <p className={s.stateText}>{error}</p>
                    <Button variant="outline" onClick={onRetry}>
                      {copy.perfil.retry}
                    </Button>
                  </section>
                ) : (
                  <Moldura lugar="aviso" as="section" className={s.state}>
                    <h3 className={s.stateTitle}>{copy.perfil.emptyTitle}</h3>
                    <p className={s.stateText}>{copy.perfil.emptyText}</p>
                  </Moldura>
                )
              ) : (
                <>
                  <ul className={s.list}>
                    {matches.map((match) => (
                      <MatchRow
                        key={match.match_id}
                        match={match}
                        meId={userId}
                        busy={busyId === match.match_id}
                        onWatch={onWatch}
                        onDownload={onDownload}
                      />
                    ))}
                  </ul>
                  {hasMore && (
                    <Button
                      variant="outline"
                      className={s.more}
                      disabled={loadingMore}
                      onClick={onLoadMore}
                    >
                      {loadingMore ? copy.perfil.loadingMore : copy.perfil.loadMore}
                    </Button>
                  )}
                </>
              )}
            </section>
          </>
        )}
      </div>
    </main>
  );
}

/** Liga a tela ao login e ao histórico; 401 limpa a sessão e abre o login. */
export function Profile() {
  const token = useAuthStore((st) => st.jwtToken);
  const openLogin = useAuthStore((st) => st.openLogin);
  const name = useMemo(() => usernameFromToken(token), [token]);
  const userId = useMemo(() => playerIdFromToken(token), [token]);

  const [matches, setMatches] = useState<MyMatch[]>([]);
  const [limit, setLimit] = useState(MATCHES_PAGE_SIZE);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [busyId, setBusyId] = useState("");
  const [reload, setReload] = useState(0);
  const requestSeq = useRef(0);

  const load = useCallback(
    async (nextLimit: number, append: boolean) => {
      if (!token) return;
      const request = ++requestSeq.current;
      if (append) setLoadingMore(true);
      else setLoading(true);
      setError("");
      setNotice("");
      try {
        const items = await listMyMatches(token, nextLimit);
        if (request !== requestSeq.current) return;
        setMatches(items);
        setLimit(nextLimit);
        setHasMore(items.length >= nextLimit && nextLimit < MATCHES_MAX_LIMIT);
      } catch (failure) {
        if (request !== requestSeq.current || isAbortError(failure)) return;
        if (failure instanceof ReplayApiError && failure.status === 401) {
          useAuthStore.getState().clearSession();
          return;
        }
        setError(failure instanceof Error ? failure.message : copy.perfil.loadError);
      } finally {
        if (request === requestSeq.current) {
          setLoading(false);
          setLoadingMore(false);
        }
      }
    },
    [token],
  );

  useEffect(() => {
    if (!token) {
      requestSeq.current++;
      setMatches([]);
      setLoading(false);
      setLoadingMore(false);
      setHasMore(false);
      setError("");
      setNotice("");
      setBusyId("");
      return;
    }
    setBusyId("");
    void load(MATCHES_PAGE_SIZE, false);
    return () => {
      requestSeq.current++;
    };
  }, [token, reload, load]);

  const onLoadMore = () => {
    if (loading || loadingMore || !hasMore) return;
    void load(Math.min(limit + MATCHES_PAGE_SIZE, MATCHES_MAX_LIMIT), true);
  };

  const onDownload = async (match: MyMatch) => {
    if (!token || busyId) return;
    setBusyId(match.match_id);
    setError("");
    setNotice("");
    try {
      saveReplayFile(await downloadReplay(token, match.match_id));
      setNotice(copy.perfil.downloadStarted);
    } catch (failure) {
      if (failure instanceof ReplayApiError && failure.status === 401) {
        useAuthStore.getState().clearSession();
        return;
      }
      setError(failure instanceof Error ? failure.message : copy.perfil.downloadError);
    } finally {
      setBusyId("");
    }
  };

  const onWatch = (matchId: string) => {
    if (typeof window === "undefined") return;
    window.location.hash = `#replay?match=${encodeURIComponent(matchId)}`;
  };

  return (
    <>
      <ProfileView
        signedIn={!!token}
        name={name}
        userId={userId}
        matches={matches}
        loading={loading}
        loadingMore={loadingMore}
        hasMore={hasMore}
        error={error}
        notice={notice}
        busyId={busyId}
        onLogin={openLogin}
        onWatch={onWatch}
        onDownload={(match) => void onDownload(match)}
        onLoadMore={onLoadMore}
        onRetry={() => setReload((n) => n + 1)}
      />
      <LoginModal />
    </>
  );
}
