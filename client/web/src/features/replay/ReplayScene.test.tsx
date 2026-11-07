import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { ReplayScene } from "./ReplayScene";

describe("ReplayScene sem replay", () => {
  it("mostra o estado vazio e o seletor de .hwr", () => {
    const html = renderToStaticMarkup(<ReplayScene />);
    expect(html).toContain("Nenhum replay aberto");
    expect(html).toContain("Abrir um arquivo de replay do disco");
    expect(html).toContain('type="file"');
    expect(html).toContain('accept=".hwr,application/octet-stream"');
  });
});
