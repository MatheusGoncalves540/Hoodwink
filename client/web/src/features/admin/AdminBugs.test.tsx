import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import type { AdminBugReport } from "../../net/adminApi";
import { Detail } from "./AdminBugs";

const report: AdminBugReport = {
  id: 9,
  created_at: "2026-11-06T12:00:00Z",
  updated_at: "2026-11-06T12:05:00Z",
  user_id: "6f1c1d3e-2a4b-4c5d-8e9f-0123456789ab",
  username: "ana",
  source: "partida",
  match_id: "01936f5e-7a2b-7c3d-8e9f-0123456789ab",
  turns: [3, 4],
  category: "moedas",
  description: "<img src=x onerror=alert(1)>",
  context: { build: "abc", errors: [] },
  ip: "127.0.0.1",
  user_agent: "Mozilla/5.0",
  status: "novo",
};

const html = (item: AdminBugReport, saving: Parameters<typeof Detail>[0]["saving"] = null) =>
  renderToStaticMarkup(
    <Detail item={item} saving={saving} saveError={null} notice={null} onStatus={() => {}} />,
  );

describe("AdminBugs Detail", () => {
  it("mostra os campos do relato e escapa descrição/contexto como texto", () => {
    const out = html(report);

    expect(out).toContain("Moedas erradas");
    expect(out).toContain("3, 4");
    expect(out).toContain("01936f5e-7a2b-7c3d-8e9f-0123456789ab");
    expect(out).toContain("&lt;img src=x onerror=alert(1)&gt;");
    expect(out).not.toContain("<img");
    expect(out).toContain("&quot;build&quot;: &quot;abc&quot;");
    expect(out).toContain("127.0.0.1");
    expect(out).toContain("Mozilla/5.0");
    expect(out).toContain("Marcar como resolvido");
  });

  it("usa os textos de apoio quando o relato é anônimo e sem contexto", () => {
    const out = html({
      ...report,
      user_id: null,
      username: null,
      match_id: null,
      turns: [],
      category: null,
      description: null,
      context: null,
      ip: null,
      user_agent: null,
    });

    expect(out).toContain("Anônimo");
    expect(out).toContain("Sem descrição.");
    expect(out).toContain("Sem contexto.");
    expect(out).toContain("Sem User-Agent.");
    expect(out).toContain("Sem IP.");
    expect(out).toContain("Sem categoria");
  });

  it("desabilita os botões de status enquanto salva", () => {
    const out = html(report, "visto");

    expect(out).toContain("Salvando…");
    expect(out.match(/aria-disabled="true"/g)?.length).toBeGreaterThanOrEqual(3);
  });
});
