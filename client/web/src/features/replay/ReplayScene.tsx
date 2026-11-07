/**
 * Tela `#replay` (plano 05, T14): mesa só de leitura, alimentada pelo `.hwr`.
 *
 * O arquivo entra por dois caminhos: escolhido no disco (máx. 256 KB, conferido
 * antes de ler) ou baixado de `GET /replays/{matchId}` com o Bearer. O motor é o
 * Worker do T13; aqui só ligamos a sessão à mesa, à barra e ao vencedor.
 *
 * Tudo que vem do arquivo entra como texto do React (nunca HTML): nomes passam
 * por `sanitizeText` (32 caracteres, sem controle/bidi). O selo "replay não
 * oficial" aparece quando a assinatura não confere.
 */
import { useCallback, useEffect, useMemo, useRef, useState, type ChangeEvent } from "react";
import { useLayout } from "../../app/useLayout";
import { ENV } from "../../config/envs";
import { copy } from "../../game/copy";
import {
  playOrder,
  publicRoomSchema,
  type PublicPlayer,
  type PublicRoom,
} from "../../game/protocol";
import type { GameView } from "../../game/selectors";
import { selectFootlights } from "../../game/selectors";
import {
  MAX_REPLAY_BYTES,
  ReplayInvalidError,
  ReplayTimeoutError,
  openReplay,
  type ReplayFileLike,
  type ReplaySession,
} from "../../replay/loader";
import { useAuthStore } from "../../store/authStore";
import { Button } from "../../ui/Button/Button";
import { CardBack, CardDead, CardFace } from "../../ui/Card/Card";
import { Moldura } from "../../ui/Moldura/Moldura";
import { Sheet } from "../../ui/Sheet/Sheet";
import { cx } from "../../ui/cx";
import { Table } from "../table/Table";
import {
  PLAYBACK_SPEEDS,
  clampPosition,
  formatClock,
  nextPosition,
  positionAtTime,
  previousPosition,
  stepDelayMs,
  timeAtPosition,
  timelineFromDeltas,
  type PlaybackSpeed,
  type ReplayTimeline,
} from "./controls";
import { sanitizeText } from "./sanitize";
import { readReplayDeltas } from "./timeline";
import s from "./ReplayScene.module.css";

const fallbackPlayer = () => copy.replay.unknownPlayer;

/** `#replay?match=<id>`; o id vai cru para a URL e saneado para a tela. */
function matchFromHash(): string {
  if (typeof window === "undefined") return "";
  const hash = window.location.hash;
  if (!hash.startsWith("#replay") || !hash.includes("?")) return "";
  return new URLSearchParams(hash.slice(hash.indexOf("?") + 1)).get("match")?.trim() ?? "";
}

function roomFromState(raw: string): PublicRoom | null {
  let value: unknown;
  try {
    value = JSON.parse(raw);
  } catch {
    return null;
  }

  const parsed = publicRoomSchema.safeParse(value);
  if (!parsed.success) return null;

  const players = Object.fromEntries(
    Object.entries(parsed.data.players).map(([id, player]) => [
      id,
      { ...player, name: sanitizeText(player.name) || fallbackPlayer() },
    ]),
  );

  return { ...parsed.data, name: sanitizeText(parsed.data.name), players };
}

class FriendlyReplayError extends Error {}

function messageFor(error: unknown): string {
  if (error instanceof FriendlyReplayError) return error.message;
  if (error instanceof ReplayInvalidError) return copy.replay.invalid;
  if (error instanceof ReplayTimeoutError) return copy.replay.timeout;
  return copy.replay.loadError;
}

const isAbort = (error: unknown) => (error as { name?: string } | null)?.name === "AbortError";

type SheetState = { kind: "cards" } | { kind: "seat"; seat: string } | null;

// ---------------------------------------------------------------------------
// pedaços visuais
// ---------------------------------------------------------------------------

