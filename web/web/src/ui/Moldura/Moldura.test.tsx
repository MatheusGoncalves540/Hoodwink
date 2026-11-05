import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { Moldura } from "./Moldura";

describe("Moldura", () => {
  it("marca o lugar e põe os três quadros antes do conteúdo", () => {
    const out = renderToStaticMarkup(
      <Moldura lugar="elenco" as="section" aria-label="Elenco">
        <h1>Sala</h1>
      </Moldura>,
    );
    expect(out).toMatch(/^<section aria-label="Elenco" data-lugar="elenco" class="hw-moldura m-contorno">/);
    expect(out.match(/<i aria-hidden="true" class="hw-moldura-f m-f"><\/i>/g)).toHaveLength(3);
    expect(out).toContain("<h1>Sala</h1></section>");
  });
});
