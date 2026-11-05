import type { CSSProperties } from "react";
import type { Layout } from "../../app/useLayout";
import { useLayout } from "../../app/useLayout";
import { copy } from "../../game/copy";
import type { selectFinale } from "../../game/selectors";
import { selectFinale as selectFinaleVm } from "../../game/selectors";
import { createInput, useSessionStore, usernameFromToken } from "../../store/sessionStore";
import { useAuthStore } from "../../store/authStore";
import { useGameStore } from "../../store/gameStore";
import { Button } from "../../ui/Button/Button";
import { Ticket } from "../../ui/Ticket/Ticket";
import { Moldura } from "../../ui/Moldura/Moldura";
import { cx } from "../../ui/cx";
import s from "./Finale.module.css";

type Props = {
  vm: ReturnType<typeof selectFinale>;
  layout: Layout;
  onNewRoom(): void;
  onLobby(): void;
};

/** Fim da partida (tela 23): cartaz com o vencedor e a ordem de saída. */
export function Finale({ vm, layout, onNewRoom, onLobby }: Props) {
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
            “{vm.fala}” <span>Proso</span>
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
          {copy.newRoom}
        </Ticket>
        <Button variant="link" onClick={onLobby}>
          {copy.backToEntrance}
        </Button>
      </section>
    </main>
  );
}

const clearRoomUrl = () => {
  if (typeof window === "undefined") return;
  const url = new URL(window.location.href);
  url.searchParams.delete("sala");
  url.searchParams.delete("cena");
  url.searchParams.delete("fixture");
  window.history.replaceState(null, "", url);
};

/** Liga o cartaz ao estado: nova sala usa o login da sessão; voltar limpa tudo. */
export function FinaleScene() {
  const me = useGameStore((st) => st.me);
  const room = useGameStore((st) => st.room);
  const log = useGameStore((st) => st.log);
  const jwtToken = useAuthStore((st) => st.jwtToken);
  const create = useSessionStore((st) => st.create);
  const layout = useLayout();
  const vm = selectFinaleVm({ me, room, log });

  const onNewRoom = async () => {
    if (!jwtToken) {
      onLobby();
      return;
    }
    const code = await create(jwtToken, createInput(usernameFromToken(jwtToken)));
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

  return <Finale vm={vm} layout={layout} onNewRoom={onNewRoom} onLobby={onLobby} />;
}
