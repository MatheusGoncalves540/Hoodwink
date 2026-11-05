/**
 * Suíte de caos HTTP/WS do Hoodwink. Cada rodada é independente e registra
 * status, erro e tempo. Não altera dados fora das salas de teste.
 */
import { connect, getTicket, jwt, newRoom, post, ticketJwt } from "./lib.mjs";

const runId = Date.now().toString(36);
const results = [];
const run = async (name, fn) => {
  const started = Date.now();
  try {
    const detail = await fn();
    results.push({ name, ok: true, detail, ms: Date.now() - started });
  } catch (error) {
    results.push({ name, ok: false, detail: String(error?.message ?? error), ms: Date.now() - started });
  }
};
const expectStatus = (res, allowed) => {
  if (!allowed.includes(res.status)) throw new Error(`status ${res.status}: ${res.text.slice(0, 200)}`);
  return res.status;
};
const oneRoom = async (input) => {
  const token = jwt(`s-${Math.random().toString(36).slice(2)}`, "Caos");
  const res = await newRoom(token, input);
  if (res.status !== 201) throw new Error(`newRoom ${res.status}: ${res.text}`);
  return { token, roomId: res.json.RoomId };
};
const tryOpen = (ticket, ms = 2500) =>
  new Promise((resolve) => {
    const ws = new WebSocket(`ws://localhost:4444/game?Ticket=${encodeURIComponent(ticket)}`);
    const timer = setTimeout(() => { ws.close(); resolve({ opened: false, reason: "timeout" }); }, ms);
    ws.onopen = () => { clearTimeout(timer); resolve({ opened: true, ws }); };
    ws.onerror = () => { clearTimeout(timer); resolve({ opened: false, reason: "error" }); };
    ws.onclose = (e) => { clearTimeout(timer); resolve({ opened: false, reason: `close ${e.code}` }); };
  });

// --- HTTP ---
await run("newRoom sem Authorization", async () => {
  const res = await post("/newRoom", "", { roomName: "x", maxPlayers: 6 });
  return expectStatus(res, [401]);
});
await run("newRoom com JWT inválido", async () => {
  const res = await post("/newRoom", "abc.def.ghi", { roomName: "x", maxPlayers: 6 });
  return expectStatus(res, [401]);
});
await run("newRoom JSON inválido", async () => {
  const res = await fetch("http://localhost:4444/newRoom", { method: "POST", headers: { Authorization: `Bearer ${jwt("bad-json", "Caos")}`, "Content-Type": "application/json" }, body: "{" });
  return expectStatus({ status: res.status, text: await res.text() }, [400]);
});
await run("newRoom maxPlayers 1", async () => {
  const res = await newRoom(jwt("mp1", "Caos"), { maxPlayers: 1 });
  return expectStatus(res, [400]);
});
await run("newRoom maxPlayers 11", async () => {
  const res = await newRoom(jwt("mp11", "Caos"), { maxPlayers: 11 });
  return expectStatus(res, [400]);
});
await run("newRoom nome vazio", async () => {
  const res = await newRoom(jwt("nome-vazio", "Caos"), { roomName: "" });
  return expectStatus(res, [400]);
});
await run("newRoom nome de 10 mil caracteres", async () => {
  const res = await newRoom(jwt("nome-longo", "Caos"), { roomName: "x".repeat(10_000) });
  return expectStatus(res, [201, 400]);
});
await run("newRoom com campo extra", async () => {
  const res = await newRoom(jwt("extra", "Caos"), { roomName: "Extra", hack: true });
  return expectStatus(res, [201, 400]);
});
await run("newRoom duas vezes com o mesmo JWT", async () => {
  const token = jwt("duas-salas", "Caos");
  const a = await newRoom(token, { roomName: "A" });
  const b = await newRoom(token, { roomName: "B" });
  return { a: a.status, b: b.status, ids: [a.json?.RoomId, b.json?.RoomId] };
});
await run("getTicket sala inexistente", async () => {
  const res = await getTicket(jwt("sem-sala", "Caos"), "0000000000").catch((e) => ({ status: "throw", text: String(e.message) }));
  return res.status ?? res;
});
await run("getTicket JWT inválido", async () => {
  const res = await post("/getTicket/abc", "x.y.z");
  return expectStatus(res, [401]);
});

