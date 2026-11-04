/**
 * Mesa (telas 3, 7 e 13): cabeçalho, camarotes, cone, palco, ribalta e faixa "Você".
 * A view é pura (testável em Node); TableScene liga ao gameStore/uiStore.
 * Slots para as camadas de T18–T24: stageOverlay, bottomPanel, sheet, scenery e rail.
 */
import type { ReactNode } from "react";
import { useLayout } from "../../app/useLayout";
import { desktopSeats, mobileSeats } from "../../game/seats";
import type { GameView } from "../../game/selectors";
import { selectFootlights, selectSeats, selectStage, selectYou } from "../../game/selectors";
import { useGameStore } from "../../store/gameStore";
import { useUiStore } from "../../store/uiStore";
import { cx } from "../../ui/cx";
import { SeatSheet } from "../sheets/SeatSheet";
import { YouSheet } from "../sheets/YouSheet";
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
  /** Ações/contestação no pé, no lugar da faixa "Você" (T21/T22). */
  bottomPanel?: ReactNode;
  /** Gaveta/menu/conversa abertos (T18/T19). */
  sheet?: ReactNode;
  /** Aviso/cortina de morte súbita atrás de tudo (T24). */
  scenery?: ReactNode;
  /** Conteúdo do trilho de 420 px no desktop (T20). */
  rail?: ReactNode;
};

type Props = {
  view: GameView;
  unread?: number;
  onSeat?(id: string): void;
  onMenu?(): void;
  onChat?(): void;
  slots?: TableSlots;
};

export function Table({ view, unread = 0, onSeat, onMenu, onChat, slots = {} }: Props) {
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
        <SeatRing seats={seats} layoutSeats={layoutSeats} layout={layout} onSeat={onSeat} />
        <Stage vm={stage} layout={layout} />
        <Footlights vm={footlights} layout={layout} />
        {slots.bottomPanel && <div className={s.bottomPanel}>{slots.bottomPanel}</div>}
      </div>

      {!slots.bottomPanel && <YouBar vm={you} suddenDeath={footlights.suddenDeath} layout={layout} />}
      {desktop && <aside className={s.rail}>{slots.rail}</aside>}
      {slots.stageOverlay && <div className={s.stageOverlay}>{slots.stageOverlay}</div>}
      {slots.sheet}
    </main>
  );
}

/** Liga a mesa ao estado real: gavetas pelo uiStore, conversa com o contador de novas. */
export function TableScene() {
  const me = useGameStore((st) => st.me);
  const room = useGameStore((st) => st.room);
  const log = useGameStore((st) => st.log);
  const chatSeen = useUiStore((st) => st.chatSeen);
  const open = useUiStore((st) => st.open);
  const openSheet = useUiStore((st) => st.openSheet);
  const close = useUiStore((st) => st.close);
  const layout = useLayout();
  const view = { me, room, log };
  const unread = Math.max(0, (room?.chat.length ?? 0) - chatSeen);

  const sheet =
    openSheet === "you" ? (
      <YouSheet view={view} mode={layout} onClose={close} />
    ) : openSheet && typeof openSheet === "object" ? (
      <SeatSheet view={view} seatId={openSheet.seat} mode={layout} onClose={close} />
    ) : undefined;

  return (
    <Table
      view={view}
      unread={unread}
      onMenu={() => open("menu")}
      onChat={() => open("chat")}
      onSeat={(seat) => open({ seat })}
      slots={{ sheet }}
    />
  );
}
