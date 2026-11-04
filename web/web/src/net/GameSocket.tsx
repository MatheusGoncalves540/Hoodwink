/**
 * Conexão com o gameServer: ws(s)://<VITE_GAMESERVER_URL>/game?Ticket=...
 * Cada mensagem é um PublicRoomForUpdates sem envelope → parseRoom → gameStore.applyRoom.
 * O ticket é um JWT reaproveitável até expirar, então a reconexão usa o mesmo.
 */
import { useEffect } from "react";
import { ENV } from "../config/envs";
import { parseRoom } from "../game/protocol";
import { useGameStore } from "../store/gameStore";
import type { ClientMessage } from "./actions";

// ponytail: uma conexão de jogo por aba, guardada no módulo; send() funciona de qualquer componente.
let socket: WebSocket | null = null;

export function send(msg: ClientMessage): boolean {
  if (socket?.readyState !== WebSocket.OPEN) return false;
  socket.send(JSON.stringify(msg));
  return true;
}

export const gameUrl = (ticket: string) =>
  `${ENV.VITE_GAMESERVER_URL.replace(/^http/, "ws").replace(/\/$/, "")}/game?Ticket=${encodeURIComponent(ticket)}`;

export function GameSocket({ ticket }: { ticket: string }) {
  useEffect(() => {
    const { setConnection, applyRoom } = useGameStore.getState();
    let alive = true;
    let tries = 0;
    let timer: ReturnType<typeof setTimeout> | undefined;

    const connect = () => {
      setConnection("connecting");
      const ws = new WebSocket(gameUrl(ticket));
      socket = ws;
      ws.onopen = () => {
        tries = 0;
        setConnection("open");
      };
      ws.onmessage = (e) => {
        let raw: unknown;
        try {
          raw = JSON.parse(e.data);
        } catch {
          console.warn("gameServer mandou algo que não é JSON:", e.data);
          return;
        }
        const room = parseRoom(raw);
        if (room) applyRoom(room);
      };
      ws.onclose = () => {
        if (!alive || socket !== ws) return;
        setConnection("closed");
        // ponytail: backoff 0,5 s → 10 s sem limite; ticket vencido fica tentando até a T25 pedir outro.
        timer = setTimeout(connect, Math.min(10_000, 500 * 2 ** tries++));
      };
    };

    connect();
    return () => {
      alive = false;
      clearTimeout(timer);
      socket?.close();
      socket = null;
      setConnection("idle");
    };
  }, [ticket]);

  return null;
}
