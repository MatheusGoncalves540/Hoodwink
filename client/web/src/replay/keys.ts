/**
 * Chaves públicas Ed25519 que assinam os replays `.hwr` (plano 05, T13).
 *
 * O backend assina cada arquivo com `REPLAY_SIGNING_KEY` e grava o `kid` no
 * cabeçalho; o worker confere a assinatura com a pública correspondente. O kid
 * é público, então a lista pode ser embutida no bundle.
 *
 * DESENVOLVIMENTO: `replay-v1` abaixo é a pública derivada da chave privada de
 * exemplo em `server/backend/.env.example`. PRODUÇÃO: a chave privada vem do
 * secret do deploy e NÃO é a mesma — troque a pública aqui antes de publicar,
 * senão todo replay oficial sai "não oficial". Ao rotacionar a chave, mantenha
 * as antigas enquanto houver replay assinado com elas (o formato já aceita mais
 * de um kid).
 */

export type ReplaySigningKey = {
  kid: string;
  /** Chave pública Ed25519 em hex (32 bytes = 64 caracteres). */
  publicKeyHex: string;
};

const devKeys: readonly ReplaySigningKey[] = [
  {
    kid: "replay-v1",
    publicKeyHex: "55a8e79444667b78382e29a6f8d11706187041922ac35ec4f64836520132c95b",
  },
];

// O deploy da VM passa a pública de produção no build (VITE_REPLAY_PUBLIC_KEY,
// 64 hex) e o kid em VITE_REPLAY_KID. Sem isso, o bundle usa a chave de dev.
const envKey = (import.meta.env.VITE_REPLAY_PUBLIC_KEY ?? "").trim().toLowerCase();
const envKid = (import.meta.env.VITE_REPLAY_KID ?? "replay-v1").trim() || "replay-v1";

const normalizedEnvKey = /^[0-9a-f]{64}$/.test(envKey) ? envKey : "";

export const replaySigningKeys: readonly ReplaySigningKey[] = normalizedEnvKey
  ? [{ kid: envKid, publicKeyHex: normalizedEnvKey }, ...devKeys.filter((key) => key.kid !== envKid)]
  : devKeys;

/** Resolve o kid para a chave pública embutida. */
export const replaySigningKeyByKid = (
  kid: string,
  keys: readonly ReplaySigningKey[] = replaySigningKeys,
): ReplaySigningKey | undefined => keys.find((key) => key.kid === kid);
