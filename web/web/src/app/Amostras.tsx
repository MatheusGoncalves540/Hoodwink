/**
 * Página de amostras (?cena=amostras), só para desenvolvimento: tokens nos dois temas, as três fontes,
 * as classes de components.css e motion.css. Cada primitivo de ui/ ganha uma seção aqui (T10–T13).
 */
import { useState, type CSSProperties } from "react";
import { Button } from "../ui/Button/Button";
import { ROLE_GROUPS } from "../game/roles";
import { CLASSIC_RULES } from "../game/rules.classic";
import type { RoleId } from "../game/protocol";
import { CardBack, CardDead, CardFace, CardHidden, CardMini } from "../ui/Card/Card";
import { Coin } from "../ui/Coin/Coin";
import { CountdownDots } from "../ui/CountdownDots/CountdownDots";
import { Timer } from "../ui/Timer/Timer";
import { Label } from "../ui/Label/Label";
import { Moldura } from "../ui/Moldura/Moldura";
import { ProsoLine } from "../ui/ProsoLine/ProsoLine";
import { Sheet } from "../ui/Sheet/Sheet";
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

/** Grupo do papel, como rótulo da faixa de baixo do catálogo. */
const g = (r: RoleId) => ROLE_GROUPS.find((x) => x.roles.includes(r))?.title;

/** T10–T12: primitivos de ui/ com todas as variantes. */
function Primitivos() {
  const [code, setCode] = useState("");
  const [on, setOn] = useState(true);
  const [t0] = useState(() => Date.now());
  const [sheetOpen, setSheetOpen] = useState(false);
  const [sheetMode, setSheetMode] = useState<"mobile" | "desktop">("mobile");
  const openSheet = (m: "mobile" | "desktop") => {
    setSheetMode(m);
    setSheetOpen(true);
  };
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

      <Moldura lugar="lobby" className="am-letreiro">Carla</Moldura>
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

      <Label tone="ouro">tempo</Label>
      <Timer expiresAt={t0 + 8000} total={10} label="Tempo da jogada de Carla" caption={<>Vez de <b>Carla</b>. Se declarar, você reage.</>} />
      <Timer expiresAt={t0 + 7000} total={10} label="Tempo para reagir" decision />
      <Timer expiresAt={t0 + 3000} total={5} label="Tempo da jogada, morte súbita" tone="alerta" />
      <CountdownDots seconds={6} total={10} />

      <Label tone="ouro">cartas: seis estados (Cartas.png)</Label>
      <div className="am-linha">
        <CardBack size="xl" />
        <CardHidden size="xl" />
        <CardFace size="xl" role="TRILLIONAIRE" />
        <CardFace size="xl" role="GUARDIAN" protected />
        <CardFace size="xl" role="ASSASSIN" stamp="revelada" />
        <CardDead size="xl" />
      </div>
      <div className="am-linha">
        {ROLE_GROUPS.flatMap((g) => g.roles).map((r) => (
          <CardFace key={r} size="xl" role={r} cost={r === "KAMIKAZE" ? null : CLASSIC_RULES.cards[r].value} bottom={g(r)} />
        ))}
      </div>
      <Label tone="ouro">no jogo: faixa Você (s), gaveta (m), contestação (l)</Label>
      <div className="am-linha">
        <CardFace size="s" role="GUARDIAN" protected bottom="protegida" />
        <CardFace size="s" role="TRILLIONAIRE" />
        <CardBack size="s" />
        <CardHidden size="s" />
        <CardDead size="s" role="ASSASSIN" />
      </div>
      <div className="am-grade">
        <CardFace size="m" role="GUARDIAN" protected top="a mesa vê: carta protegida" bottom="protegida" />
        <CardFace size="m" role="TRILLIONAIRE" top="a mesa vê: carta virada" bottom="escondida" />
        <CardFace size="l" role="CROUPIER" bottom="revelada" as="button" aria-label="Derrubar a carta revelada de Ana, Crupiê" />
        <CardHidden size="l" label="escondida" as="button" aria-label="Derrubar a carta escondida de Ana" />
      </div>
      <Label tone="ouro">miniaturas do camarote</Label>
      <div className="am-linha">
        {(["l", "s", "desk-l", "desk-s"] as const).map((z) =>
          (["back", "revealed", "protected", "dead"] as const).map((st) => <CardMini key={z + st} size={z} state={st} />),
        )}
      </div>

      <Label tone="ouro">gaveta</Label>
      <div className="am-linha">
        <Button variant="outline" onClick={() => openSheet("mobile")}>
          Abrir gaveta (mobile)
        </Button>
        <Button variant="outline" onClick={() => openSheet("desktop")}>
          Abrir trilho (desktop)
        </Button>
      </div>
      <Sheet open={sheetOpen} mode={sheetMode} label="Camarote de Bruno" onClose={() => setSheetOpen(false)}>
        <div className="am-linha">
          <span className="am-display">Bruno</span>
          <Coin size={56} value={2} label="2 moedas" />
        </div>
        <div className="am-grade">
          <CardDead size="m" role="ASSASSIN" />
          <CardHidden size="m" label="escondida" />
        </div>
        <Label tone="suave">O que a mesa já viu</Label>
      </Sheet>

      <div>
        <Toggle label="Som" value={on} onChange={setOn} />
        <Toggle label="Menos movimento" hint="cartas e cortinas sem animação" value={!on} onChange={(v) => setOn(!v)} />
      </div>
    </div>
  );
}

export function Amostras() {
  const { theme, override, setOverride } = useTheme();
  const motion = useMotionPref();
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
        <button type="button" className="am-botao" aria-pressed={motion.all} onClick={() => motion.setAll(!motion.all)}>
          Animações
        </button>
        <button type="button" className="am-botao" aria-pressed={motion.idle} onClick={() => motion.setIdle(!motion.idle)}>
          Animações de espera
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
              <span className="am-texto">Literata: quem mente melhor fica na mesa.</span>
              <span className="hw-rotulo">Josefin Sans · rótulo</span>
              <span className="hw-num am-texto">0123456789 · 20 moedas</span>
            </div>

            <div className="am-linha">
              <button type="button" className="ticket am-ticket">Entrar em cena</button>
              <Moldura lugar="aviso" className="am-letreiro">Sala K7Q2</Moldura>
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
