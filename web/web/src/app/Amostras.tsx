/**
 * Página de amostras (?cena=amostras), só para desenvolvimento: tokens nos dois temas, as três fontes,
 * as classes de components.css e motion.css. Cada primitivo de ui/ ganha uma seção aqui (T10–T13).
 */
import { useState, type CSSProperties } from "react";
import { Button } from "../ui/Button/Button";
import { Coin } from "../ui/Coin/Coin";
import { Label } from "../ui/Label/Label";
import { Marquee } from "../ui/Marquee/Marquee";
import { ProsoLine } from "../ui/ProsoLine/ProsoLine";
import { Stamp } from "../ui/Stamp/Stamp";
import { TextField } from "../ui/TextField/TextField";
import { Ticket } from "../ui/Ticket/Ticket";
import { Toggle } from "../ui/Toggle/Toggle";
import { useMotionPref } from "./useMotionPref";
import { useTheme, type Theme } from "./useTheme";

const CORES = [
  "plateia", "luz", "tinta", "tinta-suave", "titulo", "veludo", "sobre-veludo", "alerta", "ouro",
  "ouro-texto", "lampada", "filete", "carta", "carta-tinta", "carta-rotulo", "poca", "cone", "scrim",
];
const MOVIMENTOS = ["m-rise", "m-drop", "m-flip", "m-stamp", "m-hit", "m-bow", "m-hang", "m-slam"];

/** T10: primitivos de ui/ com todas as variantes. */
function Primitivos() {
  const [code, setCode] = useState("");
  const [on, setOn] = useState(true);
  return (
    <div className="am-fontes">
      <Label tone="ouro">ui · botões</Label>
      <Ticket size="l">Criar sala</Ticket>
      <Ticket>
        Bloquear com Guardiã <i>por 6 moedas</i>
      </Ticket>
      <Ticket disabled>Abrir as cortinas</Ticket>
      <div className="am-linha">
        <Button variant="danger">Contestar</Button>
        <Button variant="outline">Deixar passar</Button>
        <Button variant="link">☜ Voltar</Button>
        <Button variant="outline" disabled>
          Sem moedas
        </Button>
      </div>

      <Label tone="ouro">campo</Label>
      <TextField label="Código da sala" placeholder="Código da sala" maxLength={6} upper value={code} onChange={setCode} action="Entrar" />
      <TextField label="Mensagem" placeholder="Fale com a mesa" value="" onChange={() => {}} />

      <Label tone="ouro">moedas</Label>
      <div className="am-linha">
        {([22, 26, 34, 46, 50, 56, 60] as const).map((n) => (
          <Coin key={n} size={n} value={7} label={`${n} px, 7 moedas`} />
        ))}
        <Coin size={46} label="sem custo" />
      </div>

      <Label tone="ouro">rótulos</Label>
      <div className="am-linha">
        <Label>Rodada 4 · Taxa +1</Label>
        <Label tone="ouro">Em cartaz</Label>
        <Label tone="suave">Para ganhar moedas</Label>
      </div>

      <Marquee className="am-letreiro">Carla</Marquee>
      <ProsoLine text="Quem mente melhor fica em cena." />
      <ProsoLine text="" />

      <div className="am-linha">
        <span className="am-carimbo">
          <Stamp text="revelada" animate />
        </span>
        <span className="am-carimbo">
          <Stamp text="Blefe?" animate />
        </span>
        <span className="am-carimbo">
          <Stamp text="Verdade" />
        </span>
      </div>

      <div>
        <Toggle label="Som" value={on} onChange={setOn} />
        <Toggle label="Menos movimento" hint="cartas e cortinas sem animação" value={!on} onChange={(v) => setOn(!v)} />
      </div>
    </div>
  );
}

export function Amostras() {
  const { theme, override, setOverride } = useTheme();
  const { reduced, setReduced } = useMotionPref();
  const [take, setTake] = useState(0);
  const temas: [Theme | null, string][] = [[null, "Sistema"], ["light", "Matinê"], ["dark", "Estreia"]];

  return (
    <main className="am">
      <header className="am-barra">
        <h1 className="am-display">Amostras</h1>
        <span className="hw-rotulo">tema atual: {theme}</span>
        {temas.map(([t, nome]) => (
          <button key={nome} type="button" className="am-botao" aria-pressed={override === t} onClick={() => setOverride(t)}>
            {nome}
          </button>
        ))}
        <button type="button" className="am-botao" aria-pressed={reduced} onClick={() => setReduced(!reduced)}>
          Menos movimento
        </button>
        <button type="button" className="am-botao" onClick={() => setTake((n) => n + 1)}>
          Repetir movimentos
        </button>
      </header>

      <div className="am-paineis">
        {(["light", "dark"] as const).map((t) => (
          <section key={t} data-theme={t} className="am-painel">
            <h2 className="hw-rotulo">{t === "light" ? "Matinê" : "Estreia"}</h2>

            <div className="am-cores">
              {CORES.map((c) => (
                // ponytail: amostra de token por variável inline; a página é só de desenvolvimento.
                <span key={c} className="am-cor" style={{ "--c": `var(--hw-${c})` } as CSSProperties}>
                  {c}
                </span>
              ))}
            </div>

            <div className="am-fontes">
              <span className="am-display">Hoodwink</span>
              <span className="am-texto">Bodoni Moda: quem mente melhor fica em cena.</span>
              <span className="hw-rotulo">Josefin Sans · rótulo</span>
              <span className="hw-num am-texto">0123456789 · 20 moedas</span>
            </div>

            <div className="am-linha">
              <button type="button" className="ticket am-ticket">Entrar em cena</button>
              <span className="hw-letreiro am-letreiro">Sala K7Q2</span>
            </div>

            <div className="am-linha">
              <span className="am-verso" aria-label="verso da carta" role="img" />
              <span className="stub am-stub hw-rotulo">canhoto</span>
            </div>

            <div className="am-linha" key={take}>
              {MOVIMENTOS.map((m) => (
                <span key={m} className={`am-bloco ${m}`}>{m}</span>
              ))}
              <span className="am-bloco m-tick">m-tick</span>
            </div>

            <Primitivos />
          </section>
        ))}
      </div>
    </main>
  );
}