// --- WS ---
await run("WS ticket inválido", async () => {
  const out = await tryOpen("abc.def.ghi");
  if (out.opened) throw new Error("abriu com ticket inválido");
  return out.reason;
});
await run("WS ticket expirado", async () => {
  const { roomId } = await oneRoom({ roomName: "Expirado" });
  const out = await tryOpen(ticketJwt("exp", "Caos", roomId, -10));
  if (out.opened) throw new Error("abriu com ticket expirado");
  return out.reason;
});
await run("WS mesmo ticket duas vezes", async () => {
  const { roomId } = await oneRoom({ roomName: "Duplicado" });
  const ticket = ticketJwt("dup", "Caos", roomId);
  const first = await tryOpen(ticket);
  const second = await tryOpen(ticket);
  if (!first.opened) throw new Error("primeira conexão não abriu");
  first.ws.close();
  return { primeira: first.opened, segunda: second.opened, motivo: second.reason };
});
await run("WS JSON malformado", async () => {
  const { roomId } = await oneRoom({ roomName: "Malformado" });
  const s = connect(ticketJwt("malform", "Caos", roomId));
  await s.waitOpen;
  s.ws.send("{");
  await new Promise((r) => setTimeout(r, 500));
  return { estado: s.ws.readyState };
});
await run("WS tipo desconhecido", async () => {
  const { roomId } = await oneRoom({ roomName: "Desconhecido" });
  const s = connect(ticketJwt("unknown", "Caos", roomId));
  await s.waitOpen;
  s.send({ type: "NAO_EXISTE", payload: {} });
  await new Promise((r) => setTimeout(r, 500));
  return { estado: s.ws.readyState };
});
await run("WS jogada fora do turno", async () => {
  const a = jwt("turno-a", "Ana");
  const room = await newRoom(a, { roomName: "Fora do turno" });
  const roomId = room.json.RoomId;
  const s1 = connect(ticketJwt("turno-a", "Ana", roomId));
  const s2 = connect(ticketJwt("turno-b", "Bruno", roomId));
  await s1.waitOpen; await s2.waitOpen;
  s2.send({ type: "TRILLIONAIRE", payload: null });
  await new Promise((r) => setTimeout(r, 500));
  return { estado: s2.ws.readyState };
});
await run("WS campo a mais na jogada", async () => {
  const { roomId } = await oneRoom({ roomName: "Campo extra" });
  const s = connect(ticketJwt("extra-play", "Caos", roomId));
  await s.waitOpen;
  s.send({ type: "TRILLIONAIRE", payload: { targetPlayer: "x" } });
  await new Promise((r) => setTimeout(r, 500));
  return { estado: s.ws.readyState };
});
await run("sala cheia por tickets", async () => {
  const token = jwt("cheia", "Caos");
  const room = await newRoom(token, { roomName: "Cheia", maxPlayers: 2 });
  const roomId = room.json.RoomId;
  const s1 = connect(ticketJwt("cheia-1", "Um", roomId));
  const s2 = connect(ticketJwt("cheia-2", "Dois", roomId));
  await s1.waitOpen; await s2.waitOpen;
  const s3 = await tryOpen(ticketJwt("cheia-3", "Tres", roomId));
  s1.close(); s2.close();
  if (s3.opened) s3.ws.close();
  return { terceiroAbriu: s3.opened, motivo: s3.reason };
});
await run("start por jogador que não é o anfitrião", async () => {
  const token = jwt(`host-${runId}`, "Host");
  const room = await newRoom(token, { roomName: "Start" });
  const roomId = room.json.RoomId;
  const host = connect(ticketJwt(`host-${runId}`, "Host", roomId));
  const guest = connect(ticketJwt(`guest-${runId}`, "Guest", roomId));
  await host.waitOpen; await guest.waitOpen;
  try {
    await host.waitFor((m) => m?.players && Object.keys(m.players).length === 2, 12000);
    guest.send({ type: "COMMAND_START_GAME", payload: {} });
  } catch { /* segue para o resultado */ }
  const started = await guest.waitFor((m) => m?.startTime && m.startTime !== "0001-01-01T00:00:00Z", 12000).then(() => true).catch(() => false);
  host.close(); guest.close();
  return { iniciouSemSerHost: started };
});
await run("reconexão com o mesmo ticket", async () => {
  const { roomId } = await oneRoom({ roomName: "Reconexão" });
  const ticket = ticketJwt("reconecta", "Caos", roomId);
  const first = connect(ticket);
  await first.waitOpen;
  await first.waitFor((m) => m?.players, 5000);
  first.close();
  await new Promise((r) => setTimeout(r, 600));
  const again = connect(ticket);
  try {
    await again.waitOpen;
    const room = await again.waitFor((m) => m?.players, 5000);
    return { reconectou: true, jogadores: Object.keys(room.players).length };
  } finally {
    again.close();
  }
});

console.log(JSON.stringify(results, null, 2));
const failures = results.filter((r) => !r.ok);
console.error(`
${results.length} rodadas, ${failures.length} falhas`);

process.exit(0);
