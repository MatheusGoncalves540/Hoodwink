/** Fumaça: cria uma sala, conecta três jogadores e imprime o estado público. */
import { connect, jwt, newRoom, ticketJwt } from "./lib.mjs";

const run = Date.now().toString(36);
const created = await newRoom(jwt(`caos-ana-${run}`, "Ana Caos"), { roomName: "Fumaça caos" });
if (created.status !== 200 && created.status !== 201) throw new Error(`newRoom ${created.status}: ${created.text}`);
const roomId = created.json.RoomId;
console.log("sala", roomId);

const players = [
  [`caos-ana-${run}`, "Ana Caos"],
  [`caos-bruno-${run}`, "Bruno Caos"],
  [`caos-carla-${run}`, "Carla Caos"],
];
const sockets = [];
for (const [id, name] of players) {
  const socket = connect(ticketJwt(id, name, roomId));
  await socket.waitOpen;
  sockets.push({ id, name, socket });
}
await sockets[0].socket.waitFor((m) => m && m.players && Object.keys(m.players).length === 3);
const room = [...sockets[0].socket.messages].reverse().find((m) => m && m.players);
console.log("estado", {
  id: room.id,
  jogadores: Object.values(room.players).map((p) => `${p.name} (${p.coins})`),
  deck: room.deck,
  startTime: room.StartTime ?? room.startTime,
});
for (const { socket } of sockets) socket.close();
