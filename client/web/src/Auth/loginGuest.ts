/**
 * Entrada de visitante: um nome, sem conta. O backend cria um usuário
 * temporário que entra em salas pelo código; criar sala exige conta.
 */
import { ENV } from "../config/envs";
import { copy } from "../game/copy";
import decodeJWT from "./decodeJWT";
import type { LoginOutcome } from "./handleCredentialResponse";

export default async function loginGuest(
  username: string,
  setAuth: (token: string, playerId: string) => void,
): Promise<LoginOutcome> {
  let data: { message?: string; data?: { token?: string } } | null = null;
  try {
    const res = await fetch(`${ENV.VITE_BACKEND_URL}/auth/external/guest`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username }),
    });
    data = await res.json().catch(() => null);
  } catch {
    // servidor fora do ar ou porta bloqueada no firewall: o modal mostra o aviso
    return { status: "error", message: copy.loginModal.offline };
  }

  const token = data?.data?.token;
  const id = token ? decodeJWT(token)?.id : undefined;
  if (!token || !id) {
    const reason = data?.message === "username_taken" ? copy.loginModal.usernameTaken : (data?.message ?? copy.loginModal.noToken);
    return { status: "error", message: copy.loginModal.error(reason) };
  }
  setAuth(token, id);
  return { status: "ok" };
}
