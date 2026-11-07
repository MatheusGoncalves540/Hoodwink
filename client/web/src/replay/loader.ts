/**
 * Carregador do motor de replay `.hwr` (plano 05, T13).
 *
 * Fluxo:
 *  1. recusa o arquivo acima de 256 KB ANTES de ler (o `.hwr` é um envelope
 *     hostil: vem do disco ou de um download);
 *  2. lê o manifest embutido (src/replay/manifest.gen.ts) e escolhe o motor;
 *  3. baixa o `.wasm.br` com cache (Cache Storage + HTTP cache), confere o
 *     sha256 dos bytes DESCOMPRIMIDOS contra o manifest e só então manda os
 *     bytes para o Worker. Nunca usamos `instantiateStreaming` direto da rede;
 *     se o hash falhar, apaga do Cache Storage e tenta uma vez;
 *  4. o Worker instancia o Go e roda o replay; cada pedido tem 10 s e o Worker
 *     é encerrado se estourar;
 *  5. a assinatura Ed25519 é conferida pelo motor (export `hoodwinkReplayVerify`)
 *     com a chave pública do `kid` (src/replay/keys.ts). Sem assinatura, kid
 *     desconhecido ou assinatura errada, o replay abre como "não oficial".
 *
 * Nenhuma função aqui usa `eval`/`new Function`.
 */
import { replaySigningKeyByKid, replaySigningKeys, type ReplaySigningKey } from "./keys";
import {
  replayAssetUrl,
  replayEngineByVersion,
  replayEngines,
  type ReplayEngineAsset,
} from "./manifest.gen";

/** Teto do arquivo `.hwr`, igual ao do leitor Go (`format.MaxFileSize`). */
export const MAX_REPLAY_BYTES = 256 * 1024;
/** Tempo máximo de cada pedido ao Worker e do download do motor. */
export const REPLAY_TIMEOUT_MS = 10_000;
/** Nome do cache de artefatos do motor (imutáveis por versão). */
export const WASM_CACHE_NAME = "hoodwink-replay-wasm-v1";

/** Arquivo de replay acima do teto, truncado ou com envelope inválido. */
export class ReplayInvalidError extends Error {
  constructor(message = "replay inválido") {
    super(message);
    this.name = "ReplayInvalidError";
  }
}

/** O motor passou do tempo máximo e o Worker foi encerrado. */
export class ReplayTimeoutError extends Error {
  constructor() {
    super("o motor de replay não respondeu em 10s e foi encerrado");
    this.name = "ReplayTimeoutError";
  }
}

/** Só o que o loader usa de um File/Blob (mantém os testes puros). */
export type ReplayFileLike = {
  readonly size: number;
  arrayBuffer(): Promise<ArrayBuffer>;
};

// ---------------------------------------------------------------------------
// Envelope `.hwr` e utilitários puros
// ---------------------------------------------------------------------------

export type ReplayFraming = {
  /** Comprimento do corpo gzipado (uint32 big-endian no começo do arquivo). */
  bodyLength: number;
  /** Corpo gzipado, exatamente como chegou (é o que a assinatura cobre). */
  gzipped: Uint8Array;
  /** 64 bytes de assinatura Ed25519, ou null quando o arquivo não é assinado. */
  signature: Uint8Array | null;
};

/** Recusa o tamanho antes de qualquer leitura do conteúdo. */
export function assertReplayFileSize(size: number): void {
  if (!Number.isSafeInteger(size) || size < 0 || size > MAX_REPLAY_BYTES) {
    throw new ReplayInvalidError();
  }
}

/**
 * Separa comprimento, corpo gzipado e assinatura sem descomprimir nada.
 * O leitor de verdade (limites de jogadores/entradas/etc.) é o Go; aqui só
 * validamos o envelope para escolher o motor e para nunca entregar lixo ao WASM.
 */
