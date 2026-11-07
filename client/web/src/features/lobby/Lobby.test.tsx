import { renderToStaticMarkup } from "react-dom/server";
import { beforeEach, describe, expect, it } from "vitest";
import { useAuthStore } from "../../store/authStore";
import { useSessionStore } from "../../store/sessionStore";
import { Lobby, Profile } from "./Lobby";

const html = (el: JSX.Element) => renderToStaticMarkup(el);

beforeEach(() => {
  useAuthStore.setState({ jwtToken: "", playerId: "", isLoginOpen: false });
  useSessionStore.getState().reset();
});

describe("Lobby (fumaça em Node)", () => {
  it("mostra o cartaz, criar sala, entrar pelo código e o link das regras", () => {
    const out = html(<Lobby />);
    expect(out).toContain('aria-label="Cartaz"');
    expect(out).toContain("Criar sala");
    expect(out).toContain("Código da sala");
    expect(out).toMatch(/>Entrar<\/button>/);
    expect(out).toContain("Como se joga");
  });

  it("sem login, o perfil vira Entrar com Google", () => {
    const out = html(<Lobby />);
    expect(out).toContain("Entrar com Google");
    expect(out).not.toContain("Perfil de");
  });

  it("com login, o avatar vira botão para abrir o perfil", () => {
    const out = html(<Profile signedIn name="Matheus" onLogin={() => {}} onProfile={() => {}} />);
    expect(out).toContain('aria-label="Abrir perfil"');
    expect(out).toMatch(/<button type="button" aria-label="Abrir perfil"/);
    expect(out).toMatch(/>M<\/button>/);
  });
});
