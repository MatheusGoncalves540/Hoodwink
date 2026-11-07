import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it, vi } from "vitest";
import { CardFace, CardMini } from "../ui/Card/Card";
import { Coin } from "../ui/Coin/Coin";
import { CountdownDots } from "../ui/CountdownDots/CountdownDots";
import { Timer } from "../ui/Timer/Timer";
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

  it("timer: segundos restantes no número e na barra; barra parte da fração que falta", () => {
    vi.useFakeTimers({ now: 1_000_000 });
    const out = html(<Timer expiresAt={1_007_500} total={10} label="Tempo para reagir" decision />);
    vi.useRealTimers();
    expect(out).toContain('role="progressbar" aria-label="Tempo para reagir" aria-valuemin="0" aria-valuemax="10" aria-valuenow="8"');
    expect(out).toContain("transform:scaleX(0.75)");
    // o "?" virou tira de 3 desenhos (fase 4); o número continua parado
    expect(out).toMatch(/restam <b class="hw-num">8<span class="m-tira [^"]*"><span><i>\?<\/i><i>\?<\/i><i>\?<\/i><\/span><\/span><\/b>/);
  });

  it("timer vencido para em zero", () => {
    vi.useFakeTimers({ now: 2_000_000 });
    const out = html(<Timer expiresAt={1_000_000} total={5} label="t" />);
    vi.useRealTimers();
    expect(out).toContain('aria-valuenow="0"');
    expect(out).toContain("transform:scaleX(0)");
  });

  it("pontos da contestação: 6 cheios de 10, o sexto pisca", () => {
    const out = html(<CountdownDots seconds={6} total={10} />);
    expect(out.match(/<span class="[^"]*"/g)?.map((c) => c.includes("m-tick"))).toEqual([false, false, false, false, false, true, false, false, false, false]);
  });

  it("carta Cortina: faixa de baixo primeiro, carimbo no meio, nome por último (column-reverse)", () => {
    const out = html(<CardFace size="s" role="ASSASSIN" bottom="protegida" stamp="revelada" protected />);
    const order = [...out.matchAll(/(protegida|data-carimbo|Assassino)/g)].map((m) => m[1]);
    expect(order).toEqual(["protegida", "data-carimbo", "Assassino"]);
    expect(out).toMatch(/^<div class="hw-carta [^"]*protected/);
  });

  it("carta sem custo desenha o losango no lugar da moeda", () => {
    const kamikaze = html(<CardFace size="xl" role="KAMIKAZE" cost={null} />);
    expect(kamikaze).toMatch(/class="[^"]*diamond[^"]*"/);
    expect(kamikaze).not.toMatch(/class="hw-num/);
    expect(html(<CardFace size="xl" role="ASSASSIN" cost={3} />)).toMatch(/class="hw-num [^"]*">3</);
  });

  it("miniatura é decorativa", () => {
    expect(html(<CardMini size="l" state="dead" />)).toMatch(/^<span aria-hidden="true" class="[^"]*mini-dead/);
  });

  it("linha do Proso some sem fala", () => {
    expect(html(<ProsoLine text="" />)).toBe("");
    expect(html(<ProsoLine text="Luz." />)).toContain("“Luz.” <span");
  });
});
