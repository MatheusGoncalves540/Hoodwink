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
  { version: 1, file: "replay-v1-faa9a71c.wasm.br", sha256: "faa9a71c6fcc021e184b8aaaacd6540823af4ee78332bb2401cee6dfb69f767d", sha256Br: "7cd6e89f7c18d772add4ff3bd6cefbb87b5b5ee5dc066486099b2df1b1c7bb9e", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
  { version: 2, file: "replay-v2-06bc609b.wasm.br", sha256: "06bc609ba724eed6f5b882e960dd3c90759ae3a75fb0d885ee26c2b0d330c79a", sha256Br: "962041e1fed2eac63f497716590e75ce9e222e784cf534b0c0007d7e3bbfa548", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
  { version: 3, file: "replay-v3-97baceb3.wasm.br", sha256: "97baceb3a1ea3a3dea824c1cdcd4c92e4d01b9966ce0db92d5a92ea6c7321553", sha256Br: "6ed9c36afd8a7729fca4561560b83566d75f4f3782a79efd1d7b88a051c87a70", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
  { version: 4, file: "replay-v4-84c63195.wasm.br", sha256: "84c63195ea5735a6440f9b489c86d02402bcedb445036cb7c75537cfc9b14362", sha256Br: "a2daee2f461b68ff7567122a34db3cbd40d8dbea378546f376ee90123b92e22e", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
  { version: 5, file: "replay-v5-48fff3a2.wasm.br", sha256: "48fff3a29ab96b8ad9df497990889d56efd1112dacf5b6756beb5c52b41f020a", sha256Br: "7d23f7880aadbbac57833d16fd2893d2c34dd38aecf31d04dd611108b08e854b", wasmExecSha256: "0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14", brotli: true },
];

export const replayEngineByVersion = (version: number): ReplayEngineAsset | undefined =>
  replayEngines.find((engine) => engine.version === version);

export const replayAssetUrl = (file: string): string => REPLAY_WASM_BASE + file;
