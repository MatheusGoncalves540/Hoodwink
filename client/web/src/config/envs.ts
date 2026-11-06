import { z } from "zod";

const envSchema = z.object({
  VITE_BACKEND_URL: z.url(),
  VITE_FRONTEND_URL: z.url(),
  VITE_GAMESERVER_URL: z.url(),
  VITE_GOOGLE_CLIENT_ID: z.string(),
});

const parsed = envSchema.parse(import.meta.env);

/**
 * Em dev os .env apontam para localhost. Quem entra pelo IP do Radmin carrega a página
 * desse IP, e "localhost" no navegador dele não é esta máquina: trocamos o host pelo
 * hostname atual, sem precisar fixar o IP em cada .env.
 */
const sameHost = (url: string) =>
  typeof location === "undefined" ? url : url.replace("//localhost", `//${location.hostname}`);

export const ENV = {
  ...parsed,
  VITE_BACKEND_URL: sameHost(parsed.VITE_BACKEND_URL),
  VITE_FRONTEND_URL: sameHost(parsed.VITE_FRONTEND_URL),
  VITE_GAMESERVER_URL: sameHost(parsed.VITE_GAMESERVER_URL),
};
