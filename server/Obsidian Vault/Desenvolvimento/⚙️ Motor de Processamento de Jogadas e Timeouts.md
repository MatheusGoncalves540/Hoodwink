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

-   `gameProcessor.go`

---

## 📝 Definições necessárias para criar uma jogada

1.  timeout - é o registro no redis com um timestamp (UTC) de quando aquela ação "vence".
    -   ou seja, é processada automaticamente pelo motor, sem ação de playe
    -   exemplo:
        ```go
        expiresAt := time.Now().Add(7 * time.Second).UTC()
        rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
            Score:  float64(expiresAt.UnixMilli()),
            Member: roomId,
        })
        ```

2.  gameEvent - é o que está sendo mostrado na tela agora para os players.
    -   Não é responsavel por executar nada
    -   Não tem uma lista de gameEvent, é unico
    -   O gameEvent é removido assim que ele é processado
    -   exemplo:
        ```go
        expiresAt := time.Now().Add(7 * time.Second).UTC()
        roomData.GameEvent = &roomStructs.GameEvent{
            PlayerID:  "ID DO PLAYER QUE FEZ AÇÃO",
            Type:      roomStructs."QUAL AÇÃO FEITA",
            ExpiresAt: expiresAt, // -> Tempo de Timeout
            Payload: map[string]interface{}{
                "TargetPlayer": payload.TargetPlayer,
                "TargetCard":   payload.TargetCard,
                "Cause":        effect.Cause,
            },
        }
        ```

3.  pendingEffect - é uma fila de efeitos que realmente vão ser executados na sala.
    -   Segundo o motor, só é executado quando não tem um gameEvent pois significa que chegou a hora de executar o efeito (ou lista de efeitos), que estava na tela.
        -   A mesma coisa vale para quando ocorre ação de player (gameEvent é marcado como nil)
    -   Exemplo:
    ```go
    roomData.PendingEffects = append(roomData.PendingEffects,
        roomStructs.Effect{
            Cause:        roomStructs.EffectAssassin,
            SourcePlayer: playerPlay.PlayerId,
            Payload: roomStructs.AssassinPayload{
                TargetPlayer: assassinPayload.TargetPlayer,
                TargetCard:   assassinPayload.TargetCard,
            },
        },
    )
    ```
    -   Na maioria das vezes, os effects adicionam GameEvents e outros Effects quando são executados
        -   Exemplo: Assassino quando mata alguem -> GameEvent mostrando para todos que ele matou a carta, e se possivel, abre janela para o atacado usar Kamikaze.

---

## 🔄 Ciclo de vida de uma jogada (exemplo simples)

### Exemplo: turno automático (sem ação do jogador)

1.  A sala entra em um estado que exige espera  
    Exemplo: “aguardando fim do turno”
    
2.  Um timeout, um gameEvent e um gameEffect é agendado  
    Exemplo: “daqui a 5 segundos”
    
3.  O motor detecta que o tempo venceu
    
4.  O estado da sala avança  
    Exemplo: incrementa o turno

5.  O estado da sala é salvo e enviado via broadcast  
    
6.  Um novo timeout é agendado  
    Exemplo: próximo turno
    
7.  O ciclo se repete indefinidamente
    

Arquivos envolvidos:

-   `gameProcessor.go`
    
-   `resolveEffects.go`

---

## 🎮 Interação com WebSocket

O WebSocket **pode avançar o jogo** diretamente.

Ele apenas:

-   Recebe ações dos jogadores
    
-   Valida essas ações
    
-   Atualiza o estado da sala (adicionando GameEvents ou effects)

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
    
-   Possui exatamente um evento e efeitos pendentes
    
-   Possui exatamente um timeout ativo
    

Exemplos conceituais:

-   Aguardando ação do jogador
    
-   Aguardando contestação
    
-   Aguardando escolha de carta
    
-   Aguardando decisão final
    

Arquivo de referência:
    
-   `gameProcessor.go`

-   `resolveEvents.go`
    

---

## ❗ Regras de ouro do motor

-   Toda sala **sempre** deve ter um estado válido
    
-   Todo estado **sempre** deve ter um timeout
    
-   Todo timeout vencido **deve** gerar o próximo passo
    
-   Timeout antigo é removido **antes** de criar o próximo
    
-   Nunca usar timers em memória
    
-   Nunca confiar em apenas um pod
    

Se qualquer uma dessas regras for quebrada, o jogo pode “travar silenciosamente”.
