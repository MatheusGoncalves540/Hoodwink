/**
 * Lê só os deltas das entradas do `.hwr` para a barra de tempo (plano 05, T14).
 *
 * O leitor de verdade continua sendo o Go (`openReplay`): aqui não confiamos no
 * arquivo para nada além de números usados no visor e na reprodução. Qualquer
 * problema devolve lista vazia (o chamador cai no passo a passo), nunca bytes
 * crus na tela e nunca um laço sem fim.
 */
import {
  ReplayInvalidError,
  parseReplayFraming,
  readReplayBytes,
  type ReplayFileLike,
} from "../../replay/loader";

const MAX_BODY_BYTES = 1024 * 1024;
const MAX_PLAYERS = 10;
const MAX_ENTRIES = 5000;
const MAX_DURATION_MS = 6 * 60 * 60 * 1000;

/** Lê o arquivo (com o mesmo teto do loader) e devolve os deltas das entradas. */
export async function readReplayDeltas(file: ReplayFileLike): Promise<number[]> {
  const bytes = await readReplayBytes(file);
  return deltasFromBytes(bytes);
}

export async function deltasFromBytes(bytes: Uint8Array): Promise<number[]> {
  const framing = parseReplayFraming(bytes);
  const body = await gunzipBody(framing.gzipped);
  return decodeDeltas(body);
}

async function gunzipBody(gzipped: Uint8Array): Promise<Uint8Array> {
  const Decompression = (globalThis as { DecompressionStream?: typeof DecompressionStream })
    .DecompressionStream;
  if (typeof Decompression !== "function") return new Uint8Array();

  let reader: ReadableStreamDefaultReader<Uint8Array>;
  try {
    const stream = new Blob([gzipped as unknown as BlobPart])
      .stream()
      .pipeThrough(new Decompression("gzip"));
    reader = stream.getReader();
  } catch {
    throw new ReplayInvalidError();
  }

  const chunks: Uint8Array[] = [];
  let total = 0;
  try {
    for (;;) {
      const { value, done } = await reader.read();
      if (done) break;
      if (!value || value.byteLength === 0) continue;
      total += value.byteLength;
      if (total > MAX_BODY_BYTES) throw new ReplayInvalidError();
      chunks.push(value);
    }
  } catch (error) {
    if (error instanceof ReplayInvalidError) throw error;
    throw new ReplayInvalidError();
  } finally {
    await reader.cancel().catch(() => undefined);
  }

  const out = new Uint8Array(total);
  let offset = 0;
  for (const chunk of chunks) {
    out.set(chunk, offset);
    offset += chunk.byteLength;
  }
  return out;
}

/** Lê o cabeçalho e anda pelas entradas só para pegar os deltas. */
export function decodeDeltas(body: Uint8Array): number[] {
  if (body.byteLength === 0) return [];
  if (body.byteLength < 4 + 1 + 2 + 1 + 16 + 16 + 8 + 1 + 1 + 2) {
    throw new ReplayInvalidError();
  }
  if (body[0] !== 0x48 || body[1] !== 0x57 || body[2] !== 0x52 || body[3] !== 0x31) {
    throw new ReplayInvalidError();
  }
  // Formato novo pode mudar o cabeçalho: a barra cai no passo a passo em vez de ler errado.
  if (body[4] !== 1) return [];

  let o = 4 + 1 + 2; // magic, versão do formato, versão do motor
  const kidLen = need(body, o, 1) ? body[o++] : 0;
  o = skip(body, o, kidLen + 16 + 16 + 8); // kid, matchId, semente, início

  const modeLen = need(body, o, 1) ? body[o++] : 0;
  o = skip(body, o, modeLen);

  const players = need(body, o, 1) ? body[o++] : 0;
  if (players < 2 || players > MAX_PLAYERS) throw new ReplayInvalidError();
  for (let i = 0; i < players; i++) {
    o = skip(body, o, 1); // índice
    const nameLen = need(body, o, 1) ? body[o++] : 0;
    o = skip(body, o, nameLen);
  }

  o = skip(body, o, 1); // flags
  if (!need(body, o, 2)) throw new ReplayInvalidError();
  const entries = (body[o] << 8) | body[o + 1];
  o += 2;
  if (entries < 1 || entries > MAX_ENTRIES) throw new ReplayInvalidError();

  const bitmap = Math.ceil(players / 8);
  const deltas: number[] = [];
  let elapsed = 0;

  for (let i = 0; i < entries; i++) {
    const [raw, next] = readVarint(body, o);
    o = next;
    const delta = raw % 2 === 0 ? raw / 2 : -(raw + 1) / 2;
    if (!Number.isSafeInteger(delta) || delta < 0) throw new ReplayInvalidError();
    elapsed += delta;
    if (elapsed > MAX_DURATION_MS) throw new ReplayInvalidError();
    deltas.push(delta);

    if (!need(body, o, 3)) throw new ReplayInvalidError();
    const flags = body[o + 2];
    o += 3;
    if (flags & 0x01) o = skip(body, o, bitmap);
    if (flags & 0x02) {
      if (!need(body, o, 2)) throw new ReplayInvalidError();
      const payload = (body[o] << 8) | body[o + 1];
      o = skip(body, o + 2, payload);
    }
  }
  return deltas;
}

function need(body: Uint8Array, offset: number, length: number): boolean {
  return offset >= 0 && length >= 0 && offset + length <= body.byteLength;
}

function skip(body: Uint8Array, offset: number, length: number): number {
  if (!need(body, offset, length)) throw new ReplayInvalidError();
  return offset + length;
}

/** Varint zig-zag do Go (`binary.PutVarint`). O teto de 6h mantém tudo em Number seguro. */
function readVarint(body: Uint8Array, offset: number): [number, number] {
  let value = 0;
  let multiplier = 1;
  for (let i = 0; i < 10; i++) {
    if (!need(body, offset, 1)) throw new ReplayInvalidError();
    const byte = body[offset++];
    value += (byte & 0x7f) * multiplier;
    if ((byte & 0x80) === 0) {
      if (!Number.isSafeInteger(value)) throw new ReplayInvalidError();
      return [value, offset];
    }
    multiplier *= 128;
  }
  throw new ReplayInvalidError();
}
