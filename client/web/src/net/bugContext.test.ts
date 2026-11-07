import { afterAll, beforeEach, describe, expect, it, vi } from "vitest";
import { collectBugContext, installBugContext, resetBugContext, sanitizeBugText } from "./bugContext";
import { GAME_VERSION } from "../version";

const handlers = new Map<string, (event: unknown) => void>();

const fakeWindow = {
  innerWidth: 390,
  innerHeight: 844,
  location: { hash: "", search: "" },
  addEventListener(type: string, handler: (event: unknown) => void) {
    handlers.set(type, handler);
  },
};

const fireError = (error: unknown, extra: Record<string, unknown> = {}) =>
  handlers.get("error")!({ error, message: "", filename: "", lineno: 0, colno: 0, ...extra });

const fireRejection = (reason: unknown) => handlers.get("unhandledrejection")!({ reason });

describe("contexto do relato de bug", () => {
  vi.stubGlobal("window", fakeWindow);
  installBugContext();

  beforeEach(() => {
    resetBugContext();
    fakeWindow.location.hash = "";
    fakeWindow.location.search = "";
    fakeWindow.innerWidth = 390;
    fakeWindow.innerHeight = 844;
  });

  afterAll(() => {
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
  });

  it("tira a query e o hash de URLs", () => {
    fireError(new Error("falhou em http://localhost:4444/game?Ticket=abc.def#frag"));

    const error = collectBugContext()!.errors[0]!;
    expect(error).toContain("http://localhost:4444/game");
    expect(error).not.toContain("Ticket");
    expect(error).not.toContain("abc.def");
    expect(error).not.toContain("#frag");
  });

  it("tira query de caminho relativo e de URL sozinha", () => {
    expect(sanitizeBugText("em /sala/K7Q2?Ticket=abc")).toBe("em /sala/K7Q2");
    expect(sanitizeBugText("?Ticket=abc")).not.toContain("Ticket");
  });

  it("apaga Bearer, token e ticket da mensagem e do stack", () => {
    fireError(new Error("fail Bearer abc.def.ghi ticket=xyz token=senha-secreta"));

    const error = collectBugContext()!.errors[0]!;
    expect(error).not.toContain("abc.def.ghi");
    expect(error).not.toContain("xyz");
    expect(error).not.toContain("senha-secreta");
    expect(error).toContain("Bearer ***");
    expect(error).toContain("ticket=***");
    expect(error).toContain("token=***");
  });

  it("guarda os últimos 10 erros, do mais antigo para o mais novo", () => {
    for (let i = 0; i < 12; i++) fireError(new Error(`erro ${i}`));

    const list = collectBugContext()!.errors;
    expect(list).toHaveLength(10);
    expect(list[0]).toContain("erro 2");
    expect(list[9]).toContain("erro 11");
  });

  it("ignora erro de carregamento sem mensagem", () => {
    fireError(undefined, { message: "" });

    expect(collectBugContext()!.errors).toEqual([]);
  });

  it("corta cada erro em 500 caracteres", () => {
    fireError(new Error("x".repeat(900)));

    expect(collectBugContext()!.errors[0]).toHaveLength(500);
  });

  it("captura promessa rejeitada", () => {
    fireRejection(new Error("caiu Bearer zzz"));

    const error = collectBugContext()!.errors[0]!;
    expect(error).toContain("caiu");
    expect(error).not.toContain("zzz");
  });

  it("devolve cena, versão do build, tela e erros", () => {
    fakeWindow.location.hash = "#admin";

    expect(collectBugContext()).toEqual({
      scene: "admin",
      buildVersion: GAME_VERSION,
      screen: "390x844",
      errors: [],
    });
  });

  it("reconhece a sala pela query e usa a âncora quando existe", () => {
    fakeWindow.location.search = "?sala=K7Q2";
    expect(collectBugContext()!.scene).toBe("sala");

    fakeWindow.location.hash = "#camarim";
    expect(collectBugContext()!.scene).toBe("camarim");
  });

  it("permite desligar a coleta", () => {
    fireError(new Error("não deve vazar"));

    expect(collectBugContext(false)).toBeNull();
  });
});
