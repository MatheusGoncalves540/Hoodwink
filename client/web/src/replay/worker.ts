/**
 * Worker clássico do motor de replay (plano 05, T13).
 *
 * O loader manda os bytes JÁ verificados do WASM e do `.hwr`; aqui só rodamos o
 * Go e respondemos aos pedidos. Não há `eval` nem `new Function`: o
 * `wasm_exec.js` da toolchain é carregado com `importScripts` (o nginx e o Vite
 * servem `/replay/`; a CSP tem `worker-src 'self'` e `wasm-unsafe-eval`).
 *
 * Protocolo (uma resposta por pedido, sempre ecoando o id):
 *   -> { id, type: "load", wasm, replay }   <- { id, ok, value: true } | { id, ok: false, error }
 *   -> { id, type: "step" }                 <- bool (false no fim, sem erro)
 *   -> { id, type: "seek", position }       <- bool
 *   -> { id, type: "state" }                <- JSON da sala (string)
 *   -> { id, type: "hash" }                 <- sha256 hex (string)
 *   -> { id, type: "kid" }                  <- kid da assinatura ("" se não assinado)
 *   -> { id, type: "verify", publicKey }    <- bool (false = "não oficial")
 *   -> { id, type: "info" }                 <- JSON {buildVersion, engineVersion, ...}
 *   erro sem id                             <- { type: "error", error }
 *
 * O loader encerra o Worker se um pedido passar de 10 s; o `select {}` do Go
 * mantém o runtime vivo entre os pedidos.
 */
import type {
  ReplayWorkerCommand,
  ReplayWorkerRequest,
  ReplayWorkerResponse,
} from "./loader";

// Mesmo valor de REPLAY_WASM_BASE + REPLAY_WASM_EXEC_FILE (manifest.gen.ts). Fica
// literal porque um worker clássico não pode ter imports em runtime no dev do
// Vite; o import de tipo acima é apagado no build.
const WASM_EXEC_URL = "/replay/wasm_exec.js";

declare function importScripts(...urls: string[]): void;

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

type GoConstructor = new () => GoRuntime;

type WorkerScope = {
  onmessage: ((event: MessageEvent<ReplayWorkerRequest>) => void) | null;
  postMessage(message: ReplayWorkerResponse): void;
};

const scope = self as unknown as WorkerScope;
/** O wasm_exec.js instala as funções do Go no global do Worker. */
const engineScope = globalThis as unknown as Partial<ReplayExports>;

let engineReady: Promise<ReplayExports> | null = null;

/** Exige um replay já carregado (evita instanciar WASM vazio por engano). */
function requireEngine(): ReplayExports {
  if (typeof engineScope.hoodwinkReplayLoad !== "function") {
    throw new Error("replay não carregado");
  }
  return engineScope as ReplayExports;
}

function engineError(): string {
  return typeof engineScope.hoodwinkReplayError === "function"
    ? engineScope.hoodwinkReplayError()
    : "";
}

function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  return String(error);
}

/** Instancia o Go uma vez e espera as funções exportadas aparecerem. */
async function ensureEngine(wasm: Uint8Array): Promise<ReplayExports> {
  if (typeof engineScope.hoodwinkReplayLoad === "function") {
    return engineScope as ReplayExports;
  }
  if (!engineReady) {
    engineReady = (async () => {
      importScripts(WASM_EXEC_URL);

      const Go = (globalThis as { Go?: GoConstructor }).Go;
      if (typeof Go !== "function") {
        throw new Error("wasm_exec.js não definiu Go no Worker");
      }

      const go = new Go();
      const { instance } = await WebAssembly.instantiate(wasm, go.importObject);
      // main() termina em select{}: o runtime fica vivo e a promise só resolve
      // quando o worker for encerrado. Não pode ser aguardada; se o programa
      // sair ou falhar, avisamos o loader como erro fatal (sem id).
      void go.run(instance).then(
        () => scope.postMessage({ type: "error", error: "o motor de replay encerrou" }),
        (error: unknown) => scope.postMessage({ type: "error", error: errorMessage(error) }),
      );

      for (let i = 0; i < 500 && typeof engineScope.hoodwinkReplayLoad !== "function"; i++) {
        await new Promise((resolve) => setTimeout(resolve, 10));
      }
      if (typeof engineScope.hoodwinkReplayLoad !== "function") {
        throw new Error("motor de replay não inicializou");
      }
      return engineScope as ReplayExports;
    })();
  }
  return engineReady;
}

async function dispatch(command: ReplayWorkerCommand): Promise<unknown> {
  switch (command.type) {
    case "load": {
      const engine = await ensureEngine(command.wasm);
      if (!engine.hoodwinkReplayLoad(command.replay)) {
        throw new Error(engineError() || "replay inválido");
      }
      return true;
    }
    case "step": {
      const engine = requireEngine();
      const ok = engine.hoodwinkReplayStep();
      // false no fim é esperado; false com erro é falha de verdade.
      if (!ok && engineError()) throw new Error(engineError());
      return ok;
    }
    case "seek": {
      const engine = requireEngine();
      const ok = engine.hoodwinkReplaySeek(command.position);
      if (!ok) throw new Error(engineError() || "posição inválida no replay");
      return true;
    }
    case "state": {
      return requireEngine().hoodwinkReplayState();
    }
    case "hash": {
      return requireEngine().hoodwinkReplayHash();
    }
    case "kid": {
      return requireEngine().hoodwinkReplayKid();
    }
    case "verify": {
      return requireEngine().hoodwinkReplayVerify(command.publicKey) === true;
    }
    case "info": {
      return requireEngine().hoodwinkReplayInfo();
    }
    default: {
      const unknown = command as { type?: unknown };
      throw new Error(`pedido desconhecido no worker de replay: ${String(unknown.type)}`);
    }
  }
}

async function handle(request: ReplayWorkerRequest): Promise<void> {
  try {
    const value = await dispatch(request);
    scope.postMessage({ id: request.id, ok: true, type: request.type, value });
  } catch (error) {
    scope.postMessage({ id: request.id, ok: false, type: "error", error: errorMessage(error) });
  }
}

scope.onmessage = (event) => {
  const request = event.data;
  if (!request || typeof request.id !== "number") {
    scope.postMessage({ type: "error", error: "pedido inválido no worker de replay" });
    return;
  }
  void handle(request);
};
