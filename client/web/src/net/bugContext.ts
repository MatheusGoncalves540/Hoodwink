/**
 * Contexto automático dos relatos de bug (plano 06, T4).
 * Guarda os últimos erros de JS sem cair em URL com ticket nem em credenciais.
 * O backend não confia neste contexto: ele ainda passa pela máscara do plano 04.
 */

import { GAME_VERSION } from "../version";

export type BugContext = {
  scene: string;
  buildVersion: string;
  screen: string;
  errors: string[];
};

const MAX_ERRORS = 10;
const MAX_ERROR_CHARS = 500;

const errors: string[] = [];
let installed = false;

/** Tira query e hash de URLs, preservando o :linha:coluna do stack. */
const stripUrlDetails = (text: string) => {
  const strip = (match: string) => {
    const cut = match.search(/[?#]/);
    const suffix = match.match(/(?::\d+){1,2}$/)?.[0] ?? "";
    return `${match.slice(0, cut)}${suffix}`;
  };
  return text
    .replace(/(?:https?:\/\/|\/\/)[^\s?#]*[?#][^\s)"']*/g, strip)
    .replace(/\/[^\s?#]*[?#][^\s)"']*/g, strip)
    .replace(/(^|[\s(])\?[^\s)"']+/g, "$1");
};

const stripSecrets = (text: string) =>
  text
    .replace(/\bBearer\s+[^\s,;)"']+/gi, "Bearer ***")
    .replace(/\b(token|ticket)=[^\s&,;)"']*/gi, "$1=***");

/** Limpa uma mensagem ou stack antes de guardar. Exportado para o teste e para o cartão. */
export const sanitizeBugText = (text: string) => stripSecrets(stripUrlDetails(text));

const describeValue = (value: unknown): string => {
  if (value instanceof Error) return `${value.name}: ${value.message}${value.stack ? `\n${value.stack}` : ""}`;
  if (typeof value === "string") return value;
  try {
    return JSON.stringify(value) ?? String(value);
  } catch {
    return String(value);
  }
};

const pushError = (text: string) => {
  errors.push(sanitizeBugText(text).slice(0, MAX_ERROR_CHARS));
  if (errors.length > MAX_ERRORS) errors.splice(0, errors.length - MAX_ERRORS);
};

const onError = (event: ErrorEvent) => {
  if (!event.error && !event.message) return; // erro de carregamento de imagem/script não traz mensagem
  const detail = describeValue(event.error ?? event.message);
  const where = event.filename
    ? ` (${event.filename}${event.lineno ? `:${event.lineno}` : ""}${event.colno ? `:${event.colno}` : ""})`
    : "";
  pushError(`${detail}${where}`);
};

const onRejection = (event: PromiseRejectionEvent) => pushError(describeValue(event.reason));

/** Escuta `error` e `unhandledrejection`. Idempotente e inofensivo fora do navegador. */
export function installBugContext(): void {
  if (installed || typeof window === "undefined" || typeof window.addEventListener !== "function") return;
  installed = true;
  window.addEventListener("error", onError);
  window.addEventListener("unhandledrejection", onRejection);
}

/** Zera os erros guardados: usado antes de um novo relato e nos testes. */
export function resetBugContext(): void {
  errors.length = 0;
}

const currentScene = (): string => {
  if (typeof document !== "undefined") {
    const marked = document.documentElement?.dataset.scene;
    if (marked) return marked;
  }
  if (typeof window === "undefined" || !window.location) return "desconhecida";
  const hash = window.location.hash ?? "";
  if (hash) return hash.replace(/^#/, "");
  const search = window.location.search ?? "";
  const params = new URLSearchParams(search);
  const devScene = params.get("cena");
  if (devScene) return devScene;
  return params.has("sala") ? "sala" : "lobby";
};


const currentScreen = () =>
  typeof window === "undefined" ? "0x0" : `${window.innerWidth}x${window.innerHeight}`;

/**
 * Devolve o contexto para mandar no relato. `sendDevice = false` desliga a coleta
 * (a chave "Enviar dados do aparelho" do cartão) e devolve `null`.
 */
export function collectBugContext(sendDevice = true): BugContext | null {
  if (!sendDevice) return null;
  return {
    scene: currentScene(),
    buildVersion: GAME_VERSION,
    screen: currentScreen(),
    errors: [...errors],
  };
}
