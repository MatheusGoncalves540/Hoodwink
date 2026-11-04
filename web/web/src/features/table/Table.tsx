/**
 * Mesa (telas 3, 7 e 13): cabeçalho, camarotes, cone, palco, ribalta e faixa "Você".
 * A view é pura (testável em Node); TableScene liga ao gameStore/uiStore.
 * Slots para as camadas de T21–T24: stageOverlay, bottomPanel, sheet, scenery, rail e screenOverlay.
 */
import { useRef, useState, type ReactNode } from "react";
import { useLayout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import type { RoleId } from "../../game/protocol";
import { ROLES } from "../../game/roles";
import { eventKey } from "../../game/log";
import { desktopSeats, mobileSeats } from "../../game/seats";
import type { GameView } from "../../game/selectors";
import { selectDecision, selectFootlights, selectSeats, selectStage, selectYou } from "../../game/selectors";
import { contest, contestPenalty, play, type Target } from "../../net/actions";
import { send } from "../../net/GameSocket";
import { useGameStore } from "../../store/gameStore";
import { useUiStore } from "../../store/uiStore";
import { cx } from "../../ui/cx";
import { ChatSheet } from "../sheets/ChatSheet";
import { DesktopRail } from "../sheets/DesktopRail";
import { RoomMenu } from "../sheets/RoomMenu";
import { SeatSheet } from "../sheets/SeatSheet";
import { YouSheet } from "../sheets/YouSheet";
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

export type TableSlots = {
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
  slots?: TableSlots;
};

export function Table({ view, unread = 0, onSeat, onMenu, onChat, targeting, hideYou, slots = {} }: Props) {
  const layout = useLayout();
  const desktop = layout === "desktop";
  const [ref, width] = useStageWidth(desktop ? 1020 : 390);
  const seats = selectSeats(view);
  const stage = selectStage(view);
  const footlights = selectFootlights(view);
  const you = selectYou(view);
  const layoutSeats = desktop ? desktopSeats(seats.length, width) : mobileSeats(seats.length, width);
  const activeIndex = seats.findIndex((seat) => seat.inScene);
  const activeBox = activeIndex >= 0 ? layoutSeats.boxes[activeIndex] : undefined;
  const room = view.room;

  return (
    <main className={cx(s.screen, desktop && s.desk)}>
      {slots.scenery && <div className={s.scenery}>{slots.scenery}</div>}

      <Header turn={room?.turn ?? 0} tax={room?.tax ?? 0} unread={unread} layout={layout} onMenu={onMenu} onChat={onChat} />

      <div ref={ref} className={s.tableArea}>
        <Spotlight box={activeBox} width={width} top={desktop ? 64 : 0} />
        <SeatRing seats={seats} layoutSeats={layoutSeats} layout={layout} onSeat={onSeat} targeting={targeting} />
        {!slots.stageOverlay && <Stage vm={stage} layout={layout} />}
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
  const view = { me, room, log };
  const decision = selectDecision(view);
  const declared = selectStage(view).declared;
  const evKey = room?.gameEvent ? eventKey(room.gameEvent) : "";
  const unread = Math.max(0, (room?.chat.length ?? 0) - chatSeen);

  const sendPlay = (role: RoleId, target?: Target) => {
    try {
      send(play(role, me ?? "", target));
    } catch (e) {
      console.error(e);
    }
    setTargeting(null);
  };

  const pickRole = (role: RoleId) => {
    const def = ROLES[role];
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
  };

  const closeSheet = () => {
    close();
    setTargeting(null);
  };

  const onSeat = (seatId: string) => {
    if (targeting) {
      const needsCard = targeting.role === "ASSASSIN" || targeting.role === "KAMIKAZE" || targeting.role === "CLAIRVOYANT";
      if (!needsCard) {
        sendPlay(targeting.role, { playerId: seatId });
        return;
      }
    }
    open({ seat: seatId });
  };

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

  const ownCardDisabled = (index: number) => {
    const card = me ? room?.players[me]?.cards.find((c) => c.index === index) : undefined;
    if (!card) return true;
    if (targeting?.role === "GRAVEDIGGER") return !card.dead;
    if (targeting?.role === "GUARDIAN") return card.dead || card.protected;
    return card.dead;
  };

  const sheet =
    openSheet === "you" ? (
      <YouSheet
        view={view}
        mode={layout}
        onClose={closeSheet}
        backLabel={targeting ? copy.backToActions : undefined}
        onPickCard={targeting && ROLES[targeting.role].target === "own" ? pickOwnCard : undefined}
        cardDisabled={targeting ? ownCardDisabled : undefined}
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
        onPickCard={targeting && ROLES[targeting.role].target === "other" ? pickSeatCard : undefined}
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
      />
    ) : undefined;

  const pickPanel =
    decision.kind === "pickCardToKill" || decision.kind === "waitPick" ? (
      <ContestPick view={view} decision={decision} layout={layout} onPick={(index) => send(contestPenalty(index))} />
    ) : undefined;

  const panel = reactPanel ?? pickPanel;
  const board = decision.kind === "myTurn" && !targeting ? <ActionBoard view={view} layout={layout} onPick={pickRole} /> : undefined;
  const targetBar = targeting ? <TargetBar role={targeting.role} onBack={() => setTargeting(null)} /> : undefined;

  return (
    <Table
      view={view}
      unread={unread}
      targeting={!!targeting}
      hideYou={layout === "desktop" && !!panel}
      onMenu={() => open("menu")}
      onChat={onChat}
      onSeat={onSeat}
      slots={{
        sheet,
        rail: panel ?? <DesktopRail view={view} chatInputRef={chatInputRef} />,
        stageOverlay: declared ? <DeclaredCard view={view} layout={layout} /> : undefined,
        screenOverlay: board,
        bottomPanel: targeting ? targetBar : layout === "mobile" ? panel : undefined,
      }}
    />
  );
}
