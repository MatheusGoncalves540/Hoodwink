import { generateKeyPairSync, sign, type KeyObject } from "node:crypto";
import { readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import vm from "node:vm";
import { brotliDecompressSync, gunzipSync, gzipSync } from "node:zlib";
import { beforeAll, describe, expect, it, vi } from "vitest";

import { replaySigningKeys } from "./keys";
import {
  MAX_REPLAY_BYTES,
  ReplayInvalidError,
  ReplayTimeoutError,
  assertReplayFileSize,
  classifySignature,
  createReplayWorkerChannel,
  decodePublicKeyHex,
  loadEngineBytes,
  openReplay,
  parseReplayFraming,
  readReplayBytes,
  readReplayEngineVersion,
  sha256Hex,
  type ReplayCacheStorage,
  type ReplayEngineChannel,
  type ReplayFileLike,
  type ReplayWorkerCommand,
  type ReplayWorkerLike,
} from "./loader";
import { replayEngines } from "./manifest.gen";

const here = dirname(fileURLToPath(import.meta.url));
const webRoot = resolve(here, "../..");
const repoRoot = resolve(webRoot, "../..");
const publicReplay = join(webRoot, "public", "replay");
const goldenPath = join(repoRoot, "server", "replay", "engines", "v1", "testdata", "golden_1.hwr");
/** Hash final do golden_1.hwr (engines/v1/testdata/golden_1.sha256). */
const GOLDEN_FINAL_HASH = "1a453dd51cde6e05206a58247eb8c303b6fb8274ac05dbf81a636b5ef81bb627";

const engine = replayEngines[0];
const wasmFile = join(publicReplay, engine.file.replace(/\.wasm\.br$/, ".wasm"));

let golden: Uint8Array;
let realWasm: Uint8Array;

beforeAll(() => {
  // O wasm_exec.js é um script clássico: no browser/worker vem por importScripts;
  // no Vitest (Node) rodamos no contexto global para ganhar globalThis.Go.
  vm.runInThisContext(readFileSync(join(publicReplay, "wasm_exec.js"), "utf8"));
  golden = new Uint8Array(readFileSync(goldenPath));
  realWasm = new Uint8Array(readFileSync(wasmFile));
});

function fileOf(bytes: Uint8Array): ReplayFileLike {
  return {
    size: bytes.byteLength,
    // Cópia: o loader pode transferir o buffer para o Worker.
    arrayBuffer: async () =>
      bytes.buffer.slice(bytes.byteOffset, bytes.byteOffset + bytes.byteLength) as ArrayBuffer,
  };
}

// ---------------------------------------------------------------------------
// Motor Go em processo, no lugar do Worker (o mesmo WASM real do bundle)
// ---------------------------------------------------------------------------

type ReplayExports = {
  hoodwinkReplayLoad(bytes: Uint8Array): boolean;
  hoodwinkReplayStep(): boolean;
  hoodwinkReplaySeek(position: number): boolean;
  hoodwinkReplayState(): string;
  hoodwinkReplayHash(): string;
  hoodwinkReplayKid(): string;
  hoodwinkReplayVerify(publicKey: Uint8Array): boolean;
  hoodwinkReplayInfo(): string;
  hoodwinkReplayError(): string;
};

type GoRuntime = {
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void>;
};

let enginePromise: Promise<ReplayExports> | null = null;

/** Um Go por arquivo de teste; cada `load` reseta a sessão do motor. */
function sharedEngine(): Promise<ReplayExports> {
  if (!enginePromise) {
    enginePromise = (async () => {
      const Go = (globalThis as unknown as { Go?: new () => GoRuntime }).Go;
      if (typeof Go !== "function") throw new Error("wasm_exec.js não definiu Go");

      const go = new Go();
      const { instance } = await WebAssembly.instantiate(new Uint8Array(realWasm), go.importObject);
      // main() termina em select{}: o runtime fica vivo e a promise não resolve.
      void go.run(instance);

      for (let i = 0; i < 500; i++) {
        if (typeof (globalThis as unknown as Partial<ReplayExports>).hoodwinkReplayLoad === "function") {
          break;
        }
        await new Promise((resolve) => setTimeout(resolve, 10));
      }
      const api = globalThis as unknown as Partial<ReplayExports>;
      if (typeof api.hoodwinkReplayLoad !== "function") {
        throw new Error("motor de replay não inicializou");
      }
      return api as ReplayExports;
    })();
  }
  return enginePromise;
}

function goChannel(): ReplayEngineChannel {
  return {
    async request(command: ReplayWorkerCommand) {
      const api = await sharedEngine();
      switch (command.type) {
        case "load":
          return api.hoodwinkReplayLoad(command.replay);
        case "step":
          return api.hoodwinkReplayStep();
        case "seek":
          return api.hoodwinkReplaySeek(command.position);
        case "state":
          return api.hoodwinkReplayState();
        case "hash":
          return api.hoodwinkReplayHash();
        case "kid":
          return api.hoodwinkReplayKid();
        case "verify":
          return api.hoodwinkReplayVerify(command.publicKey);
        case "info":
          return api.hoodwinkReplayInfo();
        default:
          throw new Error("pedido inesperado no teste");
      }
    },
    close() {
      // O runtime compartilhado continua vivo entre os testes.
    },
  };
}

// ---------------------------------------------------------------------------
// Assinatura de um golden para os testes de "oficial"
// ---------------------------------------------------------------------------

function signReplay(file: Uint8Array, kid: string, privateKey: KeyObject): Uint8Array {
  const view = new DataView(file.buffer, file.byteOffset, file.byteLength);
  const bodyLength = view.getUint32(0, false);
  const body = gunzipSync(file.subarray(4, 4 + bodyLength));
  if (body[7] !== 0) throw new Error("golden de teste já tem assinatura");

  const kidBytes = Buffer.from(kid, "utf8");
  const withKid = Buffer.concat([
    body.subarray(0, 7),
    Buffer.from([kidBytes.byteLength]),
    kidBytes,
    body.subarray(8),
  ]);

  const gzipped = gzipSync(withKid);
  const signature = sign(null, Buffer.concat([Buffer.from("HWR1-SIG"), gzipped]), privateKey);

  const out = Buffer.alloc(4 + gzipped.byteLength + signature.byteLength);
  out.writeUInt32BE(gzipped.byteLength, 0);
  gzipped.copy(out, 4);
  signature.copy(out, 4 + gzipped.byteLength);
  return new Uint8Array(out);
}

function publicKeyHex(publicKey: KeyObject): string {
  return publicKey.export({ type: "spki", format: "der" }).subarray(-32).toString("hex");
}

describe("tamanho e envelope do .hwr", () => {
  it("recusa 300 KB antes de ler", async () => {
    const arrayBuffer = vi.fn(async () => new ArrayBuffer(0));
    const file: ReplayFileLike = { size: 300 * 1024, arrayBuffer };

    await expect(readReplayBytes(file)).rejects.toThrow("replay inválido");
    await expect(openReplay(file)).rejects.toThrow("replay inválido");
    expect(arrayBuffer).not.toHaveBeenCalled();
    expect(() => assertReplayFileSize(MAX_REPLAY_BYTES)).not.toThrow();
    expect(() => assertReplayFileSize(MAX_REPLAY_BYTES + 1)).toThrow(ReplayInvalidError);
  });

  it("separa comprimento, gzip e assinatura", () => {
    const framing = parseReplayFraming(golden);
    expect(framing.bodyLength).toBe(golden.byteLength - 4);
    expect(framing.signature).toBeNull();
    expect(framing.gzipped[0]).toBe(0x1f);
    expect(framing.gzipped[1]).toBe(0x8b);

    const { publicKey, privateKey } = generateKeyPairSync("ed25519");
    expect(publicKey).toBeDefined();
    const signed = signReplay(golden, "replay-v1", privateKey);
    const signedFraming = parseReplayFraming(signed);
    expect(signedFraming.signature).toHaveLength(64);
    expect(signedFraming.bodyLength).toBe(signed.byteLength - 4 - 64);
  });

  it("recusa envelope truncado, com lixo no fim ou gzip errado", () => {
    expect(() => parseReplayFraming(golden.subarray(0, 10))).toThrow(ReplayInvalidError);

    const lixo = new Uint8Array(golden.byteLength + 1);
    lixo.set(golden);
    expect(() => parseReplayFraming(lixo)).toThrow(ReplayInvalidError);

    const semGzip = new Uint8Array(golden);
    semGzip[4] = 0;
    expect(() => parseReplayFraming(semGzip)).toThrow(ReplayInvalidError);
  });

  it("lê a versão do motor do cabeçalho descomprimido", async () => {
    await expect(readReplayEngineVersion(golden)).resolves.toBe(engine.version);
  });

  it("sha256 confere com o vetor conhecido", async () => {
    await expect(sha256Hex(new TextEncoder().encode("abc"))).resolves.toBe(
      "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
    );
  });
});

describe("manifest do WASM", () => {
  it("os artefatos publicados batem com os sha256 do manifest", async () => {
    for (const entry of replayEngines) {
      const compressed = new Uint8Array(readFileSync(join(publicReplay, entry.file)));
      expect(await sha256Hex(compressed)).toBe(entry.sha256Br);

      const decompressed = new Uint8Array(brotliDecompressSync(compressed));
      expect(await sha256Hex(decompressed)).toBe(entry.sha256);

      const name = entry.file.replace(/\.wasm\.br$/, ".wasm");
      expect(name).toContain(`v${entry.version}-${entry.sha256.slice(0, 8)}`);
      expect(await sha256Hex(new Uint8Array(readFileSync(join(publicReplay, name))))).toBe(entry.sha256);
    }

    const exec = new Uint8Array(readFileSync(join(publicReplay, "wasm_exec.js")));
    expect(await sha256Hex(exec)).toBe(engine.wasmExecSha256);
  });

  it("1 byte trocado no wasm é recusado e o cache é limpo", async () => {
    const tampered = new Uint8Array(realWasm);
    tampered[100] ^= 0xff;

    let stored: Uint8Array | null = null;
    const cache = {
      match: vi.fn(async () => (stored ? new Response(stored) : undefined)),
      put: vi.fn(async (_url: string, response: Response) => {
        stored = new Uint8Array(await response.arrayBuffer());
      }),
      delete: vi.fn(async () => {
        stored = null;
        return true;
      }),
    };
    const cacheStorage: ReplayCacheStorage = { open: async () => cache };
    const fetchImpl = vi.fn(async () => new Response(tampered));

    await expect(
      loadEngineBytes(engine, { cacheStorage, fetchImpl: fetchImpl as unknown as typeof fetch }),
    ).rejects.toThrow("motor de replay inválido");
    expect(fetchImpl).toHaveBeenCalledTimes(2); // tenta uma vez depois de limpar o cache
    expect(cache.delete).toHaveBeenCalled();
  });

  it("cache envenenado é apagado e o download bom é usado", async () => {
    let stored: Uint8Array | null = new Uint8Array(realWasm);
    stored[10] ^= 0xff;

    const cache = {
      match: vi.fn(async () => (stored ? new Response(stored) : undefined)),
      put: vi.fn(async (_url: string, response: Response) => {
        stored = new Uint8Array(await response.arrayBuffer());
      }),
      delete: vi.fn(async () => {
        stored = null;
        return true;
      }),
    };
    const cacheStorage: ReplayCacheStorage = { open: async () => cache };
    const fetchImpl = vi.fn(async () => new Response(new Uint8Array(realWasm)));

    const bytes = await loadEngineBytes(engine, {
      cacheStorage,
      fetchImpl: fetchImpl as unknown as typeof fetch,
    });
    expect(await sha256Hex(bytes)).toBe(engine.sha256);
    expect(cache.delete).toHaveBeenCalled();
    expect(cache.put).toHaveBeenCalled();
  });

  it("openReplay recusa wasm adulterado antes de criar o Worker", async () => {
    const tampered = new Uint8Array(realWasm);
    tampered[42] ^= 0xff;
    const cacheStorage: ReplayCacheStorage = {
      open: async () => ({
        match: async () => undefined,
        put: async () => undefined,
        delete: async () => true,
      }),
    };
    const channelFactory = vi.fn(() => goChannel());

    await expect(
      openReplay(fileOf(golden), {
        engine,
        cacheStorage,
        fetchImpl: (async () => new Response(tampered)) as unknown as typeof fetch,
        channelFactory,
      }),
    ).rejects.toThrow("motor de replay inválido");
    expect(channelFactory).not.toHaveBeenCalled();
  });

  it("servidor sem Content-Encoding br cai para o .wasm descomprimido (Vite dev)", async () => {
    const compressed = new Uint8Array(readFileSync(join(publicReplay, engine.file)));
    const fetchImpl = vi.fn(async (url: string) =>
      new Response(url.endsWith(".br") ? compressed : new Uint8Array(realWasm)),
    );

    const bytes = await loadEngineBytes(engine, {
      fetchImpl: fetchImpl as unknown as typeof fetch,
    });
    expect(await sha256Hex(bytes)).toBe(engine.sha256);
    expect(fetchImpl).toHaveBeenCalledTimes(2);
  });

  it("cache bom evita o download", async () => {
    const cache = {
      match: vi.fn(async () => new Response(new Uint8Array(realWasm))),
      put: vi.fn(async () => undefined),
      delete: vi.fn(async () => true),
    };
    const cacheStorage: ReplayCacheStorage = { open: async () => cache };
    const fetchImpl = vi.fn(async () => new Response(new Uint8Array(realWasm)));

    const bytes = await loadEngineBytes(engine, {
      cacheStorage,
      fetchImpl: fetchImpl as unknown as typeof fetch,
    });
    expect(await sha256Hex(bytes)).toBe(engine.sha256);
    expect(fetchImpl).not.toHaveBeenCalled();
  });
});

describe("classificação oficial/não oficial", () => {
  it("sem assinatura é não oficial e não chama o verificador", async () => {
    const verify = vi.fn(async () => true);
    await expect(classifySignature("", verify)).resolves.toEqual({
      official: false,
      status: "unsigned",
    });
    expect(verify).not.toHaveBeenCalled();
  });

  it("kid desconhecido é não oficial", async () => {
    const verify = vi.fn(async () => true);
    await expect(classifySignature("replay-que-nao-existe", verify, replaySigningKeys)).resolves.toEqual({
      official: false,
      status: "unknown-kid",
    });
    expect(verify).not.toHaveBeenCalled();
  });

  it("assinatura válida é oficial; inválida ou erro do motor não são", async () => {
    const key = replaySigningKeys[0];
    await expect(classifySignature(key.kid, async () => true)).resolves.toEqual({
      official: true,
      status: "valid",
    });
    await expect(classifySignature(key.kid, async () => false)).resolves.toEqual({
      official: false,
      status: "invalid",
    });
    await expect(
      classifySignature(key.kid, async () => {
        throw new Error("assinatura inválida");
      }),
    ).resolves.toEqual({ official: false, status: "invalid" });
  });

  it("a chave de desenvolvimento tem 32 bytes e o kid replay-v1", () => {
    const key = replaySigningKeys.find((entry) => entry.kid === "replay-v1");
    expect(key).toBeDefined();
    expect(decodePublicKeyHex(key!.publicKeyHex)).toHaveLength(32);
    expect(decodePublicKeyHex("não é hex")).toBeNull();
  });
});

describe("worker com 10s", () => {
  it("encerra o Worker e rejeita quando o motor não responde", async () => {
    const worker: ReplayWorkerLike = {
      postMessage: vi.fn(),
      terminate: vi.fn(),
      addEventListener: vi.fn(),
    };

    const channel = createReplayWorkerChannel({ timeoutMs: 20, workerFactory: () => worker });
    await expect(channel.request({ type: "state" })).rejects.toBeInstanceOf(ReplayTimeoutError);
    expect(worker.terminate).toHaveBeenCalled();
  });

  it("resolve a resposta pelo id", async () => {
    let onMessage: ((event: { data?: unknown }) => void) | undefined;
    const worker: ReplayWorkerLike = {
      postMessage: (message) => {
        queueMicrotask(() =>
          onMessage?.({
            data: { id: (message as { id: number }).id, ok: true, type: message.type, value: 42 },
          }),
        );
      },
      terminate: vi.fn(),
      addEventListener: (type, listener) => {
        if (type === "message") onMessage = listener as (event: { data?: unknown }) => void;
      },
    };

    const channel = createReplayWorkerChannel({ timeoutMs: 1000, workerFactory: () => worker });
    await expect(channel.request({ type: "info" })).resolves.toBe(42);
  });
});

describe("motor real com o golden_1.hwr", () => {
  it("roda as 105 entradas e chega no mesmo hash final do Go", async () => {
    const session = await openReplay(fileOf(golden), {
      engine,
      loadEngine: async () => new Uint8Array(realWasm),
      channelFactory: () => goChannel(),
    });

    expect(session.official).toBe(false);
    expect(session.signature).toBe("unsigned");
    expect(session.kid).toBe("");
    expect((await session.info()).steps).toBe(105);

    let steps = 0;
    while (await session.step()) steps++;
    expect(steps).toBe(105);
    expect(await session.hash()).toBe(GOLDEN_FINAL_HASH);
    expect(await session.state()).toContain('"gameOver":true');

    session.close();
  });

  it("volta pela barra de tempo e chega no mesmo estado", async () => {
    const session = await openReplay(fileOf(golden), {
      engine,
      loadEngine: async () => new Uint8Array(realWasm),
      channelFactory: () => goChannel(),
    });

    expect(await session.seek(100)).toBe(true);
    while (await session.step()) {
      // anda os últimos 5 passos
    }
    expect(await session.hash()).toBe(GOLDEN_FINAL_HASH);
    session.close();
  });

  it("assinatura válida marca oficial", async () => {
    const { publicKey, privateKey } = generateKeyPairSync("ed25519");
    const signed = signReplay(golden, "replay-v1", privateKey);

    const session = await openReplay(fileOf(signed), {
      engine,
      loadEngine: async () => new Uint8Array(realWasm),
      channelFactory: () => goChannel(),
      keys: [{ kid: "replay-v1", publicKeyHex: publicKeyHex(publicKey) }],
    });

    expect(session.kid).toBe("replay-v1");
    expect(session.official).toBe(true);
    expect(session.signature).toBe("valid");
    session.close();
  });

  it("assinatura trocada sai não oficial", async () => {
    const { publicKey, privateKey } = generateKeyPairSync("ed25519");
    const signed = signReplay(golden, "replay-v1", privateKey);
    signed[signed.byteLength - 1] ^= 0xff;

    const session = await openReplay(fileOf(signed), {
      engine,
      loadEngine: async () => new Uint8Array(realWasm),
      channelFactory: () => goChannel(),
      keys: [{ kid: "replay-v1", publicKeyHex: publicKeyHex(publicKey) }],
    });

    expect(session.official).toBe(false);
    expect(session.signature).toBe("invalid");
    session.close();
  });

  it("kid desconhecido sai não oficial", async () => {
    const { privateKey } = generateKeyPairSync("ed25519");
    const signed = signReplay(golden, "replay-desconhecido", privateKey);

    const session = await openReplay(fileOf(signed), {
      engine,
      loadEngine: async () => new Uint8Array(realWasm),
      channelFactory: () => goChannel(),
    });

    expect(session.official).toBe(false);
    expect(session.signature).toBe("unknown-kid");
    session.close();
  });
});
