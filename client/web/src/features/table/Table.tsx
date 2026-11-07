/**
 * Mesa (telas 3, 7 e 13): cabeçalho, camarotes, cone, palco, ribalta e faixa "Você".
 * A view é pura (testável em Node); TableScene liga ao gameStore/uiStore.
 * Slots para as camadas de T21–T24: stageOverlay, bottomPanel, sheet, scenery, rail e screenOverlay.
 */
import { lazy, Suspense, useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { useLayout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import { actionDef, type ActionId } from "../../game/roles";
import type { PublicCard } from "../../game/protocol";
import { eventKey } from "../../game/log";
import { CLASSIC_RULES } from "../../game/rules.classic";
import { desktopSeats, mobileSeats } from "../../game/seats";
import type { GameView } from "../../game/selectors";
import { kamikazeAlly, selectDecision, selectFootlights, selectSeats, selectStage, selectYou } from "../../game/selectors";
import { contest, contestPenalty, play, type Target } from "../../net/actions";
import { send } from "../../net/GameSocket";
import { useGameStore } from "../../store/gameStore";
import { useSessionStore } from "../../store/sessionStore";
import { useUiStore } from "../../store/uiStore";
import { cx } from "../../ui/cx";
import { ChatSheet } from "../sheets/ChatSheet";
import { DesktopRail } from "../sheets/DesktopRail";
import { RoomMenu } from "../sheets/RoomMenu";
import { SeatSheet } from "../sheets/SeatSheet";
import { YouSheet } from "../sheets/YouSheet";
import { OutOfScene } from "../endgame/OutOfScene";
import { ActionBoard } from "../turn/ActionBoard";
import { ContestPick } from "../turn/ContestPick";
import { DeclaredCard } from "../turn/DeclaredCard";
import { ReactionPanel } from "../turn/ReactionPanel";
import { TargetBar } from "../turn/TargetBar";
import { Footlights } from "./Footlights";
import { Header } from "./Header";
import { SeatRing } from "./SeatRing";
import { Spotlight } from "./Spotlight";
import { Stage } from "./Stage";
import { YouBar } from "./YouBar";
import { useStageWidth } from "./useStageWidth";
import s from "./Table.module.css";

const SuddenDeathStage = lazy(() => import("../endgame/SuddenDeathStage").then((m) => ({ default: m.SuddenDeathStage })));
const SuddenDeathAlert = lazy(() => import("../endgame/SuddenDeathAlert").then((m) => ({ default: m.SuddenDeathAlert })));

export type TableSlots = {
  /** Cabeçalho próprio (replay): substitui menu/rodada/conversa. */
  header?: ReactNode;
  /** Painel de reação sobre o palco (T22). */
  stageOverlay?: ReactNode;
  /** Ações/contestação no pé, no lugar da faixa "Você" (T21/T22/T23). */
  bottomPanel?: ReactNode;
  /** Gaveta/menu/conversa abertos (T18/T19). */
  sheet?: ReactNode;
  /** Aviso/cortina de morte súbita atrás de tudo (T24). */
  scenery?: ReactNode;
  /** Conteúdo do trilho de 420 px no desktop (T20). */
  rail?: ReactNode;
  /** Cobre a tela inteira, inclusive cabeçalho e trilho (ações, aviso de morte súbita). */
  screenOverlay?: ReactNode;
};

type Props = {
  view: GameView;
  unread?: number;
  onSeat?(id: string): void;
  onMenu?(): void;
  onChat?(): void;
  targeting?: boolean;
  /** Reação/contestação no trilho: a faixa Você some no desktop (MainDesk). */
  hideYou?: boolean;
  /** Morte súbita: a mesa inteira fica escura. */
  dark?: boolean;
  /** Palco anuncia a troca de jogador? No replay, não: o passo a passo já tem rótulos. */
  stageLive?: boolean;
  slots?: TableSlots;
};

export function Table({ view, unread = 0, onSeat, onMenu, onChat, targeting, hideYou, dark, stageLive, slots = {} }: Props) {
  const layout = useLayout();
  const desktop = layout === "desktop";
  const [ref, width] = useStageWidth(desktop ? 1020 : 390);
  const seats = useMemo(() => selectSeats(view), [view]);
  const stage = useMemo(() => selectStage(view), [view]);
  const footlights = useMemo(() => selectFootlights(view), [view]);
  const you = useMemo(() => selectYou(view), [view]);
  const layoutSeats = useMemo(
    () => (desktop ? desktopSeats(seats.length, width) : mobileSeats(seats.length, width)),
    [desktop, seats.length, width],
  );
  const activeIndex = useMemo(() => seats.findIndex((seat) => seat.inScene), [seats]);
  const activeBox = activeIndex >= 0 ? layoutSeats.boxes[activeIndex] : undefined;
  const decision = useMemo(() => selectDecision(view), [view]);
  const room = view.room;

  return (
    // data-deixa: na sua vez o ambiente para; só o timer e o botão de ação tremem (motion.css)
    <main data-theme={dark ? "dark" : undefined} data-deixa={decision.kind === "myTurn" || undefined} className={cx(s.screen, desktop && s.desk)}>
      {slots.scenery && <div className={s.scenery}>{slots.scenery}</div>}

      {slots.header ?? <Header turn={room?.turn ?? 0} tax={room?.tax ?? 0} unread={unread} layout={layout} sudden={footlights.suddenDeath} onMenu={onMenu} onChat={onChat} />}

      <div ref={ref} className={s.tableArea}>
        <Spotlight box={activeBox} width={width} top={desktop ? 64 : 0} />
        <SeatRing seats={seats} layoutSeats={layoutSeats} layout={layout} onSeat={onSeat} targeting={targeting} />
        {!slots.stageOverlay && <Stage vm={stage} layout={layout} live={stageLive} />}
        <Footlights vm={footlights} layout={layout} />
        {slots.bottomPanel && <div className={s.bottomPanel}>{slots.bottomPanel}</div>}
      </div>

      {!slots.bottomPanel && !hideYou && <YouBar vm={you} suddenDeath={footlights.suddenDeath} layout={layout} />}
      {desktop && <aside className={s.rail}>{slots.rail}</aside>}
      {slots.stageOverlay && <div className={s.stageOverlay}>{slots.stageOverlay}</div>}
      {slots.sheet}
      {slots.screenOverlay && <div className={s.screenOverlay}>{slots.screenOverlay}</div>}
    </main>
  );
}

/** Fluxo de alvo: a carta sua pode ser escolhida por este papel? (regras v2, 02 §3) */
export function ownCardDisabled(role: ActionId | undefined, card: PublicCard): boolean {
  if (role === "GRAVEDIGGER") return !card.dead;
  if (role === "GUARDIAN") return card.dead || card.protected;
  // O Crupiê descarta uma carta viva ou morta e puxa uma nova do baralho.
  if (role === "CROUPIER") return false;
  return card.dead;
}

/** Liga a mesa ao estado real: gavetas, trilho, ações, reação e contestação. */
export function TableScene() {
  const me = useGameStore((st) => st.me);
  const room = useGameStore((st) => st.room);
  const log = useGameStore((st) => st.log);
  const chatSeen = useUiStore((st) => st.chatSeen);
  const open = useUiStore((st) => st.open);
  const openSheet = useUiStore((st) => st.openSheet);
  const close = useUiStore((st) => st.close);
  const targeting = useUiStore((st) => st.targeting);
  const setTargeting = useUiStore((st) => st.setTargeting);
  const layout = useLayout();
  const chatInputRef = useRef<HTMLInputElement>(null);
  const [passed, setPassed] = useState<string | null>(null);
  const view = useMemo(() => ({ me, room, log }), [me, room, log]);
  const decision = useMemo(() => selectDecision(view), [view]);
  const footlights = useMemo(() => selectFootlights(view), [view]);
  const stage = useMemo(() => selectStage(view), [view]);
  const [showAlert, setShowAlert] = useState(false);
  const evKey = room?.gameEvent ? eventKey(room.gameEvent) : "";
  const unread = Math.max(0, (room?.chat.length ?? 0) - chatSeen);

  // evento novo (inclusive o fim da deixa) fecha o fluxo de alvo: a jogada antiga não vale mais
  useEffect(() => {
    if (useUiStore.getState().targeting) setTargeting(null);
  }, [evKey, setTargeting]);

  useEffect(() => {
    if (!footlights.suddenDeath || !room) return;
    const ui = useUiStore.getState();
    if (ui.suddenDeathShown === room.id) return;
    ui.setSuddenDeathShown(room.id);
    setShowAlert(true);
    const timer = setTimeout(() => setShowAlert(false), 4000);
    return () => clearTimeout(timer);
  }, [footlights.suddenDeath, room?.id]);

  const sendPlay = useCallback(
    (role: ActionId, target?: Target) => {
      try {
        // Kamikaze: o servidor exige saber qual Kamikaze seu morre (targetAllyCardIndex)
        const ally = role === "KAMIKAZE" ? (kamikazeAlly(view) ?? undefined) : undefined;
        send(play(role, me ?? "", { showToAll: targeting?.showToAll, allyCardIndex: ally, ...target }));
      } catch (e) {
        console.error(e);
      }
      setTargeting(null);
    },
    [view, me, targeting, setTargeting],
  );

  const pickRole = useCallback(
    (role: ActionId) => {
      const def = actionDef(role);
      if (def.target === "own") {
        setTargeting({ role });
        open("you");
        return;
      }
      if (def.target === "other") {
        setTargeting({ role });
        return;
      }
      sendPlay(role);
    },
    [open, sendPlay, setTargeting],
  );

  const closeSheet = () => {
    close();
    setTargeting(null);
  };

  const onSeat = useCallback(
    (seatId: string) => {
      if (targeting) {
        const needsCard = ["ASSASSIN", "KAMIKAZE", "CLAIRVOYANT", "KILL_CARD"].includes(targeting.role);
        if (!needsCard) {
          sendPlay(targeting.role, { playerId: seatId });
          return;
        }
      }
      open({ seat: seatId });
    },
    [targeting, sendPlay, open],
  );

  const pickSeatCard = (index: number) => {
    if (!targeting || typeof openSheet !== "object" || !openSheet) return;
    sendPlay(targeting.role, { playerId: openSheet.seat, cardIndex: index });
    close();
  };

  const pickOwnCard = (index: number) => {
    if (!targeting) return;
    sendPlay(targeting.role, { cardIndex: index });
    close();
  };

  const ownCardDisabledFor = (index: number) => {
    const card = me ? room?.players[me]?.cards.find((c) => c.index === index) : undefined;
    return !card || ownCardDisabled(targeting?.role, card);
  };

  const sheet =
    openSheet === "you" ? (
      <YouSheet
        view={view}
        mode={layout}
        onClose={closeSheet}
        backLabel={targeting ? copy.backToActions : undefined}
        onPickCard={targeting && actionDef(targeting.role).target === "own" ? pickOwnCard : undefined}
        cardDisabled={targeting ? ownCardDisabledFor : undefined}
      />
    ) : openSheet === "menu" ? (
      <RoomMenu view={view} mode={layout} onClose={closeSheet} />
    ) : openSheet === "chat" ? (
      <ChatSheet view={view} mode={layout} onClose={closeSheet} />
    ) : openSheet && typeof openSheet === "object" ? (
      <SeatSheet
        view={view}
        seatId={openSheet.seat}
        mode={layout}
        onClose={closeSheet}
        backLabel={targeting ? copy.backToActions : undefined}
        onPickCard={targeting && actionDef(targeting.role).target === "other" ? pickSeatCard : undefined}
      />
    ) : undefined;

  const onChat = () => {
    if (layout === "desktop") {
      if (openSheet) close();
      requestAnimationFrame(() => {
        chatInputRef.current?.scrollIntoView({ block: "center" });
        chatInputRef.current?.focus();
      });
    } else {
      open("chat");
    }
  };

  const leaveRoom = () => {
    useGameStore.getState().reset();
    useSessionStore.getState().reset();
    if (typeof window === "undefined") return;
    const url = new URL(window.location.href);
    url.searchParams.delete("sala");
    url.searchParams.delete("cena");
    url.searchParams.delete("fixture");
    window.history.replaceState(null, "", url);
  };

  const onBlock = () => {
    if (decision.kind !== "react" || !decision.block) return;
    if (decision.block.role === "GUARDIAN") {
      const index = Number(((room?.gameEvent?.payload ?? {}) as Record<string, unknown>).targetCardIndex ?? 0);
      sendPlay("GUARDIAN", { cardIndex: index });
    } else {
      sendPlay("SELFISH");
    }
  };

  const reactPanel =
    decision.kind === "react" && passed !== evKey ? (
      <ReactionPanel
        view={view}
        decision={decision}
        layout={layout}
        onContest={() => send(contest())}
        onBlock={onBlock}
        onPass={() => setPassed(evKey)}
        onKamikaze={() => setTargeting({ role: "KAMIKAZE" })}
      />
    ) : undefined;

  const pickPanel =
    decision.kind === "pickCardToKill" || decision.kind === "waitPick" ? (
      <ContestPick view={view} decision={decision} layout={layout} onPick={(index) => send(contestPenalty(index))} />
    ) : undefined;

  const outPanel = decision.kind === "outOfScene" ? <OutOfScene view={view} layout={layout} onChat={onChat} onLeave={leaveRoom} /> : undefined;
  const panel = reactPanel ?? pickPanel ?? outPanel;
  const board = decision.kind === "myTurn" && !targeting ? <ActionBoard view={view} layout={layout} onPick={pickRole} /> : undefined;
  const scenery = footlights.suddenDeath ? <Suspense fallback={null}><SuddenDeathStage layout={layout} /></Suspense> : undefined;
  const screenOverlay = board || showAlert ? (
    <>
      {board}
      {showAlert && (
        <Suspense fallback={null}>
          <SuddenDeathAlert layout={layout} seconds={CLASSIC_RULES.waitingAction.suddenDeath} />
        </Suspense>
      )}
    </>
  ) : undefined;
  const targetBar = targeting ? (
    <TargetBar
      role={targeting.role}
      onBack={() => setTargeting(null)}
      onSwapBoth={targeting.role === "SWAP_CARDS" ? () => sendPlay("SWAP_CARDS", { cardIndex: 0, both: true }) : undefined}
      showToAll={targeting.role === "CLAIRVOYANT" ? !!targeting.showToAll : undefined}
      onShowToAll={(v) => setTargeting({ ...targeting, showToAll: v })}
    />
  ) : undefined;

  return (
    <Table
      view={view}
      unread={unread}
      targeting={!!targeting}
      hideYou={layout === "desktop" && !!panel}
      dark={footlights.suddenDeath}
      onMenu={() => open("menu")}
      onChat={onChat}
      onSeat={onSeat}
      slots={{
        sheet,
        rail: panel ?? <DesktopRail view={view} chatInputRef={chatInputRef} />,
        stageOverlay: stage.declared || stage.revealed ? <DeclaredCard view={view} layout={layout} /> : undefined,
        scenery,
        screenOverlay,
        bottomPanel: targeting ? targetBar : layout === "mobile" ? panel : undefined,
      }}
    />
  );
}