function ReplayHeader({
  source,
  official,
  showBadge,
  showCards,
  onOpen,
  onCards,
}: {
  source: string;
  official: boolean;
  showBadge: boolean;
  showCards: boolean;
  onOpen(): void;
  onCards(): void;
}) {
  const badge = official ? copy.replay.official : copy.replay.unofficial;
  const hint = official ? copy.replay.officialHint : copy.replay.unofficialHint;
  return (
    <header className={s.header}>
      <a className={s.back} href="#" aria-label={copy.replay.back}>
        <span aria-hidden="true">←</span>
      </a>
      <div className={s.heading}>
        <h1 className={s.title}>{copy.replay.title}</h1>
        {(source || showBadge) && (
          <p className={s.source}>
            {showBadge && (
              <span
                className={cx(s.badge, official ? s.official : s.unofficial)}
                role="status"
                aria-label={`${badge}. ${hint}`}
              >
                {badge}
              </span>
            )}
            {source && <span className={s.sourceText}>{source}</span>}
          </p>
        )}
      </div>
      <div className={s.headerActions}>
        <Button variant="outline" className={s.tool} onClick={onOpen}>
          {copy.replay.open}
        </Button>
        {showCards && (
          <Button variant="outline" className={s.tool} onClick={onCards}>
            {copy.replay.cards}
          </Button>
        )}
      </div>
    </header>
  );
}

function PlayerCards({ player }: { player: PublicPlayer }) {
  return (
    <div className={s.cardRow}>
      {player.cards.map((card, i) =>
        card.dead ? (
          <CardDead key={i} size="s" role={card.name} />
        ) : card.name ? (
          <CardFace
            key={i}
            size="s"
            role={card.name}
            protected={card.protected}
            bottom={card.protected ? copy.cardState.protected : undefined}
          />
        ) : (
          <CardBack key={i} size="s" />
        ),
      )}
    </div>
  );
}

/** Todas as cartas da partida, na ordem de jogo. Usada no trilho e na gaveta. */
function CardsList({ room }: { room: PublicRoom }) {
  return (
    <section className={s.cardsList} aria-label={copy.replay.cardsTitle}>
      <h2 className={cx("hw-rotulo", s.cardsTitle)}>{copy.replay.cardsTitle}</h2>
      {playOrder(room).map((player) => (
        <div key={player.id} className={s.cardGroup}>
          <div className={s.cardGroupHead}>
            <h3 className={s.cardPlayer}>{player.name}</h3>
            <span className={s.cardMeta}>
              {copy.coins(player.coins)}
              {!player.alive && ` · ${copy.replay.out}`}
            </span>
          </div>
          <PlayerCards player={player} />
        </div>
      ))}
    </section>
  );
}

function ReplaySeatSheet({ room, seatId, mode, onClose }: { room: PublicRoom; seatId: string; mode: "mobile" | "desktop"; onClose(): void }) {
  const player = room.players[seatId];
  if (!player) return null;
  return (
    <Sheet open onClose={onClose} label={copy.seatSheet(player.name)} mode={mode}>
      <div className={s.seatSheet}>
        <h2 className={s.seatName}>{player.name}</h2>
        <p className={s.seatMeta}>
          {copy.coins(player.coins)}
          {!player.alive && ` · ${copy.replay.out}`}
        </p>
        <PlayerCards player={player} />
      </div>
    </Sheet>
  );
}