export function parseReplayFraming(bytes: Uint8Array): ReplayFraming {
  if (bytes.byteLength > MAX_REPLAY_BYTES || bytes.byteLength < 4 + 2) {
    throw new ReplayInvalidError();
  }

  const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
  const bodyLength = view.getUint32(0, false);
  if (bodyLength < 8 || 4 + bodyLength > bytes.byteLength) {
    throw new ReplayInvalidError();
  }

  const rest = bytes.byteLength - 4 - bodyLength;
  // Sem assinatura = 0; assinado = exatamente 64 bytes. Qualquer outro resto é lixo.
  if (rest !== 0 && rest !== 64) {
    throw new ReplayInvalidError();
  }

  const gzipped = bytes.subarray(4, 4 + bodyLength);
  if (gzipped[0] !== 0x1f || gzipped[1] !== 0x8b) {
    throw new ReplayInvalidError();
  }

  return {
    bodyLength,
    gzipped,
    signature: rest === 64 ? bytes.subarray(4 + bodyLength) : null,
  };
}

/** Lê o tamanho e os bytes do arquivo, nessa ordem (teto antes do arrayBuffer). */
export async function readReplayBytes(file: ReplayFileLike): Promise<Uint8Array> {
  assertReplayFileSize(file.size);
  const buffer = await file.arrayBuffer();
  if (buffer.byteLength > MAX_REPLAY_BYTES) {
    throw new ReplayInvalidError();
  }
  return new Uint8Array(buffer);
}

/** sha256 em hex minúsculo, com WebCrypto (browser e Node moderno). */
export async function sha256Hex(bytes: Uint8Array): Promise<string> {
  const digest = await crypto.subtle.digest("SHA-256", bytes as unknown as BufferSource);
  const out = new Uint8Array(digest);
  let hex = "";
  for (const byte of out) hex += byte.toString(16).padStart(2, "0");
  return hex;
}

function safeJson(text: string): unknown {
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return null;
  }
}

function asString(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function asNumber(value: unknown): number {
  return typeof value === "number" && Number.isFinite(value) ? value : 0;
}

/**
 * Lê a versão do motor do cabeçalho do `.hwr`. Descomprime só os primeiros
 * bytes com `DecompressionStream('gzip')`; se o navegador não tiver a API,
 * devolve null e o chamador usa o motor mais novo do manifest.
 */
export async function readReplayEngineVersion(
  bytes: Uint8Array,
  framing: ReplayFraming = parseReplayFraming(bytes),
): Promise<number | null> {
  const Decompression = (globalThis as { DecompressionStream?: typeof DecompressionStream })
    .DecompressionStream;
  if (typeof Decompression !== "function") return null;

  const prefix = await gunzipPrefix(framing.gzipped, 32, Decompression);
  // magic "HWR1" | versão do formato (u8) | versão do motor (u16 big-endian)
  if (
    prefix.length < 7 ||
    prefix[0] !== 0x48 ||
    prefix[1] !== 0x57 ||
    prefix[2] !== 0x52 ||
    prefix[3] !== 0x31
  ) {
    throw new ReplayInvalidError();
  }
  return (prefix[5] << 8) | prefix[6];
}

async function gunzipPrefix(
  gzipped: Uint8Array,
  limit: number,
  Decompression: typeof DecompressionStream,
): Promise<Uint8Array> {
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
    while (total < limit) {
      const { value, done } = await reader.read();
      if (done) break;
      if (value && value.byteLength > 0) {
        chunks.push(value);
        total += value.byteLength;
      }
    }
  } catch {
    throw new ReplayInvalidError();
  } finally {
    await reader.cancel().catch(() => undefined);
  }

  const out = new Uint8Array(Math.min(total, limit));
  let offset = 0;
  for (const chunk of chunks) {
    const take = Math.min(chunk.byteLength, out.byteLength - offset);
    out.set(chunk.subarray(0, take), offset);
    offset += take;
    if (offset >= out.byteLength) break;
  }
  return out;
}

// ---------------------------------------------------------------------------
// Selo oficial: assinatura Ed25519 por kid
// ---------------------------------------------------------------------------

