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
  { version: 1, file: "replay-v1-708e003a.wasm.br", sha256: "708e003ae3a7619f3aba73836d9ac5ae61a056f4bb6f4039f59c91e51f36ff52", sha256Br: "0aa37cb328fba746b8fb8c12b0c84297831eaaeb5d1cdc2bb7ec06cac4530132", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
  { version: 2, file: "replay-v2-042a928c.wasm.br", sha256: "042a928ccc9e2a8c660745e63d496d93ba9d69f82915693d7db88e83101cc987", sha256Br: "12d37b334b2d2458ce9b1849b74f0b8b2fd87f03be3fa120f3846980344b3003", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
  { version: 3, file: "replay-v3-541036a0.wasm.br", sha256: "541036a044d3f575ead811da8b277d2c716b65bc5f8801b78a77e4f5251792c5", sha256Br: "5ec97a9a9bdb9bca78330cea7e288df04321c9a9520a168dd5dd851f3d3486ac", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
];

export const replayEngineByVersion = (version: number): ReplayEngineAsset | undefined =>
  replayEngines.find((engine) => engine.version === version);

export const replayAssetUrl = (file: string): string => REPLAY_WASM_BASE + file;
