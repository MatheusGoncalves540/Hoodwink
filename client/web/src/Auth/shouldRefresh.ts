/** Regra de renovação do login (plano 03), pura e sem efeitos colaterais. */
export interface TokenLifetime {
  iat?: number;
  exp?: number;
}

function decode(token: string): TokenLifetime {
  try {
    const payload = token.split(".")[1];
    if (!payload) return {};
    const binary = atob(payload.replace(/-/g, "+").replace(/_/g, "/"));
    const json = new TextDecoder().decode(Uint8Array.from(binary, (c) => c.charCodeAt(0)));
    const claims = JSON.parse(json) as { iat?: unknown; exp?: unknown };
    return {
      iat: typeof claims.iat === "number" ? claims.iat : undefined,
      exp: typeof claims.exp === "number" ? claims.exp : undefined,
    };
  } catch {
    return {};
  }
}

/** `now` em segundos (Unix). Renova depois da metade da vida; sem `iat`, quando faltam menos de 12h. */
export default function shouldRefresh(token: string, now: number): boolean {
  const { iat, exp } = decode(token);
  if (typeof exp !== "number") return false;
  if (typeof iat === "number") {
    return now > iat + (exp - iat) / 2;
  }
  return exp - now < 12 * 3600;
}
