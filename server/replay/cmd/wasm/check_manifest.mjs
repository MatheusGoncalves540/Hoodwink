// Confere o que `make replay-wasm` gerou: cada versão em replay/versions tem um
// .wasm (cru) e um .wasm.br, o sha256 do manifest bate nos bytes descomprimidos
// (o que o loader confere antes do instantiate) e no arquivo .br (integridade do
// artefato publicado); o wasm_exec.js confere, não há arquivo órfão e o tamanho
// está dentro do teto.
//
// Uso: node server/replay/cmd/wasm/check_manifest.mjs [client/web]
import { createHash } from "node:crypto";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { brotliDecompressSync } from "node:zlib";

// Tetos generosos para pegar regressão (hoje ≈ 1,1 MiB brotli / 5,4 MiB wasm).
const MAX_BROTLI_BYTES = 4 * 1024 * 1024;
const MAX_WASM_BYTES = 12 * 1024 * 1024;

const here = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(here, "../../../..");
const webDir = path.resolve(process.argv[2] || path.join(repoRoot, "client/web"));
const manifestPath = path.join(webDir, "src/replay/manifest.gen.ts");
const assetDir = path.join(webDir, "public/replay");

const fail = (msg) => {
  console.error(`ERRO: ${msg}`);
  process.exit(1);
};

const sha256 = (buf) => createHash("sha256").update(buf).digest("hex");

if (!fs.existsSync(manifestPath)) fail(`manifest ausente: ${manifestPath} (rode "make replay-wasm" em server/)`);
if (!fs.existsSync(assetDir)) fail(`pasta de assets ausente: ${assetDir} (rode "make replay-wasm" em server/)`);

const manifest = fs.readFileSync(manifestPath, "utf8");
const entryRe =
  /\{\s*version:\s*(\d+),\s*file:\s*["']([^"']+)["'],\s*sha256:\s*["']([0-9a-f]{64})["'],\s*sha256Br:\s*["']([0-9a-f]{64})["'],\s*wasmExecSha256:\s*["']([0-9a-f]{64})["'],\s*brotli:\s*true\s*\}/g;

const entries = [...manifest.matchAll(entryRe)].map((m) => ({
  version: Number(m[1]),
  file: m[2],
  sha256: m[3],
  sha256Br: m[4],
  wasmExecSha256: m[5],
}));

if (entries.length === 0) fail("manifest sem nenhuma versão (formato inesperado)");

const seen = new Set();
for (const e of entries) {
  if (seen.has(e.version)) fail(`versão duplicada no manifest: v${e.version}`);
  seen.add(e.version);

  if (!new RegExp(`^replay-v${e.version}-[0-9a-f]{8}\\.wasm\\.br$`).test(e.file)) {
    fail(`nome fora do padrão replay-v<N>-<8 hex>.wasm.br: ${e.file}`);
  }
  const filePath = path.join(assetDir, e.file);
  if (!fs.existsSync(filePath)) fail(`arquivo do manifest não existe: ${filePath}`);

  	const compressed = fs.readFileSync(filePath);
  	if (compressed.length > MAX_BROTLI_BYTES) {
  		fail(`${e.file} tem ${compressed.length} B brotli (teto ${MAX_BROTLI_BYTES} B)`);
  	}
  	const compressedSha = sha256(compressed);
  	if (compressedSha !== e.sha256Br) {
  		fail(`${e.file}: sha256 do .wasm.br ${compressedSha} != manifest ${e.sha256Br}`);
  	}

  let wasm;
  try {
    wasm = brotliDecompressSync(compressed);
  } catch (err) {
    fail(`${e.file} não descomprime com brotli: ${err.message}`);
  }
  if (wasm.length > MAX_WASM_BYTES) {
    fail(`${e.file} tem ${wasm.length} B descomprimido (teto ${MAX_WASM_BYTES} B)`);
  }
  if (wasm.subarray(0, 4).toString("latin1") !== "\0asm") {
    fail(`${e.file} não começa com o magic \\0asm`);
  }

  const got = sha256(wasm);
  if (got !== e.sha256) fail(`${e.file}: sha256 ${got} != manifest ${e.sha256}`);

  // O .wasm cru fica ao lado do .br (fallback/inspeção) e tem que ser o mesmo
  // binário; os dois são conferidos contra o sha256 do manifest.
  const rawFile = e.file.replace(/\.br$/, "");
  const rawPath = path.join(assetDir, rawFile);
  if (!fs.existsSync(rawPath)) fail(`.wasm cru ausente: ${rawPath}`);
  const raw = fs.readFileSync(rawPath);
  const rawSha = sha256(raw);
  if (rawSha !== e.sha256) fail(`${rawFile}: sha256 ${rawSha} != manifest ${e.sha256}`);
  if (!raw.equals(wasm)) fail(`${rawFile} difere do conteúdo de ${e.file}`);

  	console.log(
  		`ok  v${e.version} ${e.file}: ${compressed.length} B brotli (${raw.length} B wasm), sha256 confere no .wasm e no .wasm.br`,
  	);
}

// Nenhum .wasm/.wasm.br órfão (versão antiga esquecida na pasta).
const manifestFiles = new Set(entries.map((e) => e.file));
const manifestRaw = new Set(entries.map((e) => e.file.replace(/\.br$/, "")));
for (const name of fs.readdirSync(assetDir)) {
  if (!name.startsWith("replay-v")) continue;
  if (name.endsWith(".wasm.br") && !manifestFiles.has(name)) {
    fail(`arquivo órfão em public/replay (fora do manifest): ${name}`);
  }
  if (name.endsWith(".wasm") && !manifestRaw.has(name)) {
    fail(`.wasm órfão em public/replay (fora do manifest): ${name}`);
  }
}

const wasmExecPath = path.join(assetDir, "wasm_exec.js");
if (!fs.existsSync(wasmExecPath)) fail(`wasm_exec.js ausente em ${assetDir}`);
const wasmExecSha = sha256(fs.readFileSync(wasmExecPath));
for (const e of entries) {
  if (e.wasmExecSha256 !== wasmExecSha) {
    fail(`wasm_exec.js sha256 ${wasmExecSha} != manifest ${e.wasmExecSha256}`);
  }
}
console.log(`ok  wasm_exec.js: sha256 confere (${wasmExecSha})`);

console.log(`\n${entries.length} versão(ões) do motor conferida(s) em ${assetDir}`);
