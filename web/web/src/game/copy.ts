/**
 * Todo texto de interface do Hoodwink. Frases com variável são funções.
 * Voz: 01-produto-e-voz.md (caixa de frase, vocabulário de palco, "?" para a dúvida, sem "Blefe!").
 */
import type { Price } from "./pricing";

const plural = (n: number, one: string, many: string) => `${n} ${n === 1 ? one : many}`;
const NUM_WORDS = ["nenhum", "um", "dois", "três", "quatro", "cinco", "seis", "sete", "oito", "nove", "dez"];

export const firstName = (name: string) => name.trim().split(/\s+/)[0] ?? name;
export const capitalize = (s: string) => s.charAt(0).toUpperCase() + s.slice(1);
export const signed = (n: number) => (n > 0 ? `+${n}` : n < 0 ? `−${-n}` : "0");

export const copy = {
  brand: "Hoodwink",
  you: "Você",
  coins: (n: number) => plural(n, "moeda", "moedas"),

  // cabeçalho e mesa
  roundTax: (turn: number, tax: number) => `Rodada ${turn} · Taxa ${signed(tax)}`,
  menu: "Menu",
  chatButton: (unread: number) => (unread > 0 ? `Conversa da mesa, ${plural(unread, "nova", "novas")}` : "Conversa da mesa"),
  seatsRegion: "Camarotes, toque para ver detalhes",
  seatLabel: (name: string, coins: number, out: boolean) =>
    `Ver camarote de ${name}${out ? ", fora do jogo" : `, ${coins} moedas`}`,
  stage: "Palco",
  proso: "Proso",
  footlights: "Ribalta: estado da mesa e efeitos ativos",
  tableGroup: "Mesa",
  effectsGroup: "Efeitos ativos",
  deck: "baralho",
  deckLabel: (n: number) => `${plural(n, "carta", "cartas")} no baralho`,
  deckEmpty: "baralho vazio",
  deckEmptyLabel: (s: number) => `Baralho vazio, morte súbita: cada deixa dura ${s} segundos`,
  tax: "taxa",
  taxLabel: (t: number) => `Taxa atual: ${t}`,
  effectLabel: (role: string, times: number, rounds: number) =>
    `${role} custa ${2 ** times} vezes mais; baixa em ${plural(rounds, "rodada", "rodadas")}`,

  // faixa Você
  youRegion: "Você",
  limit: (n: number) => `limite ${n}`,
  investments: (n: number, next: number | null) =>
    n === 0 ? "" : `${plural(n, "investimento", "investimentos")}${next != null ? `, rende em ${next}` : ""}`,
  turnOf: (name: string) => ({ before: "Vez de ", name, after: ". Se declarar, você reage." }),
  timerLabel: (name: string) => `Tempo da jogada de ${name}`,
  seconds: "s",

  // gavetas
  backToTable: "☜ Voltar à mesa",
  seatSheet: (name: string) => `Camarote de ${name}`,
  tableSaw: "O que a mesa já viu",
  nothingSeen: "Ainda nada. A mesa observa.",
  cardState: { dead: "morta", revealed: "revelada", protected: "protegida", hidden: "escondida" },
  yourBox: "Seu camarote",
  overLimit: (max: number) => `Passou de ${max} moedas, perde uma carta.`,
  tableSees: (what: "protected" | "back" | "revealed" | "dead") =>
    `a mesa vê: ${{ protected: "carta protegida", back: "carta virada", revealed: "carta revelada", dead: "carta morta" }[what]}`,
  investmentsGroup: "Investimentos",
  investmentsLine: (n: number) => `${plural(n, "investimento", "investimentos")}, cada um rende meia moeda por rodada`,
  endsIn: "acaba em",
  endsInLabel: (r: number) => `Acaba em ${plural(r, "rodada", "rodadas")}`,
  rolesSeen: "Papéis que a mesa já viu você usar",
  roundN: (r: number) => `rodada ${r}`,

  // conversa
  chatTitle: "Conversa da mesa",
  chatKeeps: "ficam as 30 últimas",
  chatList: "Falas",
  chatEmpty: "Ninguém falou ainda. Quebre o silêncio.",
  yourLine: "Sua fala",
  yourLinePlaceholder: "Sua fala…",
  speak: "Falar",
  sayToTable: "Diga algo à mesa",
  send: "Enviar",

  // trilho
  sceneLog: "Diário da cena",

  // menu
  menuTitle: "Menu da sala",
  room: "Sala",
  roomCode: "código",
  howToPlay: "Como se joga",
  prefs: "Preferências",
  sound: "Som",
  vibration: "Vibração",
  lessMotion: "Menos movimento",
  yes: "sim",
  no: "não",
  leaveScene: "Sair de cena",
  leaveConfirm: "Sair de cena de verdade? Suas cartas ficam na mesa.",
  leaveYes: "Sair",
  leaveNo: "Ficar",

  // turno
  yourTurnInScene: "Sua vez em cena",
  declareRole: "Declare um papel",
  youHave: (n: number) => `Você tem ${n} moedas`,
  inHand: "☞ na mão",
  missing: (n: number) => `Faltam ${plural(n, "moeda", "moedas")}`,
  moveTime: "Tempo da jogada",
  left: "restam",
  chooseAction: "Escolher ação",
  roleArt: (name: string) => `Faixa dos olhos de ${name} (provisória)`,
  roleDesc: (id: string, p: Price) => {
    switch (id) {
      case "TRILLIONAIRE": return `Recebe ${p.coins} moedas`;
      case "TRICKSTER": return p.kind === "gain" ? `Rouba ${p.steals} de alguém e fica com ${p.coins}` : "";
      case "INVESTOR": return p.kind === "invest" ? `Investe ${p.coins}, que rende por ${p.lasts + 1} turnos` : "";
      case "ASSASSIN": return "Mata uma carta de outro ator";
      case "CLAIRVOYANT": return "Revela uma carta de alguém";
      case "KAMIKAZE": return "Sacrifica-se e leva uma carta junto";
      case "GUARDIAN": return "Protege uma carta sua";
      case "GRAVEDIGGER": return "Revive uma carta do baralho morto";
      case "CROUPIER": return "Puxa uma carta nova do baralho";
      case "POLITICAL": return "Sobe a taxa em 1; o preço dobra";
      case "REBEL": return "Baixa a taxa em 1; o preço dobra";
      case "SELFISH": return p.kind === "reaction" ? `Barra moedas por ${p.coins}, investimento por ${p.investment} ou troca de carta por ${p.cardChange}` : "";
      default: return "";
    }
  },
  priceLabel: (p: Price) => (p.kind === "gain" ? `+${p.coins}` : p.kind === "reaction" ? `${p.coins}+` : String(p.coins)),
  pickTarget: "Escolha o alvo",
  pickTargetHint: (role: string) => `Toque no camarote de quem recebe o ${role}.`,
  backToActions: "☜ Voltar às ações",
  pickCardOf: (name: string) => `Escolha a carta de ${name}`,

  // reação e contestação
  claims: (name: string, role: string) => `${name} se diz ${role}`,
  yourCue: "Sua deixa",
  block: (role: string) => `Bloquear com ${role}`,
  contest: "Contestar",
  letPass: "Deixar passar",
  stampBluff: "Blefe?",
  stampTruth: "Verdade",
  stampRevealed: "revelada",
  chooseFallingCard: "Escolha a carta que cai",
  bluffed: (name: string, role: string) => `${name} não tinha ${role}.`,
  hadCard: (name: string, role: string) => `${name} tinha ${role}.`,
  choosesYourCard: (name: string) => `${name} escolhe qual carta sua cai`,
  contestTimer: (s: number) => `${s} segundos para escolher`,

  // fim de jogo
  outOfScene: "Fora de cena",
  outOfSceneText: "Suas duas cartas caíram. A peça segue sem você; dá para assistir daqui.",
  deadCardsLabel: "Suas duas cartas, mortas",
  openChat: "Abrir a conversa",
  leaveRoom: "Sair da sala",
  deckOver: "O baralho acabou.",
  suddenDeath: "Morte súbita",
  playResumes: (s: number) => `A peça volta em ${s} s`,
  suddenDeathYouLine: (s: number) => `Cada deixa dura ${s} s. Decida rápido.`,
  finale: "Fim do espetáculo",
  wonAtRound: (round: number, coins: number) => `Venceu na rodada ${round}, com ${coins} moedas.`,
  exitOrder: "Ordem de saída",
  place: (n: number) => `${n}º`,
  newRoom: "Criar outra sala",
  backToEntrance: "Voltar à entrada",

  // entrada e sala
  createRoom: "Criar sala",
  roomCodePlaceholder: "Código da sala",
  enter: "Entrar",
  nowShowing: "Em cartaz",
  epigraph: "Dê ao homem uma máscara, e ele dirá a verdade.",
  login: "Entrar com Google",
  castLine: (count: number, max: number) => `${NUM_WORDS[count] ?? count} de ${NUM_WORDS[max] ?? max} papéis preenchidos`,
  vacantRole: "Papel vago",
  waiting: "aguardando",
  host: "anfitrião",
  hostYou: "anfitrião, você",
  copyInvite: "Copiar convite",
  inviteCopied: "Convite copiado",
  openCurtains: "Abrir as cortinas",
  invalidCode: "Esse código não abre nenhuma sala. Confira as letras com quem convidou.",
  roomFull: "Essa sala está cheia. Peça outro convite ou crie a sua.",
  connectionLost: "A conexão caiu. Tentando voltar à sala…",

  // diário (predicados: o ator vem antes, no diário geral)
  log: {
    declared: (role: string) => `se disse ${role}.`,
    noContest: " Ninguém contestou.",
    contested: (role: string) => `se disse ${role}, foi contestado`,
    lostCard: " e perdeu a carta.",
    showedCard: " e mostrou a carta.",
    contestedBy: (target: string) => `contestou ${target}.`,
    lostACard: "perdeu uma carta.",
    revealed: (target: string) => `revelou uma carta de ${target} como Clarividente.`,
    revived: "reviveu uma carta como Coveiro.",
    changed: "puxou uma carta nova como Crupiê.",
    earned: (n: number, role: string) => `recebeu ${n} como ${role}.`,
    stole: (n: number, target: string) => `roubou ${n} de ${target} como Trapaceiro.`,
    invested: "investiu como Investidor.",
    taxUp: "subiu a taxa como Político.",
    taxDown: "baixou a taxa como Rebelde.",
    denied: (target: string) => `barrou o ganho de ${target} como Egoísta.`,
    exited: "saiu de cena.",
    suddenDeath: "O baralho acabou. Morte súbita.",
  },
};
