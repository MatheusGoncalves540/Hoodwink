/**
 * Harness de caos do Hoodwink (sem dependências).
 * Lê os .env locais do backend e do gameServer para emitir JWTs de teste,
 * fala HTTP com /newRoom e /getTicket e abre o WebSocket do jogo.
 *
 * Uso: rode da raiz do repo. Os caminhos podem vir de HOODWINK_BACKEND_ENV e
 * HOODWINK_GAME_ENV; as URLs, de GAME_URL e WEB_URL.
 */
import { createHmac, sign } from "node:crypto";
import { readFileSync } from "node:fs";

export const BACKEND_ENV = process.env.HOODWINK_BACKEND_ENV ?? "D:/GitHub/Hoodwink-server/backend/.env";
export const GAME_ENV = process.env.HOODWINK_GAME_ENV ?? "D:/GitHub/Hoodwink-server/gameServer/.env";
export const GAME_URL = process.env.GAME_URL ?? "http://localhost:4444";
export const WEB_URL = process.env.WEB_URL ?? "http://localhost:5180";

export function readEnv(path) {
  const out = {};
  for (const line of readFileSync(path, "utf8").split(/\r?\n/)) {
    const clean = line.trim();
    if (!clean || clean.startsWith("#") || !clean.includes("=")) continue;
    const i = clean.indexOf("=");
    out[clean.slice(0, i).trim()] = clean.slice(i + 1).trim().replace(/^"|"$/g, "");
  }
  return out;
}

let cached;
export function secrets() {
  if (cached) return cached;
  const backend = readEnv(BACKEND_ENV);
  if (!backend.JWT_SECRET) {
    throw new Error("JWT_SECRET do backend ausente");
  }
  if (!backend.TICKET_PRIVATE_KEY) {
    throw new Error("TICKET_PRIVATE_KEY do backend ausente");
  }
  cached = { backend: backend.JWT_SECRET, ticketPrivateKey: backend.TICKET_PRIVATE_KEY };
  return cached;
}

const b64 = (value) => Buffer.from(JSON.stringify(value)).toString("base64url");

/** JWT HS256 igual ao do backend (structures/jwt.go): id, username, provider, email, temp, exp. */
export function jwt(id, username, ttlSeconds = 3600) {
  const header = b64({ alg: "HS256", typ: "JWT" });
  const payload = b64({
    id,
    username,
    provider: "caos",
    email: `${id}@caos.local`,
    temp: false,
    exp: Math.floor(Date.now() / 1000) + ttlSeconds,
  });
  const signature = createHmac("sha256", secrets().backend).update(`${header}.${payload}`).digest("base64url");
  return `${header}.${payload}.${signature}`;
}

/** Ticket do WebSocket: claims playerId, username, roomId, exp, aud; assinado Ed25519 com a TICKET_PRIVATE_KEY do backend. */
export function ticketJwt(playerId, username, roomId, ttlSeconds = 3600) {
  const header = b64({ alg: "EdDSA", typ: "JWT" });
  const payload = b64({
    playerId,
    username,
    roomId,
    aud: "hoodwink-ticket",
    exp: Math.floor(Date.now() / 1000) + ttlSeconds,
  });
  const data = Buffer.from(`${header}.${payload}`);
  const privPem = Buffer.from(secrets().ticketPrivateKey, "base64").toString("utf8");
  const signature = sign(null, data, privPem).toString("base64url");
  return `${header}.${payload}.${signature}`;
}

export async function post(path, token, body) {
  const res = await fetch(`${GAME_URL}${path}`, {
    method: "POST",
    headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const text = await res.text();
  let json = null;
  try {
    json = JSON.parse(text);
  } catch {
    /* resposta não-JSON */
  }
  return { status: res.status, json, text };
}

export const newRoom = (token, input = {}) =>
  post("/newRoom", token, { roomName: "Caos E2E", password: "", maxPlayers: 6, ...input });

export async function getTicket(token, roomId) {
  const res = await post(`/getTicket/${encodeURIComponent(roomId)}`, token);
  const ticket = res.json?.data?.ticket;
  if (!ticket) throw new Error(`sem ticket (${res.status}): ${res.text}`);
  return ticket;
}

/** Abre o WebSocket e guarda toda mensagem recebida (PublicRoomForUpdates). */
export function connect(ticket) {
  const url = `${GAME_URL.replace(/^http/, "ws").replace(/\/$/, "")}/game?Ticket=${encodeURIComponent(ticket)}`;
  const ws = new WebSocket(url);
  const messages = [];
  const errors = [];
  ws.onmessage = (event) => {
    try {
      messages.push(JSON.parse(event.data));
    } catch {
      messages.push({ raw: String(event.data) });
    }
  };
  ws.onerror = (event) => errors.push(String(event.message ?? event.type));
  const waitOpen = new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error("timeout abrindo o socket")), 5000);
    ws.onopen = () => {
      clearTimeout(timer);
      resolve();
    };
  });
  const waitFor = (predicate, timeout = 8000) =>
    new Promise((resolve, reject) => {
      const found = messages.find(predicate);
      if (found) return resolve(found);
      const timer = setTimeout(() => reject(new Error("timeout esperando mensagem")), timeout);
      const check = () => {
        const hit = messages.find(predicate);
        if (!hit) return;
        clearTimeout(timer);
        ws.removeEventListener("message", check);
        resolve(hit);
      };
      ws.addEventListener("message", check);
    });
  return {
    ws,
    messages,
    errors,
    waitOpen,
    waitFor,
    send: (message) => ws.send(JSON.stringify(message)),
    close: () => ws.close(),
  };
}

/** A sala pública que o servidor manda: players é objeto por id. */
export const latestRoom = (socket) => [...socket.messages].reverse().find((m) => m && typeof m === "object" && m.players);
