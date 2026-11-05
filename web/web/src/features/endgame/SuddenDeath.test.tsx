import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";
import { fixture } from "../../game/fixtures";
import { Table } from "../table/Table";
import { SuddenDeathAlert } from "./SuddenDeathAlert";
import { SuddenDeathStage } from "./SuddenDeathStage";

const NOW = Date.UTC(2026, 10, 4, 21, 41);

describe("morte súbita", () => {
  it("MorteSubita: aviso, título e cortina em chamas", () => {
    const out = renderToStaticMarkup(<SuddenDeathAlert layout="mobile" seconds={5} />);
    expect(out).toContain('role="alert"');
    expect(out).toContain("O baralho acabou.");
    expect(out).toContain("Morte");
    expect(out).toContain("súbita");
    expect(out).toContain("Cada vez dura 5 s.");
    expect(out).toContain("A partida volta em 4 s");
    expect(out).toContain("ms-cortina");
  });

  it("MesaSubita: mesa escura, baralho vazio e o cenário com fogo", () => {
    const f = fixture("MesaSubita", NOW);
    const out = renderToStaticMarkup(
      <Table view={{ me: f.me, room: f.room, log: f.log }} dark slots={{ scenery: <SuddenDeathStage layout="mobile" /> }} />,
    );
    expect(out).toContain('data-theme="dark"');
    expect(out).toContain("morte súbita");
    expect(out).toContain("baralho vazio");
    expect(out).toContain("class=\"fogo\"");
    expect(out).toContain("class=\"brasa\"");
  });
});
