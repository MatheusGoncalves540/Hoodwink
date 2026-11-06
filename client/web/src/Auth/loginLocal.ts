/**
 * Login sem Google para jogar em rede local (Radmin). O backend só aceita o
 * provedor "local" com ENVIRONMENT=local/development; fora disso a resposta é erro.
 */
import { ENV } from "../config/envs";
import decodeJWT from "./decodeJWT";

export default async function loginLocal(
  username: string,
  setAuth: (token: string, playerId: string) => void,
) {
  let data: { message?: string; data?: { token?: string } } | null = null;
  try {
    const res = await fetch(`${ENV.VITE_BACKEND_URL}/auth/external/local`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username }),
    });
    data = await res.json().catch(() => null);
  } catch {
    // servidor fora do ar ou porta bloqueada no firewall: sem isso o botão ficaria mudo
    alert("Não deu para falar com o servidor do jogo. Confira se ele está no ar.");
    return;
  }

  const token = data?.data?.token;
  const id = token ? decodeJWT(token)?.id : undefined;
  if (!token || !id) {
    alert("Erro no login: " + (data?.message ?? "resposta sem token"));
    return;
  }
  setAuth(token, id);
}
