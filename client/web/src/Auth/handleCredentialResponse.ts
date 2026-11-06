import { ENV } from "../config/envs";
import decodeJWT from "./decodeJWT";

const post = async (body: Record<string, unknown>) => {
  const res = await fetch(`${ENV.VITE_BACKEND_URL}/auth/external/google`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  return res.json();
};

/** Só grava o login se o token existir e trouxer o id do jogador; senão o ícone apareceria logado e a sala falharia. */
function finish(data: { message?: string; data?: { token?: string } }, setAuth: (token: string, playerId: string) => void) {
  const token = data.data?.token;
  const id = token ? decodeJWT(token)?.id : undefined;
  if (!token || !id) {
    alert("Erro no login: " + (data.message ?? "resposta sem token"));
    return;
  }
  setAuth(token, id);
}

export default async function handleCredentialResponse(
  response: { credential: string },
  setAuth: (token: string, playerId: string) => void,
) {
  const first = await post({ IdToken: response.credential });

  if (first.message === "logged_in") return finish(first, setAuth);

  if (first.message === "need_additional_data") {
    const username = prompt("Novo usuário! Escolha um nome de usuário:")?.trim();
    if (!username) return; // prompt cancelado: não cria conta nem grava token
    return finish(await post({ IdToken: response.credential, username }), setAuth);
  }

  alert("Erro no login: " + first.message);
}
