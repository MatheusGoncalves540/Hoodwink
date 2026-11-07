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

/** Categorias do relato de bug. A mesma lista existe no handler e no check do banco. */
export const BUG_CATEGORIES = ["moedas", "carta", "travou", "regra", "tela", "outro"] as const;

export type BugCategory = (typeof BUG_CATEGORIES)[number];

export const BUG_CATEGORY_LABELS: Record<BugCategory, string> = {
  moedas: "Moedas erradas",
  carta: "Problema com carta",
  travou: "O jogo travou",
  regra: "Regra confusa",
  tela: "Tela quebrada",
  outro: "Outra coisa",
};

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
    `${name}${out ? ", fora do jogo" : `, ${coins} moedas`}. Ver o lugar`,
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
  versionLabel: (v: string) => `Versão ${v}`,
  menuWarning: (code: string) => `Seu lugar fica na mesa. Para voltar, use o código ${code}.`,
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
  missing: (n: number) => `Faltam ${plural(n, "moeda", "moedas")}`,
  moveTime: "Tempo da jogada",
  left: "restam",
  chooseAction: "Escolher ação",
  curtainOpen: "Abrir ações",
  curtainClose: "Fechar ações",
  yourTurnKicker: "Sua vez",
  seeStage: "Ver palco",
  openShort: "Abrir",
  heldTitle: "Na sua mão",
  heldHint: "Jogar estas não é blefe.",
  panelCards: "Suas cartas",
  roleArt: (name: string) => `Faixa dos olhos de ${name} (provisória)`,
  baseValue: (value: number) => `Valor base: ${value}`,
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
  needKamikaze: "Precisa de um Kamikaze vivo na mão",
  swapBoth: "Trocar as duas",
  swapBothWarn: "Custa todas as suas moedas.",
  revealFor: "Revelar para",
  revealQuestion: "Revelar para quem?",
  revealStep: "Clarividente · etapa 1 de 3",
  revealContext: "Você vai espiar a carta de outro ator.",
  revealEveryoneTag: "Para a plateia",
  revealOnlyMeTag: "Na coxia",
  revealEveryone: "Todos",
  revealEveryoneHint: "A mesa toda vê a carta.",
  revealOnlyMe: "Apenas você",
  revealOnlyMeHint: "Só você vê a carta.",
  revealNote: (all: boolean) => (all ? "A carta vai aparecer para todos." : "A carta vai aparecer só para você."),
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
  // escolha de alvo no holofote (TargetPrompt): uma pergunta por etapa
  targetStep: (role: string, n: number, total: number) => `${role} · etapa ${n} de ${total}`,
  targetWho: {
    ASSASSIN: "Quem vai sofrer?",
    CLAIRVOYANT: "De quem é a carta?",
    KAMIKAZE: "Quem vai junto?",
    KILL_CARD: "Quem perde a carta?",
    TRICKSTER: "De quem roubar?",
  } as Partial<Record<string, string>>,
  targetWhoFallback: "Quem é o alvo?",
  targetWhich: (name: string) => `Qual carta de ${name}?`,
  targetTapSeat: "Toque num camarote iluminado.",
  targetTapCard: "Toque numa das cartas vivas.",
  changeTarget: "☜ Trocar o alvo",

  // reação e dúvida
  claims: (name: string, role: string) => `${name} se diz ${role}`,
  yourCue: "Sua vez de reagir",
  reactionLabel: "Reagir",
  reactionTimer: "Tempo para reagir",
  block: (role: string) => `Bloquear com ${role}`,
  kamikazeStrike: "Levar uma carta junto com o Kamikaze",
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
  /** "uma carta sua" / "uma carta de Ana", para as frases com alvo. */
  targetText: (target: { name: string; mine: boolean } | null) =>
    target ? (target.mine ? "uma carta sua" : `uma carta de ${target.name}`) : "uma carta de outro jogador",
  /** "sua" / "de Ana", para "qual carta … cai". */
  possessive: (name: string, mine: boolean) => (mine ? "sua" : `de ${name}`),
  declaredEffect: (name: string, role: RoleId, price: Price, target: { name: string; mine: boolean } | null) => {
    const card = copy.targetText(target);
    const who = target ? (target.mine ? "você" : target.name) : "alguém";
    switch (role) {
      case "ASSASSIN": return `${name} paga ${price.coins} moedas e mira ${card}.`;
      case "CLAIRVOYANT": return `${name} paga ${price.coins} moedas e revela ${card}.`;
      case "TRICKSTER": return `${name} rouba ${price.kind === "gain" ? price.steals ?? 0 : 0} de ${who}.`;
      case "KAMIKAZE": return `${name} se sacrifica e leva ${card} junto.`;
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

  // contestação: os dois lados sempre com nome
  doubtedBy: (contester: string) => `${contester} duvidou.`,
  bluffEffect: (contested: string, role: string, chooser: string, poss: string) =>
    `${contested} não tinha ${role}. ${chooser} escolhe qual carta ${poss} cai.`,
  truthEffect: (contested: string, chooser: string, poss: string) =>
    `${contested} mostrou a carta e escolhe qual carta ${poss} cai.`,

  // carta revelada (Clarividente)
  revealedLead: (name: string, poss: string) => `${name} revelou uma carta ${poss}.`,
  revealedCard: (role: string) => `É ${role}.`,
  revealedPrivate: (name: string) => `Só ${name} viu a carta.`,
  revealedPrivateMe: "Só você viu a carta.",
  revealedToAll: "A mesa toda viu a carta.",
  revealedHidden: "A carta continua escondida para você.",

  /** Narração de um acontecimento, com os dois lados: alimenta o palco e o texto do relógio. */
  stageLine: {
    turn: (name: string) => `Vez de ${name}. Escolha uma ação.`,
    yourTurn: "Sua vez. Escolha uma ação.",
    declared: (name: string, role: string) => `${name} se diz ${role}. Os outros podem duvidar.`,
    doubted: (contester: string, contested: string) => `${contester} duvidou de ${contested}.`,
    chooseCardOf: (chooser: string, poss: string) => `${chooser} escolhe qual carta ${poss} cai.`,
    revealed: (name: string, poss: string) => `${name} revelou uma carta ${poss}.`,
    revealedCard: (role: string, poss: string) => `${role}, carta ${poss}.`,
    revealedPrivateMe: (role: string, poss: string) => `Só você viu: ${role}, carta ${poss}.`,
    killed: (name: string, card: string) => `${name} matou ${card}.`,
    kamikaze: (name: string, card: string) => `${name} levou ${card} junto com o Kamikaze.`,
    lostCard: (name: string) => `${name} perdeu uma carta.`,
    greed: (name: string) => `${name} passou de 20 moedas e perdeu uma carta.`,
    revived: (name: string) => `${name} reviveu uma carta.`,
    changed: (name: string) => `${name} trocou uma carta.`,
    earned: (name: string, coins: number) => `${name} recebeu ${plural(coins, "moeda", "moedas")}.`,
    took: (name: string, coins: number) => `${name} pegou ${plural(coins, "moeda", "moedas")}.`,
    stole: (name: string, coins: number, target: string) => `${name} roubou ${plural(coins, "moeda", "moedas")} de ${target}.`,
    invested: (name: string) => `${name} investiu.`,
    taxUp: (name: string) => `${name} subiu a taxa.`,
    taxDown: (name: string) => `${name} baixou a taxa.`,
    denied: (name: string, target: string) => `${name} barrou o ganho de ${target}.`,
    suddenDeath: "O baralho acabou. Morte súbita.",
  },

  // fim de jogo
  outOfScene: "Fora da partida",
  outOfSceneRegion: "Você, fora da partida",
  outOfSceneAt: (round: number) => `Suas cartas caíram na rodada ${round}. Você ainda vê a mesa e pode conversar.`,
  deadCardsLabel: (names: string[]) => `Suas duas cartas, mortas: ${names.join(" e ")}`,
  openChat: "Abrir a conversa",
  deckOver: "O baralho acabou.",
  suddenDeath: "Morte súbita",
  suddenDeathLines: ["Morte", "súbita"],
  suddenDeathLower: "morte súbita",
  roundSudden: (turn: number) => `Rodada ${turn},`,
  eachCue: (s: number) => `Cada vez dura ${s} s.`,
  suddenTurnAfter: (s: number) => `. Agora cada vez dura ${s} s.`,
  playResumes: (s: number) => `A partida volta em ${s} s`,
  suddenDeathYouLine: (s: number) => `Cada vez dura ${s} s. Decida rápido.`,
  finale: "Fim da partida",
  finaleCart: "Cartaz do fim da partida",
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
  prosoPlaceholder: "Troque seu personagem",
  prosoCaption: "arte do Proso · 3:4",
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
  login: "Entrar com Google",
  loginModal: {
    kicker: "Bilheteria",
    title: "Entrar",
    text: "Entre com sua conta Google. Sem conta, entre como visitante pelo código de uma sala.",
    loading: "Carregando o Google…",
    back: "☜ Voltar",
    guestOr: "ou entre como visitante",
    guestName: "Seu nome",
    guestEnter: "Entrar como visitante",
    error: (reason: string) => `Não deu para entrar. ${reason}`,
    noToken: "A resposta do servidor veio sem login. Tente de novo.",
    usernameTaken: "Esse nome já está em uso. Escolha outro.",
    offline: "Não deu para falar com o servidor do jogo. Confira se ele está no ar.",
    expired: "Sua sessão expirou. Entre de novo.",
    needName: "Primeira vez aqui. Escolha um nome de usuário.",
    nameLabel: "Nome de usuário",
    nameSubmit: "Criar conta",
  },
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
  guestCanOnlyJoin: "Visitante entra pelo código de uma sala. Para criar a sua, use uma conta Google.",
  serverOffline: "Não deu para falar com o servidor do jogo. Confira se ele está no ar.",

  // relato de bug (plano 06). A pergunta do fim usa os turnos marcados no rascunho.
  bugReportButton: "Algo deu errado",
  bugReportConfirm: "Marcar que algo deu errado agora?",
  bugReportConfirmYes: "Marcar",
  bugReportConfirmNo: "Cancelar",
  bugReportMarked: "Marcado. No fim da partida a gente pergunta o que foi.",
  bugReportQuestion: (turns: number | number[]) => {
    const list = Array.isArray(turns) ? turns : [turns];
    const text =
      list.length <= 1
        ? `${list[0] ?? "?"}`
        : `${list.slice(0, -1).join(", ")} e ${list[list.length - 1]}`;
    return `Você marcou que algo deu errado ${list.length <= 1 ? "no turno" : "nos turnos"} ${text}. O que aconteceu?`;
  },
  bugReportCategories: BUG_CATEGORIES,
  bugReportCategoryLabels: BUG_CATEGORY_LABELS,
  bugReportCategoriesLabel: "O que aconteceu",
  bugReportDescriptionLabel: "Quer contar mais?",
  bugReportDescriptionPlaceholder: "Escreva o que aconteceu",
  bugReportSend: "Enviar",
  bugReportSending: "Enviando…",
  bugReportSkip: "Pular",
  bugReportDevice: "Enviar dados do aparelho",
  bugReportDeviceHint:
    "Inclui a versão do jogo, o tamanho da tela e os últimos erros do navegador. A conversa da partida não vai.",
  bugReportSuccess: "Relato enviado. Obrigado.",
  bugReportError: "Não deu para enviar. Tente de novo.",
  bugReportLimit: "Muitos relatos em pouco tempo. Espere um pouco e tente de novo.",
  bugReportGone: "Esse relato não existe mais",
  bugReportChooseCategory: "Escolha o que aconteceu.",
  bugReportShortDescription: "Escreva pelo menos 10 caracteres.",
  bugReportLobby: "Reportar bug",
  bugReportLobbyTitle: "O que deu errado?",

  // histórico (predicados: o jogador vem antes, no histórico geral)
  log: {
    declared: (role: string) => `se disse ${role}.`,
    noContest: " Ninguém duvidou.",
    contested: (role: string, contester: string) => `se disse ${role}, ${contester} duvidou.`,
    penaltyHad: (contested: string, contester: string) => ` ${contested} mostrou a carta e escolhe uma carta de ${contester}.`,
    penaltyBluff: (contested: string, contester: string) => ` ${contested} perdeu a carta e ${contester} escolhe qual cai.`,
    contestedBy: (target: string) => `duvidou de ${target}.`,
    lostACard: "perdeu uma carta.",
    killedByRole: (target: string, role: string) => `matou uma carta de ${target} como ${role}.`,
    kamikaze: (target: string) => `levou uma carta de ${target} com o Kamikaze.`,
    revealed: (target: string, card?: string) => `revelou uma carta de ${target}${card ? `: ${card}` : ""} como Clarividente.`,
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
    motion: (on: boolean) => `Movimento: ${on ? "ligado" : "desligado"}`,
    nav: [
      ["vez", "Sua vez"],
      ["acoes", "Ações"],
      ["duvidar", "Duvidar"],
      ["taxa", "Taxa"],
      ["elenco", "Personagens"],
      ["casa", "Outras regras"],
    ] as const,
    navLabel: "Seções",
    intro:
      "Cada jogador recebe duas cartas viradas para baixo. Cada carta é um personagem, como o Assassino ou a Guardiã. Na sua vez, você pode dizer que tem qualquer personagem e usar o poder dele, mesmo sem ter a carta. Isso é blefar. Quem desconfiar pode pedir para ver.",
    summaryTitle: "Resumo",
    summary: [
      ["2–10", "jogadores por sala"],
      ["2", "cartas para cada um"],
      ["20", "moedas no máximo"],
      ["±5", "limites da taxa, que muda os preços"],
      ["1", "vencedor: o último com carta"],
    ] as const,
    roundTitle: "Como é uma rodada",
    round: [
      ["Escolha uma ação", "Pegar moeda, matar, trocar cartas ou usar um personagem."],
      ["Os outros respondem", "Se você usou um personagem, os outros podem duvidar, bloquear ou deixar passar."],
      ["O efeito acontece", "Se ninguém impediu, a jogada vale e aparece no histórico da partida."],
      ["A vez passa", "Joga o próximo. Ganha quem ficar por último com alguma carta."],
    ] as const,
    actionsTitle: "Ações do turno",
    actionsScribble: (tax: string) => `As três primeiras você faz sem ter carta nenhuma, e ninguém pode duvidar. Os preços já contam a taxa de agora: ${tax}.`,
    take: "Pegar moeda",
    takeText: "Recebe 1 moeda mais a taxa. Nada impede.",
    kill: "Matar",
    killText: "Paga 6 moedas e elimina uma carta de outro jogador. A Guardiã pode impedir.",
    killTaxLabel: "Como a taxa muda o preço do Matar",
    killTax: ["Taxa positiva: soma só a metade. Taxa +4 → 8.", "Taxa negativa: desconta inteira. Taxa −2 → 4.", "Nunca custa menos de 3."],
    swap: "Trocar cartas",
    swapText: "Paga metade das suas moedas, para cima, e troca uma carta por outra do baralho. Só com 6 moedas ou mais. Trocar as duas custa tudo.",
    use: "Usar um personagem",
    useText: ["Use o poder de qualquer ", "personagem", ", tendo ou não a carta. Aqui os outros podem duvidar."] as const,
    doubtTitle: "Quando alguém duvida",
    doubtLead: "Quando alguém usa um personagem, qualquer jogador que ainda está na partida pode duvidar. A carta é mostrada, e quem errou perde uma carta.",
    doubtBlock: "Bloquear também conta como usar um personagem: quem bloqueia com a Guardiã ou o Egoísta também pode ser desmentido. A Guardiã não protege a carta perdida numa dúvida.",
    bluffStamp: "Blefe?",
    bluffTitle: "Ele não tinha",
    bluffText: "A jogada não acontece. Quem duvidou escolhe qual carta do blefador morre.",
    truthStamp: "Verdade",
    truthTitle: "Ele tinha mesmo",
    truthText: "Ele mostra a carta e escolhe qual carta de quem duvidou morre. Depois troca a carta mostrada por outra do baralho. A jogada vale.",
    taxTitle: "A taxa mexe nos preços",
    taxLead: "A taxa começa em 0 e vai de −5 a +5. Ela entra no preço de quase tudo, mas cada personagem tem um preço mínimo. Conta quebrada é arredondada para baixo.",
    taxNote: "A taxa não muda o preço do Kamikaze, do Político e do Rebelde. O Político sobe 1 ponto, o Rebelde baixa 1.",
    taxTry: "Teste outra taxa",
    taxHint: "Os preços desta página mudam junto.",
    rolesTitle: "Os 12 personagens",
    rolesScribble: (tax: string) => `O baralho tem 2 cartas de cada personagem. Toque num nome para ver o que ele faz. Preços com a taxa de agora: ${tax}.`,
    rolesPick: "Personagens",
    groups: ["Para ganhar moedas", "Para atacar", "Para mexer nas cartas", "Para mudar a taxa", "Fora do seu turno"],
    artSoon: "arte em breve",
    artLabel: (name: string) => `Espaço reservado para a arte animada do ${name}`,
    artFrames: "4:5 · quadro a quadro",
    fields: { quando: "Quando", alvo: "Alvo", faixa: "Preço", barra: "Quem impede", contesta: "Dá para duvidar?" },
    priceAria: (p: string) => `Preço ${p}`,
    houseTitle: "Outras regras",
    house: [
      ["Ninguém fica com mais de ", "20 moedas", ". Se passar disso, o que sobrar se perde e você perde uma carta. Ninguém pode impedir."],
      ["", "Carta morta", " fica virada para cima e não vale mais. Ela vai para a pilha de descarte."],
      ["Cada vez que alguém usa o Político ou o Rebelde, o preço dele ", "dobra para todos", ", até 8. Se passarem 3 turnos sem ninguém usar, cai pela metade, até voltar a 2."],
      ["Quem perde as duas cartas ", "sai da partida", " e só assiste."],
    ] as const,
    endLabel: "Fim da partida",
    endText: "Ganha quem ficar por último com alguma carta.",
    ready: "Pronto para jogar?",
    create: "Criar sala",
  },

  // Camarim (fase 8, plans/03-skins.md)
  camarim: {
    title: "Camarim",
    lead: "Escolha o visual do jogo. Ele vale para todas as telas e fica salvo neste aparelho.",
    preview: "Prévia",
    back: "☜ Voltar ao lobby",
    previewRoom: "Sala nº K7Q2",
    previewCast: "Elenco",
    previewCount: "4 de 10 jogadores",
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

  // Replay (plano 05, T14). Só texto: o arquivo e o estado nunca viram HTML.
  replay: {
    title: "Replay",
    back: "Voltar ao início",
    open: "Abrir .hwr",
    openLabel: "Abrir um arquivo de replay do disco",
    cards: "Cartas",
    cardsTitle: "Todas as cartas",
    official: "replay oficial",
    unofficial: "replay não oficial",
    officialHint: "A assinatura confere com a chave do Hoodwink.",
    unofficialHint: "Este arquivo não tem a assinatura do servidor. Pode ter sido alterado.",
    loading: "Abrindo o replay…",
    emptyTitle: "Nenhum replay aberto",
    emptyText: "Abra um arquivo .hwr do disco ou chegue aqui pelo histórico do seu perfil.",
    emptyFile: "Escolher um arquivo",
    errorTitle: "Não deu para abrir o replay",
    error: "Não foi possível continuar o replay. Abra o arquivo de novo.",
    loadError: "Não foi possível abrir esse replay.",
    retry: "Tentar de novo",
    invalid: "Esse arquivo não é um replay do Hoodwink.",
    tooBig: "Esse arquivo passa de 256 KB. O replay de uma partida cabe em poucos KB.",
    timeout: "O replay demorou demais para abrir. Tente de novo.",
    needLogin: "Entre na sua conta para ver esse replay.",
    notFound: "Esse replay não está disponível.",
    network: "Não foi possível falar com o servidor. Tente de novo.",
    controls: "Controles do replay",
    play: "Tocar",
    pause: "Pausar",
    previous: "Passo anterior",
    next: "Próximo passo",
    speed: "Velocidade",
    speedLabel: (n: number) => `${n} vezes`,
    progress: "Posição do replay",
    match: "Partida",
    localFile: "Arquivo do disco",
    position: (at: number, total: number) => `Passo ${at} de ${total}`,
    finish: "Fim do replay",
    winner: (name: string) => `${name} venceu`,
    noWinner: "A partida terminou.",
    out: "fora do jogo",
    unknownPlayer: "Jogador",
  },

  // Perfil (plano 05, T15). Histórico de partidas e atalhos do replay.
  perfil: {
    title: "Seu perfil",
    back: "Voltar à entrada",
    openLabel: "Abrir perfil",
    history: "Partidas",
    historyHint: "Seus replays ficam guardados por 30 dias.",
    signedOutTitle: "Entre para ver seu perfil",
    signedOutText: "Com sua conta, o Hoodwink guarda os replays das suas partidas.",
    loading: "Carregando suas partidas…",
    loadingMore: "Carregando mais…",
    emptyTitle: "Nenhuma partida ainda",
    emptyText: "Quando uma partida terminar, ela aparece aqui. Os replays ficam guardados por 30 dias.",
    errorTitle: "Não deu para carregar suas partidas",
    retry: "Tentar de novo",
    loadMore: "Carregar mais",
    watch: "Assistir",
    watchLabel: (date: string) => `Assistir à partida de ${date}`,
    download: "Baixar",
    downloadLabel: (date: string) => `Baixar o replay da partida de ${date}`,
    downloading: "Baixando…",
    downloadStarted: "O arquivo está sendo baixado.",
    players: (list: string) => `Jogadores: ${list}`,
    you: "Você",
    player: (id: string) => `Jogador ${id}`,
    youWon: "Você venceu",
    winner: (name: string) => `${name} venceu`,
    noWinner: "Sem vencedor",
    noPlayers: "Jogadores não informados",
    unknownDate: "Data não informada",
    downloadError: "Não foi possível baixar o replay. Tente de novo.",
    loadError: "Não foi possível carregar suas partidas. Tente de novo.",
    needLogin: "Entre na sua conta para ver seus replays.",
    forbidden: "Sua conta não pode ver esses replays.",
    notFound: "Esse replay não está mais disponível.",
    network: "Não foi possível falar com o servidor. Tente de novo.",
    badResponse: "O servidor respondeu de um jeito inesperado. Tente de novo.",
    generic: "Não foi possível concluir o pedido. Tente de novo.",
    replayTitle: "Replay da partida",
    replayChecking: "Procurando o replay…",
    replayPending: "O replay ainda está sendo processado.",
    replayPendingHint: "Isso costuma levar alguns segundos.",
    login: "Entrar",
    logout: "Sair da conta",
  },
};
