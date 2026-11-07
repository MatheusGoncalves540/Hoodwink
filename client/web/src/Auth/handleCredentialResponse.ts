import { ENV } from "../config/envs";
import { copy } from "../game/copy";
import decodeJWT from "./decodeJWT";

export type LoginOutcome =
  | { status: "ok" }
  | { status: "need_name" }
  | { status: "error"; message: string };

type CredentialReply = { message?: string; data?: { token?: string } };

const post = async (body: Record<string, unknown>): Promise<CredentialReply> => {
  const res = await fetch(`${ENV.VITE_BACKEND_URL}/auth/external/google`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  return res.json();
};

/** Mensagem clara para o que o backend devolveu; "logged_in" sem token vira o aviso genérico. */
function reasonOf(data: CredentialReply): string {
  if (!data.message || data.message === "logged_in") return copy.loginModal.noToken;
  if (data.message === "username_taken") return copy.loginModal.usernameTaken;
  return data.message;
}

/** Só grava o login se o token existir e trouxer o id do jogador; senão o ícone apareceria logado e a sala falharia. */
function finish(data: CredentialReply, setAuth: (token: string, playerId: string) => void): LoginOutcome {
  const token = data.data?.token;
  const id = token ? decodeJWT(token)?.id : undefined;
  if (!token || !id) return { status: "error", message: copy.loginModal.error(reasonOf(data)) };
  setAuth(token, id);
  return { status: "ok" };
}

export default async function handleCredentialResponse(
  response: { credential: string },
  setAuth: (token: string, playerId: string) => void,
  username?: string,
): Promise<LoginOutcome> {
  let first: CredentialReply;
  try {
    first = await post({ IdToken: response.credential });
  } catch {
    return { status: "error", message: copy.loginModal.offline };
  }

  if (first.message === "logged_in") return finish(first, setAuth);

  if (first.message === "need_additional_data") {
    const clean = username?.trim();
    if (!clean) return { status: "need_name" };
    try {
      return finish(await post({ IdToken: response.credential, username: clean }), setAuth);
    } catch {
      return { status: "error", message: copy.loginModal.offline };
    }
  }

  return { status: "error", message: copy.loginModal.error(reasonOf(first)) };
}