function Controls({
  position,
  steps,
  timeline,
  playing,
  speed,
  onPlay,
  onPrevious,
  onNext,
  onSeek,
  onSpeed,
}: {
  position: number;
  steps: number;
  timeline: ReplayTimeline;
  playing: boolean;
  speed: PlaybackSpeed;
  onPlay(): void;
  onPrevious(): void;
  onNext(): void;
  onSeek(position: number): void;
  onSpeed(speed: PlaybackSpeed): void;
}) {
  const time = timeAtPosition(timeline, position);
  const atStart = position <= 1;
  const atEnd = steps > 0 && position >= steps;
  return (
    <div className={s.controls} role="group" aria-label={copy.replay.controls}>
      <div className={s.timeline}>
        <span className={cx("hw-num", s.clock)}>{formatClock(time)}</span>
        <input
          type="range"
          className={s.range}
          min={0}
          max={Math.max(1, Math.round(timeline.durationMs))}
          step={1}
          value={Math.min(Math.max(0, Math.round(time)), Math.max(1, Math.round(timeline.durationMs)))}
          disabled={steps === 0 || timeline.durationMs === 0}
          aria-label={copy.replay.progress}
          aria-valuetext={`${formatClock(time)} · ${copy.replay.position(position, steps)}`}
          onChange={(event) => onSeek(Math.max(1, positionAtTime(timeline, Number(event.target.value))))}
        />
        <span className={cx("hw-num", s.clock)}>{formatClock(timeline.durationMs)}</span>
      </div>
      <div className={s.buttons}>
        <Button variant="outline" className={s.control} onClick={onPrevious} disabled={atStart} aria-label={copy.replay.previous}>
          <span aria-hidden="true">◀</span>
        </Button>
        <Button variant="outline" className={s.control} onClick={onPlay}>
          {playing ? copy.replay.pause : copy.replay.play}
        </Button>
        <Button variant="outline" className={s.control} onClick={onNext} disabled={atEnd} aria-label={copy.replay.next}>
          <span aria-hidden="true">▶</span>
        </Button>
        <div className={s.speeds} role="group" aria-label={copy.replay.speed}>
          {PLAYBACK_SPEEDS.map((value) => (
            <Button
              key={value}
              variant="outline"
              className={cx(s.control, s.speed)}
              aria-pressed={value === speed}
              aria-label={copy.replay.speedLabel(value)}
              onClick={() => onSpeed(value)}
            >
              {value}×
            </Button>
          ))}
        </div>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// tela
// ---------------------------------------------------------------------------

export function ReplayScene() {
  const layout = useLayout();
  const token = useAuthStore((state) => state.jwtToken);
  const rawMatch = matchFromHash();
  const matchId = sanitizeText(rawMatch, 64);
  const fileRef = useRef<HTMLInputElement>(null);
  const sessionRef = useRef<ReplaySession | null>(null);
  const loadSeq = useRef(0);
  const requestSeq = useRef(0);

  const [session, setSession] = useState<ReplaySession | null>(null);
  const [room, setRoom] = useState<PublicRoom | null>(null);
  const [official, setOfficial] = useState(false);
  const [source, setSource] = useState("");
  const [loading, setLoading] = useState(() => !!matchFromHash());
  const [error, setError] = useState("");
  const [position, setPosition] = useState(1);
  const [steps, setSteps] = useState(0);
  const [timeline, setTimeline] = useState(() => timelineFromDeltas([]));
  const [playing, setPlaying] = useState(false);
  const [speed, setSpeed] = useState<PlaybackSpeed>(1);
  const [sheetState, setSheetState] = useState<SheetState>(null);

  const view = useMemo<GameView | null>(() => (room ? { me: null, room, log: [] } : null), [room]);
  const winner = useMemo(() => (room ? playOrder(room).find((player) => player.alive) : undefined), [room]);

  const sync = useCallback(async (current: ReplaySession) => {
    const [state, info] = await Promise.all([current.state(), current.info()]);
    const next = roomFromState(state);
    if (!next) throw new Error("estado do replay inválido");
    setRoom(next);
    setPosition(info.position);
    setSteps(info.steps);
  }, []);

  const load = useCallback(
    async (file: ReplayFileLike, label: string) => {
      const seq = ++loadSeq.current;
      requestSeq.current++;
      setLoading(true);
      setError("");
      setSource("");
      setPlaying(false);
      setSheetState(null);
      sessionRef.current?.close();
      sessionRef.current = null;
      setSession(null);
      setRoom(null);
      setSteps(0);
      setPosition(1);

      let opened: ReplaySession | null = null;
      try {
        const [next, deltas] = await Promise.all([
          openReplay(file),
          readReplayDeltas(file).catch(() => [] as number[]),
        ]);
        opened = next;
        if (seq !== loadSeq.current) {
          opened.close();
          return;
        }

        const info = await opened.info();
        const start = info.steps > 0 ? 1 : 0;
        if (start > 0) await opened.seek(start);
        if (seq !== loadSeq.current) {
          opened.close();
          return;
        }

        const state = await opened.state();
        const parsed = roomFromState(state);
        if (!parsed) throw new Error("estado do replay inválido");
        if (seq !== loadSeq.current) {
          opened.close();
          return;
        }

        sessionRef.current = opened;
        setSession(opened);
        setOfficial(opened.official);
        setSource(label);
        setSteps(info.steps);
        setPosition(start);
        setTimeline(timelineFromDeltas(deltas, info.steps));
        setRoom(parsed);
      } catch (failure) {
        opened?.close();
        if (seq !== loadSeq.current) return;
        setError(messageFor(failure));
      } finally {
        if (seq === loadSeq.current) setLoading(false);
      }
    },
    [],
  );

  const loadMatch = useCallback(
    async (signal?: AbortSignal) => {
      if (!matchId) return;
      const request = ++requestSeq.current;
      loadSeq.current++;
      sessionRef.current?.close();
      sessionRef.current = null;
      setSession(null);
      setRoom(null);
      setSteps(0);
      setPosition(1);
      setSource("");
      setPlaying(false);
      setSheetState(null);
      setError("");

      if (!token) {
        setLoading(false);
        setError(copy.replay.needLogin);
        return;
      }

      setLoading(true);
      try {
        const response = await fetch(
          `${ENV.VITE_BACKEND_URL.replace(/\/$/, "")}/replays/${encodeURIComponent(matchId)}`,
          {
            headers: { Authorization: `Bearer ${token}` },
            cache: "no-store",
          },
        );

        if (response.status === 401) throw new FriendlyReplayError(copy.replay.needLogin);
        if (response.status === 404) throw new FriendlyReplayError(copy.replay.notFound);
        if (!response.ok) throw new FriendlyReplayError(copy.replay.network);

        const blob = await response.blob();
        if (signal?.aborted || request !== requestSeq.current) return;
        if (blob.size > MAX_REPLAY_BYTES) throw new ReplayInvalidError();
        await load(blob, `${copy.replay.match} ${matchId.slice(0, 8)}`);
      } catch (failure) {
        if (signal?.aborted || isAbort(failure) || request !== requestSeq.current) return;
        setError(messageFor(failure));
      } finally {
        if (!signal?.aborted && request === requestSeq.current) setLoading(false);
      }
    },
    [load, matchId, token],
  );

  useEffect(() => {
    if (!rawMatch) return;
    const controller = new AbortController();
    void loadMatch(controller.signal);
    return () => controller.abort();
  }, [rawMatch, loadMatch]);

  useEffect(
    () => () => {
      loadSeq.current++;
      requestSeq.current++;
      sessionRef.current?.close();
      sessionRef.current = null;
    },
    [],
  );

  const openFile = useCallback(() => fileRef.current?.click(), []);

  const onFile = (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) return;
    requestSeq.current++;
    if (file.size > MAX_REPLAY_BYTES) {
      loadSeq.current++;
      sessionRef.current?.close();
      sessionRef.current = null;
      setSession(null);
      setRoom(null);
      setSource("");
      setError(copy.replay.tooBig);
      setLoading(false);
      return;
    }
    // Arquivo escolhido: o replay não pertence mais a um match do histórico.
    if (typeof window !== "undefined" && window.location.hash.includes("?")) {
      window.history.replaceState(null, "", "#replay");
    }
    void load(file, sanitizeText(file.name) || copy.replay.localFile);
  };

  const seekTo = useCallback(
    async (target: number) => {
      const current = sessionRef.current;
      if (!current) return;
      const next = clampPosition(target, steps);
      if (next === position) return;
      setPlaying(false);
      try {
        if (await current.seek(next)) await sync(current);
      } catch {
        setError(copy.replay.error);
        setPlaying(false);
      }
    },
    [position, steps, sync],
  );

  const togglePlay = useCallback(async () => {
    if (!sessionRef.current || steps === 0) return;
    if (playing) {
      setPlaying(false);
      return;
    }
    if (position >= steps) await seekTo(1);
    setPlaying(true);
  }, [playing, position, seekTo, steps]);

  useEffect(() => {
    if (!playing || !session || steps === 0) return;
    if (position >= steps) {
      setPlaying(false);
      return;
    }

    let cancelled = false;
    const timer = setTimeout(async () => {
      if (cancelled) return;
      try {
        const advanced = await session.step();
        if (cancelled) return;
        if (!advanced) {
          setPlaying(false);
          return;
        }
        await sync(session);
      } catch {
        if (cancelled) return;
        setPlaying(false);
        setError(copy.replay.error);
      }
    }, stepDelayMs(timeline, position, speed));

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [playing, position, session, speed, steps, sync, timeline]);

  const fileInput = (
    <input
      ref={fileRef}
      type="file"
      accept=".hwr,application/octet-stream"
      className={s.fileInput}
      aria-label={copy.replay.openLabel}
      tabIndex={-1}
      onChange={onFile}
    />
  );

  const header = (
    <ReplayHeader
      source={source}
      official={official}
      showBadge={!!session}
      showCards={layout === "mobile" && !!room}
      onOpen={openFile}
      onCards={() => setSheetState({ kind: "cards" })}
    />
  );

  if (!session || !room || !view) {
    return (
      <main className={s.emptyScreen}>
        {header}
        <div className={s.emptyBody}>
          {loading ? (
            <p role="status" className={s.status}>
              {copy.replay.loading}
            </p>
          ) : error ? (
            <div role="alert" className={s.errorBox}>
              <h2 className={s.emptyTitle}>{copy.replay.errorTitle}</h2>
              <p className={s.emptyText}>{error}</p>
              <div className={s.emptyActions}>
                {rawMatch && (
                  <Button variant="outline" onClick={() => void loadMatch()}>
                    {copy.replay.retry}
                  </Button>
                )}
                <Button variant="outline" onClick={openFile}>
                  {copy.replay.open}
                </Button>
              </div>
            </div>
          ) : (
            <Moldura lugar="aviso" className={s.emptyCard}>
              <h2 className={s.emptyTitle}>{copy.replay.emptyTitle}</h2>
              <p className={s.emptyText}>{copy.replay.emptyText}</p>
              <Button variant="outline" onClick={openFile}>
                {copy.replay.emptyFile}
              </Button>
            </Moldura>
          )}
        </div>
        {fileInput}
      </main>
    );
  }

  const footlights = selectFootlights(view);
  const sheet =
    sheetState?.kind === "cards" ? (
      <Sheet open onClose={() => setSheetState(null)} label={copy.replay.cardsTitle} mode={layout}>
        <CardsList room={room} />
      </Sheet>
    ) : sheetState?.kind === "seat" ? (
      <ReplaySeatSheet room={room} seatId={sheetState.seat} mode={layout} onClose={() => setSheetState(null)} />
    ) : undefined;

  return (
    <>
      <Table
        view={view}
        hideYou
        stageLive={false}
        dark={!room.gameOver && footlights.suddenDeath}
        onSeat={(seat) => setSheetState({ kind: "seat", seat })}
        slots={{
          header,
          rail: (
            <div className={s.railContent}>
              <CardsList room={room} />
            </div>
          ),
          sheet,
          bottomPanel: (
            <div className={s.bottom}>
              {room.gameOver && (
                <div className={s.winnerBanner} role="status" aria-live="polite">
                  <span className={cx("hw-rotulo", s.winnerKicker)}>{copy.replay.finish}</span>
                  <strong className={s.winnerName}>
                    {winner ? copy.replay.winner(winner.name) : copy.replay.noWinner}
                  </strong>
                </div>
              )}
              {error && (
                <p role="alert" className={s.inlineError}>
                  {error}
                </p>
              )}
              <Controls
                position={position}
                steps={steps}
                timeline={timeline}
                playing={playing}
                speed={speed}
                onPlay={() => void togglePlay()}
                onPrevious={() => void seekTo(Math.max(1, previousPosition(position, steps)))}
                onNext={() => void seekTo(nextPosition(position, steps))}
                onSeek={(next) => void seekTo(next)}
                onSpeed={setSpeed}
              />
            </div>
          ),
        }}
      />
      {fileInput}
    </>
  );
}
