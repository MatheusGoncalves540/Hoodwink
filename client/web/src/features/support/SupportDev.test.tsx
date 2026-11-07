import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { FEATURES } from "../../config/features";
import { SupportDev } from "./SupportDev";
import { watchSupportVideo } from "./supportVideo";

describe("Apoie o dev", () => {
  it("fica desligado por enquanto: sem a chave, não aparece nada", () => {
    expect(FEATURES.supportVideo).toBe(false);
    expect(renderToStaticMarkup(<SupportDev layout="mobile" />)).toBe("");
  });

  it("ligado, mostra só o link discreto (a gaveta abre no toque)", () => {
    const out = renderToStaticMarkup(<SupportDev layout="mobile" enabled />);
    expect(out).toContain("Apoie o dev · assista a um vídeo");
    expect(out).not.toContain("Assistir ao vídeo");
  });

  it("sem provedor de anúncios, o vídeo responde indisponível", async () => {
    expect(await watchSupportVideo()).toBe(false);
  });
});
