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
  { version: 1, file: "replay-v1-dfb7ec00.wasm.br", sha256: "dfb7ec00ca1d841727df38f4c9cebb9e0a8b3a0d37f1e038c604f36281c3d215", sha256Br: "1e4e2740fab560c2f9c595993ea5ef1c690fd0007b179307cac8cd645536733a", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
  { version: 2, file: "replay-v2-9a5b66e8.wasm.br", sha256: "9a5b66e8a92e23854da7f9770cfc2ae36ff8b2aa198403d3f7bc3f413bb21d87", sha256Br: "89a5aebc4a16edd524f3a9bdaedbf8dcb8e692716c031fcfbb0342df901bcbc2", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
];

export const replayEngineByVersion = (version: number): ReplayEngineAsset | undefined =>
  replayEngines.find((engine) => engine.version === version);

export const replayAssetUrl = (file: string): string => REPLAY_WASM_BASE + file;
