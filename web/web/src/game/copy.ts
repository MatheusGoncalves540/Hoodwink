/**
 * Todo texto de interface do Hoodwink. Frases com variável são funções.
 * Voz: 01-produto-e-voz.md (caixa de frase, vocabulário de palco, "?" para a dúvida, sem "Blefe!").
 */
import type { Price } from "./pricing";
import type { RoleId } from "./protocol";

const plural = (n: number, one: string, many: string) => `${n} ${n === 1 ? one : many}`;

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
  seatsRegion: "Lugares na mesa, toque para ver detalhes",
  seatLabel: (name: string, coins: number, out: boolean) =>
    `Ver o lugar de ${name}${out ? ", fora do jogo" : `, ${coins} moedas`}`,
  stage: "Palco",
  proso: "Proso",
  footlights: "Ribalta: estado da mesa e efeitos ativos",
  tableGroup: "Mesa",
  effectsGroup: "Efeitos ativos",
  deck: "baralho",
  deckLabel: (n: number) => `${plural(n, "carta", "cartas")} no baralho`,
  deckEmpty: "baralho vazio",
  deckEmptyLabel: (s: number) => `Baralho vazio, morte súbita: cada vez dura ${s} segundos`,
  tax: "taxa",
  taxLabel: (t: number) => `Taxa atual: ${t}`,
  effectLabel: (role: string, times: number, rounds: number) =>
    `${role} custa ${2 ** times} vezes mais. Baixa em ${plural(rounds, "rodada", "rodadas")}`,

  // faixa Você
  youRegion: "Você",
  limit: (n: number) => `limite ${n}`,
  investments: (n: number, next: number | null) =>
    n === 0 ? "" : `${plural(n, "investimento", "investimentos")}${next != null ? `, rende em ${next}` : ""}`,
  turnOf: (name: string) => ({ before: "Vez de ", name, after: ". Se usar um personagem, você pode reagir." }),
  timerLabel: (name: string) => `Tempo da jogada de ${name}`,
  seconds: "s",

  // gavetas
  backToTable: "☜ Voltar à mesa",
  seatSheet: (name: string) => `Lugar de ${name}`,
  tableSaw: "O que a mesa já viu",
  nothingSeen: "Ainda nada. A mesa observa.",
  cardState: { dead: "morta", revealed: "revelada", protected: "protegida", hidden: "escondida" },
  inHandState: "na mão",
  yourBox: "Seu lugar",
  overLimit: (max: number) => `Passou de ${max} moedas, perde uma carta.`,
  tableSees: (what: "protected" | "back" | "revealed" | "dead") =>
    `a mesa vê: ${{ protected: "carta protegida", back: "carta virada", revealed: "carta revelada", dead: "carta morta" }[what]}`,
  investmentsGroup: "Investimentos",
  investmentsLine: (n: number) => `${plural(n, "investimento", "investimentos")}, cada um rende meia moeda por rodada`,
  endsIn: "acaba em",
  endsInLabel: (r: number) => `Acaba em ${plural(r, "rodada", "rodadas")}`,
  rolesSeen: "Personagens que a mesa já viu você usar",
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
  sceneLog: "Histórico da partida",

  // menu
  menuTitle: "Menu da sala",
  room: "Sala",
  roomCode: "código",
  chatShort: "Conversa",
  howToPlay: "Como se joga",
  rolesLink: "os 12 personagens",
  vibrationHint: "quando chega a sua vez",
  motionAll: "Animações",
  motionAllHint: "Todas, inclusive as transições.",
  motionIdle: "Animações de espera",
  motionIdleHint: "Desenhos tremendo, texto respirando, lâmpadas.",
  motionIdleOff: "Desligada junto com Animações",
  prefsTitle: "Preferências",
  menuWarning: (code: string) => `Seu lugar fica na mesa e o tempo corre sem você. Para voltar, entre de novo com o código ${code}.`,
  prefs: "Preferências",
  sound: "Som",
  vibration: "Vibração",
  yes: "sim",
  no: "não",
  leaveScene: "Sair da partida",
  leaveConfirm: "Quer mesmo sair? Suas cartas ficam na mesa.",
  leaveYes: "Sair",
  leaveNo: "Ficar",

  // turno
  yourTurnInScene: "Sua vez",
  declareRole: "Escolha uma ação",
  youHave: (n: number) => `Você tem ${n} moedas`,
  inHand: "☞ na mão",
  missing: (n: number) => `Faltam ${plural(n, "moeda", "moedas")}`,
  moveTime: "Tempo da jogada",
  left: "restam",
  chooseAction: "Escolher ação",
  roleArt: (name: string) => `Faixa dos olhos de ${name} (provisória)`,
  roleDesc: (id: string, p: Price) => {
    switch (id) {
      case "TAKE_COINS": return `Recebe ${plural(p.coins, "moeda", "moedas")}`;
      case "KILL_CARD": return "Mata uma carta de outro jogador";
      case "SWAP_CARDS": return "Troca uma carta sua por uma do baralho";
      case "TRILLIONAIRE": return `Recebe ${p.coins} moedas`;
      case "TRICKSTER": return p.kind === "gain" ? `Rouba ${p.steals} de alguém e fica com ${p.coins}` : "";
      case "INVESTOR": return p.kind === "invest" ? `Investe ${p.coins}. A moeda volta em ${p.lasts} turnos` : "";
      case "ASSASSIN": return "Mata uma carta de outro jogador";
      case "CLAIRVOYANT": return "Revela uma carta de alguém";
      case "KAMIKAZE": return "Sacrifica-se e leva uma carta junto";
      case "GUARDIAN": return "Protege uma carta sua";
      case "GRAVEDIGGER": return "Revive uma carta sua que morreu";
      case "CROUPIER": return "Troca uma carta sua por uma nova do baralho";
      case "POLITICAL": return "Sobe a taxa em 1. O preço dele dobra";
      case "REBEL": return "Baixa a taxa em 1. O preço dele dobra";
      case "SELFISH": return p.kind === "reaction" ? `Barra moedas por ${p.coins}, investimento por ${p.investment} ou troca de carta por ${p.cardChange}` : "";
      default: return "";
    }
  },
  priceLabel: (p: Price) => (p.kind === "gain" ? `+${p.coins}` : p.kind === "reaction" ? `${p.coins}+` : String(p.coins)),
  basicActions: "Ações do turno",
  needSwapCoins: (n: number) => `Precisa de ${n} moedas`,
  swapBoth: "Trocar as duas",
  swapBothWarn: "Custa todas as suas moedas.",
  showToAll: "Mostrar para todos",
  showToMe: "Desligado, só você vê a carta.",
  killHint: "Toque no lugar de quem perde a carta.",
  swapHint: "Escolha a carta sua que vai trocar.",
  pickTarget: "Escolha o alvo",
  pickTargetHint: (role: string) => `Toque no lugar de quem vai sofrer o ${role}.`,
  pickCard: (name: string, state: string) => `Escolher a carta ${state} de ${name}`,
  pickOwnCard: (role: string) => `Escolher ${role}`,
  actionCost: (name: string, cost: number) => `${name}, custa ${cost}`,
  actionGain: (name: string, coins: number) => `${name}, recebe ${plural(coins, "moeda", "moedas")}`,
  actionPoor: (name: string, cost: number, missing: number) => `${name}, custa ${cost}, faltam ${plural(missing, "moeda", "moedas")}`,
  actionOffTurn: (name: string) => `${name}, fora do seu turno`,
  backToActions: "☜ Voltar às ações",
  pickCardOf: (name: string) => `Escolha a carta de ${name}`,

  // reação e dúvida
  claims: (name: string, role: string) => `${name} se diz ${role}`,
  yourCue: "Sua vez de reagir",
  reactionLabel: "Reagir",
  reactionTimer: "Tempo para reagir",
  block: (role: string) => `Bloquear com ${role}`,
  contest: "Duvidar",
  letPass: "Deixar passar",
  stampBluff: "Blefe?",
  stampTruth: "Verdade",
  stampRevealed: "revelada",
  chooseFallingCard: "Escolha a carta que cai",
  bluffed: (name: string, role: string) => `${name} não tinha ${role}.`,
  hadCard: (name: string, role: string) => `${name} tinha ${role}.`,
  choosesYourCard: (name: string) => `${name} escolhe qual carta sua cai`,
  contestTimer: (s: number) => `${s} segundos para escolher`,
  declaredLead: (name: string) => `${name} se diz`,
  contestedLead: (name: string) => `${name} se disse`,
  truthLead: (name: string) => `${name} era mesmo`,
  declaredNotRevealed: "dita, não mostrada",
  cardArt: (role: string) => `arte do ${role}`,
  bluffFall: (name: string) => ` A jogada não acontece e ${name} perde uma carta.`,
  truthReturn: (name: string) => ` ${name} mostra a carta, devolve ao baralho e pega outra.`,
  hand: (names: string[]) => (names.length ? `Na mão: ${names.join(" e ")}` : "Na mão: nada"),
  blockCost: (cost: number) => `por ${plural(cost, "moeda", "moedas")}`,
  dropCard: (name: string) => `Derrube uma carta de ${name}`,
  dropRevealed: (name: string, role: string) => `Derrubar a carta revelada de ${name}, ${role}`,
  dropHidden: (name: string) => `Derrubar a carta escondida de ${name}`,
  cardsOf: (name: string) => `Cartas de ${name}`,
  yourCards: "Suas cartas",
  waitPickHint: "Depois disso a jogada dela ainda mira você.",
  protectedWarn: "se for esta, cai só a proteção",
  sorteEscolhe: (s: number) => `Em ${s} s, a sorte escolhe por você.`,
  sorteEscolheOther: (name: string, s: number) => `${name} tem ${s} s. Depois, a sorte escolhe.`,
  declaredEffect: (name: string, role: RoleId, price: Price, targetsMe: boolean) => {
    const target = targetsMe ? "uma carta sua" : "uma carta de outro jogador";
    switch (role) {
      case "ASSASSIN": return `${name} paga ${price.coins} moedas e mira ${target}.`;
      case "CLAIRVOYANT": return `${name} paga ${price.coins} moedas e revela ${target}.`;
      case "TRICKSTER": return `${name} rouba ${price.kind === "gain" ? price.steals ?? 0 : 0} de ${targetsMe ? "você" : "alguém"}.`;
      case "KAMIKAZE": return `${name} se sacrifica e leva ${target} junto.`;
      case "GUARDIAN": return `${name} paga ${price.coins} moedas e protege uma carta sua.`;
      case "GRAVEDIGGER": return `${name} paga ${price.coins} moedas e revive uma carta.`;
      case "CROUPIER": return `${name} paga ${price.coins} moedas e puxa uma carta nova.`;
      case "INVESTOR": return `${name} investe ${price.coins}. A moeda volta em ${price.kind === "invest" ? price.lasts : 0} turnos.`;
      case "TRILLIONAIRE": return `${name} recebe ${price.coins} moedas.`;
      case "POLITICAL": return `${name} sobe a taxa em 1. O preço do Político dobra.`;
      case "REBEL": return `${name} baixa a taxa em 1. O preço do Rebelde dobra.`;
      default: return "";
    }
  },

  // fim de jogo
  outOfScene: "Fora da partida",
  outOfSceneRegion: "Você, fora da partida",
  outOfSceneAt: (round: number) => `Suas cartas caíram na rodada ${round}. Você ainda vê a mesa e pode conversar.`,
  deadCardsLabel: (names: string[]) => `Suas duas cartas, mortas: ${names.join(" e ")}`,
  openChat: "Abrir a conversa",
  leaveRoom: "Sair da sala",
  deckOver: "O baralho acabou.",
  suddenDeath: "Morte súbita",
  suddenDeathLower: "morte súbita",
  roundSudden: (turn: number) => `Rodada ${turn},`,
  eachCue: (s: number) => `Cada vez dura ${s} s.`,
  suddenTurnAfter: (s: number) => `. Agora cada vez dura ${s} s.`,
  playResumes: (s: number) => `A partida volta em ${s} s`,
  suddenDeathYouLine: (s: number) => `Cada vez dura ${s} s. Decida rápido.`,
  finale: "Fim da partida",
  wonAtRound: (round: number, coins: number) => `Venceu na rodada ${round}, com ${coins} moedas.`,
  exitOrder: "Ordem de saída",
  place: (n: number) => `${n}º`,
  newRoom: "Criar outra sala",
  backToEntrance: "☜ Voltar à entrada",
  afterShow: "Depois da partida",

  // entrada e sala
  cartazRegion: "Cartaz",
  playRegion: "Jogar",
  castRegion: "Jogadores",
  inviteRegion: "Convite",
  startRegion: "Começar",
  createRoom: "Criar sala",
  roomCodePlaceholder: "Código da sala",
  profile: (name: string) => `Perfil de ${name}`,
  prosoPlaceholder: "me troque :c",
  prosoAlt: "Proso, o narrador. A arte chega em breve.",
  roomNotFound: (code: string) => `Não achamos a sala ${code}. Confira o código com quem convidou.`,
  backstage: "Bastidores, modo Clássico",
  roomNumber: "Sala nº",
  ready: "pronto",
  castLineLabel: "Jogadores: ",
  startHint: "Só o anfitrião começa, com pelo menos 2 jogadores.",
  houseRuleInitial: (coins: number) => `Cada jogador começa com ${plural(coins, "moeda", "moedas")}.`,
  houseRuleLimit: (max: number) => `Quem passar de ${max} perde uma carta.`,
  enter: "Entrar",
  nowShowing: "Em cartaz",
  epigraph: "Dê ao homem uma máscara, e ele dirá a verdade.",
  login: "Entrar com Google",
  castLine: (count: number, max: number) => `${count} de ${max} lugares ocupados`,
  vacantRole: "Lugar vago",
  waiting: "aguardando",
  host: "anfitrião",
  hostYou: "anfitrião, você",
  copyInvite: "Copiar convite",
  inviteCopied: "Convite copiado",
  openCurtains: "Abrir as cortinas",
  invalidCode: "Esse código não abre nenhuma sala. Confira as letras com quem convidou.",
  roomFull: "Essa sala está cheia. Peça outro convite ou crie a sua.",
  connectionLost: "A conexão caiu. Tentando voltar à sala…",

  // histórico (predicados: o jogador vem antes, no histórico geral)
  log: {
    declared: (role: string) => `se disse ${role}.`,
    noContest: " Ninguém duvidou.",
    contested: (role: string) => `se disse ${role}, alguém duvidou`,
    lostCard: " e perdeu a carta.",
    showedCard: " e mostrou a carta.",
    contestedBy: (target: string) => `duvidou de ${target}.`,
    lostACard: "perdeu uma carta.",
    revealed: (target: string) => `revelou uma carta de ${target} como Clarividente.`,
    revived: "reviveu uma carta como Coveiro.",
    changed: "puxou uma carta nova como Crupiê.",
    changedContested: "trocou a carta mostrada depois da dúvida.",
    earned: (n: number, role: string) => `recebeu ${n} como ${role}.`,
    tookCoins: (n: number) => `pegou ${plural(n, "moeda", "moedas")}.`,
    killed: (name: string) => `pagou e matou uma carta de ${name}.`,
    swapped: "trocou uma carta com o baralho.",
    greed: "passou de 20 moedas e perdeu uma carta.",
    stole: (n: number, target: string) => `roubou ${n} de ${target} como Trapaceiro.`,
    invested: "investiu como Investidor.",
    taxUp: "subiu a taxa como Político.",
    taxDown: "baixou a taxa como Rebelde.",
    denied: (target: string) => `barrou o ganho de ${target} como Egoísta.`,
    exited: "saiu da partida.",
    suddenDeath: "O baralho acabou. Morte súbita.",
  },

  // Como se joga (fase 7). Os números vêm de pricing.ts com a taxa da tela.
  howto: {
    title: "Como se joga",
    mode: "Modo Clássico",
    back: "☜ Voltar",
    motion: "Movimento",
    goalTitle: "O objetivo",
    goal: [
      "Cada jogador começa com 2 cartas viradas para baixo e 3 moedas.",
      "Cada carta é um personagem. Quando uma carta morre, ela fica virada para cima e não vale mais.",
      "Ganha quem ficar com carta viva por último.",
    ],
    bluff: "Você pode dizer que tem qualquer personagem, mesmo sem a carta. Isso é blefar.",
    turnTitle: "Na sua vez",
    turnLead: "Escolha 1 ação. As ações do turno funcionam até sem carta, e ninguém pode duvidar delas.",
    take: (n: number) => `Pegar moeda. Você recebe ${plural(n, "moeda", "moedas")}.`,
    kill: (n: number) => `Matar. Você paga ${n} e mata uma carta de outro jogador. A Guardiã pode impedir.`,
    swap: "Trocar cartas. Você paga metade das suas moedas e troca 1 carta pelo baralho. Precisa de 6 moedas. Trocar as 2 custa todas.",
    useTitle: "Usar um personagem",
    use: ["Diga qual personagem você usa e pague o preço dele.", "Os outros podem duvidar."],
    doubtTitle: "Duvidar",
    doubtHad: "Se ele tinha a carta, mostra a carta. Quem duvidou perde uma carta. A jogada acontece, e ele troca a carta mostrada por outra do baralho.",
    doubtNot: "Se ele não tinha, quem duvidou escolhe uma carta dele para morrer. A jogada não acontece.",
    taxTitle: "A taxa",
    tax: "A taxa vai de −5 a +5 e soma no preço da maioria das ações. Kamikaze, Político e Rebelde não sentem a taxa.",
    limit: "Quem passa de 20 moedas perde o que passou e uma carta.",
    taxTry: "Teste outra taxa",
    taxNow: (t: number) => `Taxa ${signed(t)}`,
    scribble: "os preços mudam aqui",
    rolesTitle: "Personagens",
    rolesPick: "Escolha um personagem",
    artSoon: "arte em breve!",
    price: (label: string) => `Preço: ${label}`,
    ficha: (id: RoleId, p: { cost: number; gain: number; steals: number; cap: number; threshold: number; cardChange: number; guard: [number, number, number] }) => {
      switch (id) {
        case "TRILLIONAIRE": return `Recebe ${p.gain} moedas. A Egoísta pode barrar se o ganho for ${p.threshold} ou mais.`;
        case "TRICKSTER": return `Tira ${p.steals} moedas de outro jogador e fica com ${p.gain}.`;
        case "INVESTOR": return `Investe 1 moeda por turno, até ${p.cap} no total. Em todo turno seu, rende metade do que está investido. Cada moeda volta para você em 6 turnos.`;
        case "ASSASSIN": return `Paga ${p.cost} e mata uma carta de outro jogador. A Guardiã pode impedir.`;
        case "CLAIRVOYANT": return `Paga ${p.cost} e mostra uma carta de outro jogador. Você escolhe se todos veem ou só você.`;
        case "KAMIKAZE": return "Não custa nada. Quando morre, leva junto uma carta de outro jogador.";
        case "GUARDIAN": return `Protege uma carta sua. Impedir um ataque custa ${p.guard[0]}. Contra o Clarividente, ${p.guard[1]}. No seu turno, ${p.guard[2]}, e protege do próximo ataque.`;
        case "GRAVEDIGGER": return `Paga ${p.cost} e revive uma carta sua que morreu. Não dá para usar com as 2 cartas vivas.`;
        case "CROUPIER": return `Paga ${p.cost}, descarta uma carta sua e puxa uma nova do baralho.`;
        case "POLITICAL": return `Paga ${p.cost} e sobe a taxa em 1. Cada uso dobra o preço para todos, até 8.`;
        case "REBEL": return `Paga ${p.cost} e baixa a taxa em 1. Cada uso dobra o preço para todos, até 8.`;
        case "SELFISH": return `Só fora do seu turno. Por 1, impede alguém de receber ${p.threshold} ou mais moedas. Por ${p.cardChange}, impede alguém de receber uma carta.`;
      }
    },
  },

  // Camarim (fase 8, plans/03-skins.md)
  camarim: {
    title: "Camarim",
    lead: "Escolha o visual do jogo. Toque numa peça para vestir.",
    preview: "Prévia",
    previewCast: "Elenco",
    previewButton: "Criar sala",
    worn: "✓ vestido",
    tabs: { pacotes: "Pacotes", cores: "Cores", letras: "Letras", molduras: "Molduras", botoes: "Botões", cartas: "Cartas" },
    tabLead: {
      pacotes: "Um visual pronto, com uma peça de cada tipo.",
      cores: "Muda as cores de todas as telas, no claro e no escuro.",
      letras: "Muda a letra dos títulos e dos rótulos.",
      molduras: "Muda o contorno do cartaz, da lista de jogadores e do fim da partida.",
      botoes: "Muda o botão principal e o secundário.",
      cartas: "Muda o verso das cartas.",
    },
  },
};
