// Code generated por `make replay-wasm` (server/makefile). NAO edite a mao.
//
// Um .wasm por versao do motor congelado. O `sha256` e dos bytes DESCOMPRIMIDOS
// do .wasm - e o que o loader confere antes de WebAssembly.instantiate. Em
// producao o nginx serve o .wasm.br com Content-Encoding: br, entao o fetch ja
// entrega os bytes descomprimidos. O `sha256Br` e do arquivo .wasm.br em si
// (integridade do artefato publicado; nao e o hash usado no instantiate).

export type ReplayEngineAsset = {
  version: number;
  /** Nome do arquivo em public/replay/, comprimido com brotli. */
  file: string;
  /** sha256 (hex) dos bytes descomprimidos do .wasm. */
  sha256: string;
  /** sha256 (hex) do arquivo .wasm.br servido pelo nginx. */
  sha256Br: string;
  /** sha256 (hex) do wasm_exec.js da toolchain Go usada no build. */
  wasmExecSha256: string;
  brotli: true;
};

export const REPLAY_WASM_BASE = "/replay/";
export const REPLAY_WASM_EXEC_FILE = "wasm_exec.js";

export const replayEngines: readonly ReplayEngineAsset[] = [
  { version: 1, file: "replay-v1-ca85cce6.wasm.br", sha256: "ca85cce60961bf9d19d306071ecff396e33117098d2011047532d2c0505399ea", sha256Br: "289d5140ecdbdeb99ae22b7ffd93d6b228905d579cafa0855fd3642374fdae11", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
  { version: 2, file: "replay-v2-51852979.wasm.br", sha256: "51852979635ba79f992edd2f6dfd0a6a5a916a559e0cd3e97acc37413f42f725", sha256Br: "175f4f013353be1de18a83e411daa7d983e8f9051320454906182c12618dffc4", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
];

export const replayEngineByVersion = (version: number): ReplayEngineAsset | undefined =>
  replayEngines.find((engine) => engine.version === version);

export const replayAssetUrl = (file: string): string => REPLAY_WASM_BASE + file;