export type ReplaySignatureStatus = "valid" | "unsigned" | "unknown-kid" | "invalid";

export type ReplaySignatureCheck = {
  official: boolean;
  status: ReplaySignatureStatus;
};

/** Converte a chave pública hex do bundle para os 32 bytes crus do Ed25519. */
export function decodePublicKeyHex(hex: string): Uint8Array | null {
  const clean = hex.trim().toLowerCase();
  if (!/^[0-9a-f]{64}$/.test(clean)) return null;
  const out = new Uint8Array(32);
  for (let i = 0; i < 32; i++) {
    out[i] = Number.parseInt(clean.slice(i * 2, i * 2 + 2), 16);
  }
  return out;
}

/**
 * Classifica o replay pelo kid e, quando a chave é conhecida, delega a
 * verificação criptográfica ao motor (Go `format.Verify`). Assinatura ausente,
 * kid desconhecido ou assinatura errada NÃO impedem a reprodução: o arquivo
 * abre com a faixa fixa "replay não oficial".
 */
export async function classifySignature(
  kid: string,
  verify: (publicKey: Uint8Array) => Promise<boolean>,
  keys: readonly ReplaySigningKey[] = replaySigningKeys,
): Promise<ReplaySignatureCheck> {
  if (!kid) return { official: false, status: "unsigned" };

  const key = replaySigningKeyByKid(kid, keys);
  const publicKey = key ? decodePublicKeyHex(key.publicKeyHex) : null;
  if (!publicKey) return { official: false, status: "unknown-kid" };

  const valid = await verify(publicKey).catch(() => false);
  return valid ? { official: true, status: "valid" } : { official: false, status: "invalid" };
}

// ---------------------------------------------------------------------------
// Download e integridade do WASM
// ---------------------------------------------------------------------------

/** Subconjunto da Cache Storage API usado aqui (os testes injetam um fake). */
export type ReplayCache = {
  match(request: string): Promise<Response | undefined>;
  put(request: string, response: Response): Promise<void>;
  delete(request: string): Promise<boolean>;
};

export type ReplayCacheStorage = {
  open(name: string): Promise<ReplayCache>;
};

export type LoadEngineOptions = {
  fetchImpl?: typeof fetch;
  /** `undefined` usa `globalThis.caches`; `null` desliga o Cache Storage. */
  cacheStorage?: ReplayCacheStorage | null;
  timeoutMs?: number;
};

function defaultCacheStorage(): ReplayCacheStorage | null {
  return (globalThis as { caches?: ReplayCacheStorage }).caches ?? null;
}

/** Motor mais novo do manifest, usado quando o navegador não sabe ler a versão do arquivo. */
function newestReplayEngine(): ReplayEngineAsset | undefined {
  return replayEngines.reduce<ReplayEngineAsset | undefined>(
    (newest, entry) => (!newest || entry.version > newest.version ? entry : newest),
    undefined,
  );
}

async function fetchAssetBytes(
  fetchImpl: typeof fetch,
  url: string,
  timeoutMs: number,
): Promise<Uint8Array> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  try {
    const response = await fetchImpl(url, { signal: controller.signal, cache: "force-cache" });
    if (!response.ok) {
      throw new Error(`motor de replay indisponível (HTTP ${response.status})`);
    }
    return new Uint8Array(await response.arrayBuffer());
  } catch (error) {
    if (controller.signal.aborted) throw new ReplayTimeoutError();
    throw error;
  } finally {
    clearTimeout(timer);
  }
}

/**
 * Baixa (ou reusa do Cache Storage) o `.wasm.br` da versão e confere o sha256
 * dos bytes descomprimidos. O navegador descomprime o brotli sozinho graças ao
 * `Content-Encoding: br`; se os bytes ainda vierem comprimidos, avisamos em vez
 * de instanciar lixo. Hash errado → apaga do cache e tenta uma vez.
 */
