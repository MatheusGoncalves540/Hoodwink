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
  { version: 1, file: "replay-v1-3c51613d.wasm.br", sha256: "3c51613dd74abf55ccb161d027de375233a3eca3327b9a436a27bddee2192240", sha256Br: "169d8d4cd37cffecd81f0077f90d43d3d63106de3a8fe847337925bcf18c2295", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
  { version: 2, file: "replay-v2-21b26e6e.wasm.br", sha256: "21b26e6ee679fbfd9e9e7743e8c80755f5c6f4bb5906ae2f70a41d945e969d50", sha256Br: "f366643705cae6a95543edb64759aad85546f75df3dcb345cfb5ffe8b824f059", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
  { version: 3, file: "replay-v3-ec6a46a4.wasm.br", sha256: "ec6a46a4e07679581c9b21f5c999e6099be4c50e0ea54d5fc49c203833f8e28d", sha256Br: "ef9482113444fe4fde7025599b45d56c9287888cdc5eb8c327c84a2e0cc3a1fd", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
  { version: 4, file: "replay-v4-2bf5eec4.wasm.br", sha256: "2bf5eec4e2f7bc6d72b7a56c93d0a1fde5b3c10fa18598ec2dd46376fe5178ac", sha256Br: "370e66838f31a5631e0c7fac0cdba8b1919b2d68aed7d5f65d8f8aee9cc35309", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
];

export const replayEngineByVersion = (version: number): ReplayEngineAsset | undefined =>
  replayEngines.find((engine) => engine.version === version);

export const replayAssetUrl = (file: string): string => REPLAY_WASM_BASE + file;
