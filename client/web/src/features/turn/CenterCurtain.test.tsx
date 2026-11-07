import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { CenterCurtain, curtainState, swipeMove } from "./CenterCurtain";

const html = (node: React.ReactElement) => renderToStaticMarkup(node);

describe("CenterCurtain", () => {
  it("a máquina de estado só abre, só fecha ou inverte", () => {
    expect(curtainState(false, "open")).toBe(true);
    expect(curtainState(true, "open")).toBe(true);
    expect(curtainState(true, "close")).toBe(false);
    expect(curtainState(false, "close")).toBe(false);
    expect(curtainState(false, "toggle")).toBe(true);
    expect(curtainState(true, "toggle")).toBe(false);
  });

  it("arrastar o handle: para cima abre, para baixo fecha, pouco movimento é toque", () => {
    expect(swipeMove(-80)).toBe("open");
    expect(swipeMove(60)).toBe("close");
    expect(swipeMove(-10)).toBeNull();
    expect(swipeMove(29)).toBeNull();
  });

  it("com peek o handle vira o cabeçalho e mantém o nome acessível", () => {
    const out = html(
      <CenterCurtain open={false} onToggle={() => {}} peek={<span>Escolha uma ação</span>}>
        <p>ações do turno</p>
      </CenterCurtain>,
    );
    expect(out).toContain('aria-label="Abrir ações"');
    expect(out).toContain("Escolha uma ação");
  });

  it("fechada deixa só o handle e esconde o conteúdo da árvore de acessibilidade", () => {
    const out = html(
      <CenterCurtain open={false} onToggle={() => {}}>
        <p>ações do turno</p>
      </CenterCurtain>,
    );
    expect(out).toContain('aria-expanded="false"');
    expect(out).toContain("aria-controls=");
    expect(out).toContain('aria-hidden="true"');
    expect(out).toContain("Abrir ações");
    // o conteúdo continua no DOM para a transição, mas fora do foco (inert entra no efeito)
    expect(out).toContain("ações do turno");
    expect(out).not.toContain("Fechar ações");
  });

  it("aberta anuncia o estado e oferece fechar", () => {
    const out = html(
      <CenterCurtain open onToggle={() => {}}>
        <p>ações do turno</p>
      </CenterCurtain>,
    );
    		expect(out).toContain('aria-expanded="true"');
    		expect(out).toContain("Fechar ações");
    		// o corpo aberto não carrega aria-hidden (só o grip decorativo)
    		expect(out).toMatch(/<div id="[^"]*" class="[^"]*body[^"]*">/);
  });

  it("sem onToggle não recolhe: durante o alvo não há handle", () => {
    const out = html(
      <CenterCurtain open>
        <p>escolha o alvo</p>
      </CenterCurtain>,
    );
    expect(out).toContain("escolha o alvo");
    expect(out).not.toContain("aria-expanded");
    expect(out).not.toContain("<button");
  });
});