export async function loadEngineBytes(
  asset: ReplayEngineAsset,
  options: LoadEngineOptions = {},
): Promise<Uint8Array> {
  const url = replayAssetUrl(asset.file);
  const expected = asset.sha256.toLowerCase();
  const expectedBr = asset.sha256Br.toLowerCase();
  const cacheStorage = options.cacheStorage === undefined ? defaultCacheStorage() : options.cacheStorage;
  const cache = cacheStorage ? await cacheStorage.open(WASM_CACHE_NAME).catch(() => null) : null;
  const fetchImpl = options.fetchImpl ?? (typeof fetch === "function" ? fetch : undefined);
  const timeoutMs = options.timeoutMs ?? REPLAY_TIMEOUT_MS;

  if (!fetchImpl && !cache) {
    throw new Error("motor de replay indisponível: sem fetch e sem cache");
  }

  // "compressed" = o servidor entregou o .br sem Content-Encoding (Vite dev);
  // nesse caso caímos para o .wasm descomprimido publicado ao lado.
  const accept = async (bytes: Uint8Array): Promise<"ok" | "compressed" | "bad"> => {
    const actual = await sha256Hex(bytes);
    if (actual === expectedBr && expectedBr !== expected) return "compressed";
    return actual === expected ? "ok" : "bad";
  };

  let compressed = false;
  for (let attempt = 0; attempt < 2; attempt++) {
    if (cache) {
      const hit = await cache.match(url).catch(() => undefined);
      if (hit) {
        const cached = new Uint8Array(await hit.arrayBuffer());
        const result = await accept(cached);
        if (result === "ok") return cached;
        if (result === "compressed") {
          compressed = true;
          break;
        }
        await cache.delete(url).catch(() => false);
      }
    }

    if (!fetchImpl) break;
    const fresh = await fetchAssetBytes(fetchImpl, url, timeoutMs);
    const result = await accept(fresh);
    if (result === "ok") {
      if (cache) {
        await cache
          .put(url, new Response(fresh as unknown as BodyInit, { headers: { "content-type": "application/wasm" } }))
          .catch(() => undefined);
      }
      return fresh;
    }
    if (result === "compressed") {
      compressed = true;
      break;
    }
    // Cache envenenado / download corrompido: não deixa o próximo carregamento
    // tropeçar no mesmo byte e tenta mais uma vez.
    if (cache) await cache.delete(url).catch(() => false);
  }

  // Fallback de desenvolvimento: o Vite não envia Content-Encoding: br, mas
  // publicamos o .wasm descomprimido ao lado do .br. Confere o mesmo sha256.
  if (compressed && fetchImpl) {
    const fallbackUrl = replayAssetUrl(asset.file.replace(/\.br$/, ""));
    const fallback = await fetchAssetBytes(fetchImpl, fallbackUrl, timeoutMs);
    if ((await accept(fallback)) === "ok") {
      if (cache) {
        await cache
          .put(fallbackUrl, new Response(fallback as unknown as BodyInit, { headers: { "content-type": "application/wasm" } }))
          .catch(() => undefined);
      }
      return fallback;
    }
  }

  if (!fetchImpl) {
    throw new Error("motor de replay indisponível: sem fetch e sem cache utilizável");
  }
  throw new ReplayInvalidError("motor de replay inválido");
}

// ---------------------------------------------------------------------------
// Protocolo com o Worker (classic worker)
// ---------------------------------------------------------------------------

export type ReplayWorkerCommand =
  | { type: "load"; wasm: Uint8Array; replay: Uint8Array }
  | { type: "step" }
  | { type: "seek"; position: number }
  | { type: "state" }
  | { type: "hash" }
  | { type: "kid" }
  | { type: "verify"; publicKey: Uint8Array }
  | { type: "info" };

export type ReplayWorkerRequest = ReplayWorkerCommand & { id: number };

export type ReplayWorkerResponse = {
  /** id do pedido; ausente em erro fatal do Worker (que não subiu, por exemplo). */
  id?: number;
  ok?: boolean;
  type?: ReplayWorkerCommand["type"] | "error";
  value?: unknown;
  error?: string;
};

