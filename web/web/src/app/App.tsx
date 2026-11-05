/**
 * Casca do app: tema e movimento na raiz do documento, e a cena que o estado pede.
 * Estado real (T25): ?sala= → ticket (HTTP) → GameSocket → selectScene.
 * Em desenvolvimento, ?cena=&fixture= abre uma prancha direto (T2); em produção não existe.
 */
import { useEffect, type ComponentType } from "react";
import { FIXTURE_NAMES, type FixtureName } from "../game/fixtures";
import { selectScene } from "../game/selectors";
import { FinaleScene } from "../features/endgame/Finale";
import { Lobby } from "../features/lobby/Lobby";
import { RoomScene } from "../features/room/WaitingRoom";
import { TableScene } from "../features/table/Table";
import { GameSocket } from "../net/GameSocket";
import { useAuthStore } from "../store/authStore";
import { useGameStore } from "../store/gameStore";
import { roomCodeFromSearch, useSessionStore } from "../store/sessionStore";
import { Amostras } from "./Amostras";
import { useMotionPref } from "./useMotionPref";
import { useTheme } from "./useTheme";
import "./app.css";

export type Scene = ReturnType<typeof selectScene>;

/** Cada tela entra aqui quando fica pronta (T14 lobby, T15 room, T16 table, T23 finale). */
const SCENES: Record<Scene, ComponentType> = { lobby: Lobby, room: RoomScene, table: TableScene, finale: FinaleScene };

const param = (name: string) => (typeof window === "undefined" ? null : new URLSearchParams(window.location.search).get(name));
const isFixture = (s: string | null): s is FixtureName => FIXTURE_NAMES.includes(s as FixtureName);
const isScene = (s: string | null): s is Scene => s === "lobby" || s === "room" || s === "table" || s === "finale";

/**
 * Fila de apresentação (09-integracao.md): o evento da frente sai quando a janela dele fecha.
 * Os painéis leem o gameEvent da sala, que o servidor manda um por vez; a fila evita que um
 * evento rápido apague o anterior e impede a fila de crescer durante a partida.
 */
function usePresentation() {
  const head = useGameStore((s) => s.presentation[0]);
  useEffect(() => {
    if (!head) return;
    const timer = setTimeout(() => useGameStore.getState().shift(), Math.max(0, head.expiresAt - Date.now()));
    return () => clearTimeout(timer);
  }, [head]);
}

export function App() {
  const { theme } = useTheme();
  const motion = useMotionPref();
  const me = useGameStore((s) => s.me);
  const room = useGameStore((s) => s.room);
  const log = useGameStore((s) => s.log);
  const ticket = useSessionStore((s) => s.ticket);
  const jwtToken = useAuthStore((s) => s.jwtToken);
  const join = useSessionStore((s) => s.join);

  // Seletor de prancha: só em desenvolvimento, e nunca junto de uma sala de verdade.
  const devScene = import.meta.env.DEV && !ticket ? param("cena") : null;
  const devFixture = import.meta.env.DEV && !ticket ? param("fixture") : null;

  usePresentation();

  useEffect(() => {
    if (isFixture(devFixture)) useGameStore.getState().loadFixture(devFixture);
  }, [devFixture]);

  // Convite (?sala=): pede o ticket assim que houver login; sem login, o código fica no cartaz.
  useEffect(() => {
    if (ticket || !jwtToken || devScene) return;
    const code = typeof window === "undefined" ? null : roomCodeFromSearch(window.location.search);
    if (code) void join(jwtToken, code);
  }, [ticket, jwtToken, join, devScene]);

  // Na raiz do documento (não num div) para as gavetas do Radix, que vão para o body, herdarem o tema.
  // Único lugar que escreve movimento na raiz: data-motion (Animações) e data-idle (espera; também com a aba escondida).
  useEffect(() => {
    const root = document.documentElement;
    root.dataset.theme = theme;
    if (!motion.all) root.dataset.motion = "reduced";
    else delete root.dataset.motion;
    const idle = () => {
      if (!motion.idle || document.hidden) root.dataset.idle = "off";
      else delete root.dataset.idle;
    };
    idle();
    document.addEventListener("visibilitychange", idle);
    return () => document.removeEventListener("visibilitychange", idle);
  }, [theme, motion.all, motion.idle]);

  if (devScene === "amostras") return <Amostras />;
  const Scene = SCENES[isScene(devScene) ? devScene : selectScene({ me, room, log })];
  return (
    <>
      {ticket && <GameSocket ticket={ticket} />}
      <Scene />
    </>
  );
}
