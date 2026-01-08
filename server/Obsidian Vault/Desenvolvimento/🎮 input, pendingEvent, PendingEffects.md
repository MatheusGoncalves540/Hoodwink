Este diagrama mostra **uma jogada real** com **Assassino**, **Contestação** e **Kamikaze**.

---

## 🟢 1️⃣ Input do jogador (WebSocket)

```css
[ CLIENTE ]
Player 1 clica:
"Usar carta ASSASSINO no Player 2"
        |
        v
[ WebSocket Handler ]
```

📌 Aqui **não executa nada do jogo ainda**.

---

## 🟡 2️⃣ Servidor recebe o input

```less
[ processWebsocketMessage ]
        |
        v
[ validateAction ]
        |
        v
[ OnPlayerEvent ]
```

Ações feitas pelo servidor:

```diff
+ Criar PendingEffect: ASSASSINO
+ Criar GameEvent: WAIT_CONTEST
+ Salvar sala no Redis
+ Broadcast: "Player 1 usou Assassino"
```

---

## 📦 Estado da sala agora

```makefile
GameEvent:
  Type = WAIT_CONTEST
  ExpiresAt = +5s

PendingEffects:
  [ ASSASSINO ]
```

📺 Frontend mostra:

> “Player 1 usou Assassino. Alguém quer contestar?”

---

## 🔴 3️⃣ Janela de reação (tempo passa)

### Caso A: Player 2 CONTESTA

```css
[ CLIENTE ]
Player 2 clica: "Contestar"
        |
        v
[ WebSocket Handler ]
```

Servidor:

```diff
+ Resolve contestação
+ Cria efeito PenalidadeContestação (se aplicável)
+ Remove GameEvent
+ Salva sala
```

Estado agora:

```makefile
GameEvent:
  nil

PendingEffects:
  [ ASSASSINO, PENALIDADE_CONTESTAÇÃO ]
```

---

### Caso B: Timeout (ninguém contesta)

```less
[ GameProcessor Tick ]
        |
        v
[ AdvanceByTimeout ]
```

Servidor:

```diff
+ Remove GameEvent
```

Estado agora:

```makefile
GameEvent:
  nil

PendingEffects:
  [ ASSASSINO ]
```

---

## ⚙️ 4️⃣ Motor começa a executar efeitos

```yaml
[ GameProcessor Tick ]
        |
        v
Se GameEvent == nil
E PendingEffects != vazio
        |
        v
Executar próximo efeito
```

---

## 🗡️ 5️⃣ Execução do efeito ASSASSINO

```arduino
[ ExecuteEffect ASSASSINO ]
        |
        v
- Mata uma carta do Player 2
- Broadcast: "Carta do Player 2 morreu"
```

Efeito colateral:

```diff
+ Criar PendingEffect: CHECK_KAMIKAZE (Player 2)
```

Estado agora:

```makefile
PendingEffects:
  [ PENALIDADE_CONTESTAÇÃO, CHECK_KAMIKAZE ]
```

---

## 🟠 6️⃣ Kamikaze precisa de decisão

```yaml
[ ExecuteEffect CHECK_KAMIKAZE ]
        |
        v
Criar GameEvent: WAIT_KAMIKAZE_DECISION
Broadcast: "Player 2 quer usar Kamikaze?"
```

Estado:

```makefile
GameEvent:
  WAIT_KAMIKAZE_DECISION

PendingEffects:
  [ PENALIDADE_CONTESTAÇÃO ]
```

📺 Frontend mostra botão:

> “Usar Kamikaze?”

---

## 🟣 7️⃣ Jogador decide (ou timeout)

### Player aceita

```css
[ CLIENTE ]
Clica: "Usar Kamikaze"
        |
        v
[ OnPlayerEvent ]
```

Servidor:

```diff
+ Criar PendingEffect: EXECUTE_KAMIKAZE
+ Remove GameEvent
```

---

### Player recusa ou timeout

```diff
+ Remove GameEvent
```

---

## ⚙️ 8️⃣ Motor continua sozinho

```diff
[ GameProcessor Tick ]
        |
        v
Executa:
- EXECUTE_KAMIKAZE (se existir)
- PENALIDADE_CONTESTAÇÃO
```

Cada morte pode gerar novo `CHECK_KAMIKAZE`.

---

## 🔁 9️⃣ Loop infinito controlado

```yaml
Enquanto:
  GameEvent == nil
  AND
  PendingEffects != vazio
```

O motor continua.

Quando acabar:

```bash
+ Avançar turno
+ Criar GameEvent do próximo turno
```

---

# 🧠 Resumo visual rápido

```css
INPUT (WS)
   ↓
Criar PendingEffect
Criar GameEvent
   ↓
Esperar input ou timeout
   ↓
Remover GameEvent
   ↓
Executar PendingEffects
   ↓
Gerou nova decisão?
   ↓
Criar novo GameEvent
```

---

# 🏁 Frase final (para fixar)

> **GameEvent é a pergunta.  
> PendingEffects é a resposta.  
> O motor só responde quando ninguém precisa falar.**