/** Canal de pedido/resposta com o motor (o Worker real ou um fake nos testes). */
export type ReplayEngineChannel = {
  request(command: ReplayWorkerCommand, transfer?: Transferable[]): Promise<unknown>;
  close(): void;
};

export type ReplayChannelFactory = () => ReplayEngineChannel | Promise<ReplayEngineChannel>;

/** Só o que o loader usa de um Worker (facilita o teste de timeout). */
export type ReplayWorkerLike = {
  postMessage(message: ReplayWorkerRequest, transfer?: Transferable[]): void;
  terminate(): void;
  addEventListener(
    type: string,
    listener: (event: { data?: ReplayWorkerResponse; message?: string }) => void,
  ): void;
};

type PendingRequest = {
  resolve(value: unknown): void;
  reject(error: Error): void;
  timer: ReturnType<typeof setTimeout>;
};

/**
 * Envolve um Worker. Cada pedido tem `timeoutMs` (10 s por padrão) e, ao
 * estourar, o Worker é encerrado — um motor travado não pode segurar a página.
 */
export function createReplayWorkerChannel(
  options: { timeoutMs?: number; workerFactory?: () => ReplayWorkerLike } = {},
): ReplayEngineChannel {
  const timeoutMs = options.timeoutMs ?? REPLAY_TIMEOUT_MS;
  const worker =
    options.workerFactory?.() ??
    (new Worker(new URL("./worker.ts", import.meta.url), {
      type: "classic",
      name: "hoodwink-replay",
    }) as unknown as ReplayWorkerLike);

  const pending = new Map<number, PendingRequest>();
  let nextId = 1;
  let closed = false;

  const failAll = (error: Error) => {
    if (closed) return;
    closed = true;
    for (const entry of pending.values()) {
      clearTimeout(entry.timer);
      entry.reject(error);
    }
    pending.clear();
    worker.terminate();
  };

  worker.addEventListener("message", (event) => {
    const response = event.data;
    if (!response) return;
    if (typeof response.id !== "number") {
      failAll(new Error(response.error || "falha no worker de replay"));
      return;
    }
    const entry = pending.get(response.id);
    if (!entry) return;
    clearTimeout(entry.timer);
    pending.delete(response.id);
    if (response.ok) entry.resolve(response.value);
    else entry.reject(new Error(response.error || "falha no motor de replay"));
  });

  worker.addEventListener("error", (event) => {
    failAll(new Error(event.message || "falha no worker de replay"));
  });

  return {
    request(command, transfer) {
      if (closed) return Promise.reject(new Error("motor de replay encerrado"));
      const id = nextId++;
      return new Promise<unknown>((resolve, reject) => {
        const timer = setTimeout(() => failAll(new ReplayTimeoutError()), timeoutMs);
        pending.set(id, { resolve, reject, timer });
        try {
          worker.postMessage({ ...command, id } as ReplayWorkerRequest, transfer);
        } catch (error) {
          clearTimeout(timer);
          pending.delete(id);
          reject(error instanceof Error ? error : new Error(String(error)));
        }
      });
    },
    close() {
      failAll(new Error("motor de replay encerrado"));
    },
  };
}

// ---------------------------------------------------------------------------
// API de alto nível
// ---------------------------------------------------------------------------

export type ReplaySessionInfo = {
  buildVersion: string;
  engineVersion: number;
  formatVersion: number;
  mode: string;
  players: number;
  steps: number;
  position: number;
  done: boolean;
};

export function parseReplayInfo(value: unknown): ReplaySessionInfo {
  const raw = typeof value === "string" ? safeJson(value) : value;
  if (!raw || typeof raw !== "object") {
    throw new Error("info do replay inválida");
  }
  const info = raw as Record<string, unknown>;
  return {
    buildVersion: asString(info.buildVersion),
    engineVersion: asNumber(info.engineVersion),
    formatVersion: asNumber(info.formatVersion),
    mode: asString(info.mode),
    players: asNumber(info.players),
    steps: asNumber(info.steps),
    position: asNumber(info.position),
    done: info.done === true,
  };
}

