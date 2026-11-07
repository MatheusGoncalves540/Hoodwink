/**
 * Sala de espera (telas 2 e 17): convite, elenco e "Abrir as cortinas".
 * WaitingRoom é a view pura (testável em Node); RoomScene liga ao gameStore/socket.
 */
import { useLayout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import type { GameView } from "../../game/selectors";
import { selectWaitingRoom } from "../../game/selectors";
import { startGame } from "../../net/actions";
import { send } from "../../net/GameSocket";
import { useGameStore } from "../../store/gameStore";
import { useSessionStore } from "../../store/sessionStore";
import { Ticket } from "../../ui/Ticket/Ticket";
import { cx } from "../../ui/cx";
import { Cast } from "./Cast";
import { Invite } from "./Invite";
import s from "./WaitingRoom.module.css";
import { SupportDev } from "../support/SupportDev";

type Props = {
  view: GameView;
  onLeave?(): void;
  onStart?(): void;
};

export function WaitingRoom({ view, onLeave, onStart }: Props) {
  const layout = useLayout();
  const vm = selectWaitingRoom(view);
  const canStart = vm.isHost && vm.count >= 2;

  return (
    <main className={cx(s.screen, layout === "desktop" && s.desk)}>
      <header className={s.header}>
        <button type="button" className={s.back} aria-label={copy.leaveScene} onClick={onLeave}>
          <svg
            aria-hidden="true"
            width="22"
            height="22"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M15 5l-7 7 7 7" />
          </svg>
        </button>
        <span className={s.backstage}>{copy.backstage}</span>
      </header>

      <Invite code={vm.code} layout={layout} className={s.inviteSlot} />
      <Cast vm={vm} layout={layout} className={s.castSlot} />

      <section aria-label={copy.startRegion} className={s.start}>
        <Ticket
          size="l"
          disabled={!canStart}
          reason={!canStart ? copy.startHint : undefined}
          onClick={onStart}
          className={cx("m-rise", s.startButton)}
        >
          {copy.openCurtains}
        </Ticket>
        <span className={s.hint}>{copy.startHint}</span>
        <SupportDev layout={layout} />
      </section>
    </main>
  );
}

/** Liga a view ao estado real: sair limpa sessão/sala; abrir as cortinas manda COMMAND_START_GAME. */
export function RoomScene() {
  const me = useGameStore((st) => st.me);
  const room = useGameStore((st) => st.room);
  const log = useGameStore((st) => st.log);

  const leave = () => {
    useGameStore.getState().reset();
    useSessionStore.getState().reset();
    if (typeof window === "undefined") return;
    const url = new URL(window.location.href);
    url.searchParams.delete("sala");
    url.searchParams.delete("cena");
    url.searchParams.delete("fixture");
    window.history.replaceState(null, "", url);
  };

  const start = () => {
    send(startGame());
  };

  return <WaitingRoom view={{ me, room, log }} onLeave={leave} onStart={start} />;
}
