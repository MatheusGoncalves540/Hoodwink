import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { Coin } from "../ui/Coin/Coin";
import { ProsoLine } from "../ui/ProsoLine/ProsoLine";
import { Ticket } from "../ui/Ticket/Ticket";
import { Toggle } from "../ui/Toggle/Toggle";
import { Amostras } from "./Amostras";

const html = (el: JSX.Element) => renderToStaticMarkup(el);

describe("primitivos de ui/ (fumaça em Node)", () => {
  it("a página de amostras renderiza os dois temas", () => {
    const out = html(<Amostras />);
    expect(out).toContain('data-theme="light"');
    expect(out).toContain('data-theme="dark"');
  });

  it("ticket desabilitado continua botão focável, com aria-disabled", () => {
    const out = html(<Ticket disabled>Abrir as cortinas</Ticket>);
    expect(out).toMatch(/^<button type="button" class="ticket [^"]*" aria-disabled="true">Abrir as cortinas<\/button>$/);
  });

  it("toggle é um switch com o estado anunciado", () => {
    expect(html(<Toggle label="Som" value={false} onChange={() => {}} />)).toContain('role="switch" aria-checked="false"');
  });

  it("moeda com rótulo é imagem acessível; sem valor vira losango", () => {
    expect(html(<Coin size={26} value={7} label="7 moedas" />)).toContain('role="img" aria-label="7 moedas">7<');
    expect(html(<Coin size={46} />)).toContain('aria-hidden="true"');
  });

  it("linha do Proso some sem fala", () => {
    expect(html(<ProsoLine text="" />)).toBe("");
    expect(html(<ProsoLine text="Luz." />)).toContain("“Luz.” <span");
  });
});