export type ReplaySession = {
  /** true só quando a assinatura confere com a chave do kid. */
  readonly official: boolean;
  readonly signature: ReplaySignatureStatus;
  readonly kid: string;
  readonly engine: ReplayEngineAsset;
  /** Pula a próxima entrada; false no fim (não é erro). */
  step(): Promise<boolean>;
  /** Posiciona no passo 0..steps. */
  seek(position: number): Promise<boolean>;
  /** JSON do estado da sala; "" antes do primeiro passo. */
  state(): Promise<string>;
  /** sha256 hex do estado canônico (confere com o checkpoint do `.hwr`). */
  hash(): Promise<string>;
  /** Info atualizada (posição, fim, total de passos). */
  info(): Promise<ReplaySessionInfo>;
  /** Encerra o Worker. */
  close(): void;
};

export type OpenReplayOptions = LoadEngineOptions & {
  /** Manifest override (testes). Sem isso, o manifest embutido decide. */
  engine?: ReplayEngineAsset;
  /** Substitui o download/verificação do WASM (testes). */
  loadEngine?: (asset: ReplayEngineAsset) => Promise<Uint8Array>;
  /** Substitui o Worker por um canal em processo (testes). */
  channelFactory?: ReplayChannelFactory;
  /** Chaves públicas; padrão = lista embutida de keys.ts. */
  keys?: readonly ReplaySigningKey[];
};

/**
 * Abre um `.hwr` no motor da versão certa. O arquivo nunca é lido acima de
 * 256 KB; o wasm só é instanciado depois do sha256 conferir; a assinatura não
 * impede a reprodução, só define o selo oficial.
 */
export async function openReplay(
  file: ReplayFileLike,
  options: OpenReplayOptions = {},
): Promise<ReplaySession> {
  const bytes = await readReplayBytes(file);
  const framing = parseReplayFraming(bytes);
  const version = options.engine
    ? options.engine.version
    : await readReplayEngineVersion(bytes, framing);
  const engine =
    options.engine ??
    (version === null ? newestReplayEngine() : replayEngineByVersion(version));

  if (!engine) {
    throw new ReplayInvalidError("versão do motor de replay indisponível");
  }

  const wasm = await (options.loadEngine ?? ((asset: ReplayEngineAsset) => loadEngineBytes(asset, options)))(
    engine,
  );
  const channel = options.channelFactory
    ? await options.channelFactory()
    : createReplayWorkerChannel({ timeoutMs: options.timeoutMs });

  try {
    // Os dois buffers são transferidos: o Worker é o único dono deles daqui em diante.
    const transfer: Transferable[] = [
      wasm.buffer as ArrayBuffer,
      bytes.buffer as ArrayBuffer,
    ];
    await channel.request({ type: "load", wasm, replay: bytes }, transfer);

    const kid = asString(await channel.request({ type: "kid" }));
    const signature = await classifySignature(
      kid,
      async (publicKey) =>
        (await channel.request({ type: "verify", publicKey }, [publicKey.buffer as ArrayBuffer])) === true,
      options.keys,
    );

    return {
      official: signature.official,
      signature: signature.status,
      kid,
      engine,
      async step() {
        return (await channel.request({ type: "step" })) === true;
      },
      async seek(position) {
        if (!Number.isInteger(position) || position < 0) return false;
        return (await channel.request({ type: "seek", position })) === true;
      },
      async state() {
        return asString(await channel.request({ type: "state" }));
      },
      async hash() {
        return asString(await channel.request({ type: "hash" }));
      },
      async info() {
        return parseReplayInfo(await channel.request({ type: "info" }));
      },
      close() {
        channel.close();
      },
    };
  } catch (error) {
    channel.close();
    throw error;
  }
}
