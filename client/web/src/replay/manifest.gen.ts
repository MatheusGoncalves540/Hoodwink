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
  { version: 1, file: "replay-v1-f9e92d0e.wasm.br", sha256: "f9e92d0e3573642253909d148c37e60dbda9a50de2cdb52387e19178676914b7", sha256Br: "da62d191ad85c444309a0c25f9c9cbac0ccab408a35aee961f9103ea131fb640", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
];

export const replayEngineByVersion = (version: number): ReplayEngineAsset | undefined =>
  replayEngines.find((engine) => engine.version === version);

export const replayAssetUrl = (file: string): string => REPLAY_WASM_BASE + file;
