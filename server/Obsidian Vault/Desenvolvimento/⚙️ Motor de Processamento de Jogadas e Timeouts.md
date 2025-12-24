Este documento descreve o funcionamento do **motor de jogadas** do servidor do jogo, responsável por avançar o estado das salas, aplicar timeouts e garantir consistência em um ambiente **distribuído e escalável**.

---

## 🎯 Objetivo do motor

O motor de jogadas existe para:

-   Avançar o jogo automaticamente quando um jogador não age
    
-   Processar eventos de forma segura em múltiplos pods
    
-   Evitar timers em memória ou goroutines frágeis
    
-   Garantir que o jogo continue mesmo se um pod cair
    

Ele **não depende de WebSocket**, **não depende de conexão ativa** e **não depende de estado local**.

---

## 🧠 Princípios fundamentais

### 1\. Redis é a fonte de verdade

Todo o estado relevante do jogo é persistido no Redis.

-   Estado da sala
    
-   Evento pendente
    
-   Timeout atual
    
-   Locks distribuídos
    

Arquivos relacionados:

-   `loadRoom.go`
    
-   `saveRoom.go`
    
-   `roomTTL.go`
    

---

### 2\. Timeout não é timer

O servidor **não espera** eventos com `time.Sleep`, `time.After` ou goroutines dedicadas.

Um timeout é apenas:

> “um instante no tempo em que algo deve acontecer”

Esse instante é salvo no Redis como um **timestamp absoluto**.

Arquivo relacionado:

-   `ScheduleNextStep.go`
    

---

### 3\. Todos os pods processam, apenas um executa

Cada instância do servidor roda o motor de jogadas em paralelo.

Antes de processar uma sala, o pod tenta adquirir um **lock distribuído**.

Se falhar, outro pod está processando.

Arquivo relacionado:

-   `tryLockRoom.go`
    
-   `lock.go`
    

---

## ⏱️ Modelo de timeout distribuído

Os timeouts das salas são armazenados em um **Redis ZSET**.

-   Cada sala aparece uma única vez
    
-   A score é o timestamp de expiração (alta precisão)
    
-   O member é o ID da sala
    

O motor apenas consulta:

> “Quais salas já passaram do horário?”

Arquivo central:

-   `gameProcessorEngine.go`
    

---

## 🔄 Ciclo de vida de uma jogada (exemplo simples)

### Exemplo: turno automático (sem ação do jogador)

1.  A sala entra em um estado que exige espera  
    Exemplo: “aguardando fim do turno”
    
2.  Um timeout é agendado  
    Exemplo: “daqui a 5 segundos”
    
3.  O motor detecta que o tempo venceu
    
4.  O estado da sala avança  
    Exemplo: incrementa o turno
    
5.  Um novo timeout é agendado  
    Exemplo: próximo turno
    
6.  O ciclo se repete indefinidamente
    

Arquivos envolvidos:

-   `gameProcessorEngine.go`
    
-   `ScheduleNextStep.go`
    
-   `state.go`
    
-   `gameEvents.go`
    

---

## ▶️ Inicialização do jogo (ponto crítico)

O motor **não inicia o jogo sozinho**.

É obrigatório que, em algum ponto do fluxo (ex: criação da sala ou primeiro jogador), o jogo seja explicitamente iniciado.

Isso significa:

-   Definir o estado inicial
    
-   Definir o primeiro evento pendente
    
-   Agendar o primeiro timeout
    

Se isso não acontecer, o motor roda corretamente, mas **não há nada para processar**.

Arquivos relacionados:

-   `createRoomHandler.go`
    
-   `roomService.go`
    

---

## 🎮 Interação com WebSocket

O WebSocket **não avança o jogo** diretamente.

Ele apenas:

-   Recebe ações dos jogadores
    
-   Valida essas ações
    
-   Atualiza o estado da sala
    
-   Cancela ou substitui timeouts existentes
    

O motor continua responsável por:

-   Avançar o estado
    
-   Resolver timeouts
    
-   Garantir progressão
    

Arquivos relacionados:

-   `connManager.go`
    
-   `pubsub.go`
    
-   `processWebsocketMessage.go`
    
-   `onMessage.go`
    

---

## 🧩 Estados e eventos (conceito)

O jogo funciona como uma **máquina de estados**.

Cada estado:

-   Representa uma fase do jogo
    
-   Possui exatamente um evento pendente
    
-   Possui exatamente um timeout ativo
    

Exemplos conceituais:

-   Aguardando ação do jogador
    
-   Aguardando contestação
    
-   Aguardando escolha de carta
    
-   Aguardando decisão final
    

Arquivo de referência:

-   `state.go`
    
-   `gameEvents.go`
    

---

## ❗ Regras de ouro do motor

-   Toda sala **sempre** deve ter um estado válido
    
-   Todo estado **sempre** deve ter um timeout
    
-   Todo timeout vencido **deve** gerar o próximo passo
    
-   Timeout antigo é removido **antes** de criar o próximo
    
-   Nunca usar timers em memória
    
-   Nunca confiar em apenas um pod
    

Se qualquer uma dessas regras for quebrada, o jogo pode “travar silenciosamente”.

---

## 🧪 Debug e observabilidade

Quando algo não funciona, verificar nesta ordem:

1.  O jogo foi iniciado?
    
2.  A sala aparece no ZSET de timeouts?
    
3.  O timeout está no futuro?
    
4.  O timeout antigo é removido?
    
5.  Um novo timeout está sendo criado?
    
6.  O estado da sala muda?
    

Arquivos úteis para debug:

-   `processDebugCommand.go`
    
-   `instanceStatusHandler.go`
    
-   `instanceUtils.go`
    

---

## 🟢 Benefícios da arquitetura

-   Escala horizontal sem coordenação
    
-   Tolerante a falhas de instância
    
-   Fácil de estender com novas cartas e regras
    
-   Funciona bem com Redis Pub/Sub
    
-   Base sólida para jogo competitivo multiplayer
    

---

## 🏁 Conclusão

O motor de jogadas do projeto é baseado em:

-   Máquina de estados explícita
    
-   Timeouts absolutos
    
-   Processamento distribuído
    
-   Redis como coordenador central
    

Essa base permite implementar:

-   Turnos
    
-   Cartas com efeitos completos
    
-   Contestação
    
-   Efeitos encadeados
    
-   Expansões futuras
    

Sem alterar o motor central, apenas adicionando novos estados e regras.