/** Cria uma sala de UI, mantém o jogador 2 conectado e inicia quando o jogador 1 (navegador) entrar. */
import { writeFileSync } from "node:fs";
import { connect, jwt, newRoom, ticketJwt } from "./lib.mjs";
const run = Date.now().toString(36);
const created = await newRoom(jwt(`ui-ana-${run}`, "Ana UI"), { roomName: "UI caos" });
const roomId = created.json.RoomId;
const p1 = { id: `ui-ana-${run}`, name: "Ana UI" };
const p2 = { id: `ui-bruno-${run}`, name: "Bruno UI" };
const socket = connect(ticketJwt(p2.id, p2.name, roomId));
await socket.waitOpen;
writeFileSync("tests/caos/ui-room.json", JSON.stringify({ roomId, p1: { ...p1, ticket: ticketJwt(p1.id, p1.name, roomId) }, p2 }, null, 2));
console.log("ui-room", roomId, p1.id, p2.id);
try {
  await socket.waitFor((m) => m?.players && Object.keys(m.players).length === 2, 90000);
  socket.send({ type: "COMMAND_START_GAME", payload: {} });
  console.log("start enviado");
} catch (e) {
  console.error("sem p1:", e.message);
}
await new Promise((r) => setTimeout(r, 240000));
socket.close();
