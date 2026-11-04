/**
 * Casca do app: tema e movimento na raiz do documento, e a cena da vez.
 * Seletor de desenvolvimento: ?cena=lobby|room|table|finale|amostras&fixture=Mesa6 (prompt 01).
 * A T25 troca o seletor pelo estado real (login → sala → ticket → GameSocket).
 */
import { useEffect, type ComponentType } from "react";
import LoginModal from "../componentsTSX/loginModal";
import { FIXTURE_NAMES, type FixtureName } from "../game/fixtures";
import { selectScene } from "../game/selectors";
import { Lobby } from "../features/lobby/Lobby";
import { RoomScene } from "../features/room/WaitingRoom";
import { useAuthStore } from "../store/authStore";
import { useGameStore } from "../store/gameStore";
import { Amostras } from "./Amostras";
import { useMotionPref } from "./useMotionPref";
import { useTheme } from "./useTheme";
import "./app.css";

export type Scene = ReturnType<typeof selectScene>;

/** Cada tela entra aqui quando fica pronta (T14 lobby, T15 room, T16 table, T23 finale). */
const SCENES: Partial<Record<Scene, ComponentType>> = { lobby: Lobby, room: RoomScene };

const param = (name: string) => (typeof window === "undefined" ? null : new URLSearchParams(window.location.search).get(name));
const isFixture = (s: string | null): s is FixtureName => FIXTURE_NAMES.includes(s as FixtureName);
const isScene = (s: string | null): s is Scene => s === "lobby" || s === "room" || s === "table" || s === "finale";

export function App() {
  const { theme } = useTheme();
  const { reduced } = useMotionPref();
  const me = useGameStore((s) => s.me);
  const room = useGameStore((s) => s.room);
  const log = useGameStore((s) => s.log);
  const cena = param("cena");
  const fixture = param("fixture");

  useEffect(() => {
    if (isFixture(fixture)) useGameStore.getState().loadFixture(fixture);
  }, [fixture]);

  // Na raiz do documento (não num div) para as gavetas do Radix, que vão para o body, herdarem o tema.
  useEffect(() => {
    const root = document.documentElement;
    root.dataset.theme = theme;
    if (reduced) root.dataset.motion = "reduced";
    else delete root.dataset.motion;
  }, [theme, reduced]);

  if (cena === "amostras") return <Amostras />;
  const Scene = SCENES[isScene(cena) ? cena : selectScene({ me, room, log })];
  return Scene ? <Scene /> : <LegacyLogin />;
}

// ponytail: login legado até a entrada (T14) existir; a T25 tira daqui.
function LegacyLogin() {
  const openLogin = useAuthStore((s) => s.openLogin);
  return (
    <div>
      <button onClick={openLogin}>Login</button>
      <LoginModal />
    </div>
  );
}
