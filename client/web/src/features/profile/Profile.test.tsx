import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import type { MyMatch } from "../../net/replayApi";
import {
  ProfileView,
  describePlayers,
  describeWinner,
  formatMatchDate,
  playerLabel,
  type ProfileViewProps,
} from "./Profile";

const match: MyMatch = {
  match_id: "01936f5e-7a2b-7c3d-8e9f-0123456789ab",
  room_id: "sala-1",
  engine_version: 1,
  started_at: "2026-11-06T12:00:00Z",
  ended_at: "2026-11-06T12:20:00Z",
  winner_id: "me",
  player_ids: ["me", "01936f5e-7a2b-7c3d-8e9f-0123456789ab"],
};

const base: ProfileViewProps = {
  signedIn: true,
  name: "Matheus",
  userId: "me",
  matches: [],
  loading: false,
  loadingMore: false,
  hasMore: false,
  error: "",
  notice: "",
  busyId: "",
  onLogin() {},
  onLogout() {},
  onWatch() {},
  onDownload() {},
  onLoadMore() {},
  onRetry() {},
};

const html = (patch: Partial<ProfileViewProps> = {}) =>
  renderToStaticMarkup(<ProfileView {...base} {...patch} />);

describe("Perfil: estados", () => {
  it("logado fora mostra o convite para entrar", () => {
    const out = html({ signedIn: false });
    expect(out).toContain("Seu perfil");
    expect(out).toContain("Entre para ver seu perfil");
    expect(out).toContain("Entrar com Google");
    expect(out).not.toContain("Nenhuma partida ainda");
  });

  it("sem login não mostra o botão de sair da conta", () => {
    expect(html({ signedIn: false })).not.toContain("Sair da conta");
  });

  it("logado mostra o botão de sair da conta", () => {
    expect(html()).toContain("Sair da conta");
  });

  it("carregando anuncia o estado", () => {
    const out = html({ loading: true });
    expect(out).toContain("Carregando suas partidas…");
    expect(out).toContain('role="status"');
    expect(out).toContain('aria-busy="true"');
  });

  it("sem partidas mostra o vazio", () => {
    const out = html();
    expect(out).toContain("Matheus");
    expect(out).toContain("Nenhuma partida ainda");
    expect(out).toMatch(/>M<\/span>/);
  });

  it("com partidas mostra data, jogadores, vencedor e as duas ações", () => {
    const out = html({ matches: [match] });
    const date = formatMatchDate(match.ended_at);

    expect(out).toContain(date);
    expect(out).toMatch(new RegExp(`date[Tt]ime="${match.ended_at}"`));
    expect(out).toContain("Você venceu");
    expect(out).toContain("Jogadores: Você e Jogador 0193");
    expect(out).toContain(">Assistir<");
    expect(out).toContain(">Baixar<");
    expect(out).toContain(`aria-label="Assistir à partida de ${date}"`);
    expect(out).toContain(`aria-label="Baixar o replay da partida de ${date}"`);
  });

  it("erro sem partidas mostra o aviso e o botão de tentar de novo", () => {
    const out = html({ error: "Sem conexão." });
    expect(out).toContain("Não deu para carregar suas partidas");
    expect(out).toContain("Sem conexão.");
    expect(out).toContain("Tentar de novo");
    expect(out).toContain('role="alert"');
  });

  it("erro com partidas mantém a lista e avisa em cima", () => {
    const out = html({ matches: [match], error: "Sem conexão." });
    expect(out).toContain("Assistir");
    expect(out).toContain("Sem conexão.");
  });

  it("carregar mais aparece só quando há mais e desabilita enquanto busca", () => {
    expect(html({ matches: [match], hasMore: true })).toContain("Carregar mais");

    const out = html({ matches: [match], hasMore: true, loadingMore: true });
    expect(out).toContain("Carregando mais…");
    expect(out).toContain('aria-disabled="true"');
  });

  it("nome do usuário entra como texto, nunca como HTML", () => {
    const out = html({ name: '<img src=x onerror=alert(1)>' });
    expect(out).toContain("&lt;img src=x onerror=alert(1)&gt;");
    expect(out).not.toContain("<img");
  });
});

describe("Perfil: rótulos", () => {
  it("dá nome curto ao jogador e destaca o usuário atual", () => {
    expect(playerLabel("me", "me")).toBe("Você");
    expect(playerLabel(match.match_id, "me")).toBe("Jogador 0193");
    expect(playerLabel("", "me")).toBe("Jogador ?");
  });

  it("junta os jogadores com 'e' antes do último", () => {
    expect(describePlayers([], "me")).toBe("Jogadores não informados");
    expect(describePlayers(["me"], "me")).toBe("Você");
    expect(describePlayers(["me", match.match_id], "me")).toBe("Você e Jogador 0193");
  });

  it("descreve o vencedor", () => {
    expect(describeWinner("me", "me")).toBe("Você venceu");
    expect(describeWinner(match.match_id, "me")).toBe("Jogador 0193 venceu");
    expect(describeWinner(null, "me")).toBe("Sem vencedor");
  });

  it("formata a data em PT-BR curto e protege data inválida", () => {
    expect(formatMatchDate(match.ended_at)).toMatch(/^\d{2}\/\d{2}\/2026$/);
    expect(formatMatchDate("não é data")).toBe("Data não informada");
  });
});
