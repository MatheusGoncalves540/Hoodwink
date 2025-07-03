 no seu sistema atual do *Hoodwink*, usando a estrutura já modularizada que você criou.
https://chatgpt.com/c/68651119-a1c4-8000-a05c-c5a76fdc821a
---

## 🧠 Conceito Básico

Adicionar uma carta nova envolve:

1.  Criar a **identidade** da carta (nome, ID, custo, etc.).
    
2.  Criar a **ação/jogada** que essa carta representa.
    
3.  Criar o **tratamento da jogada** no backend:
    
    -   Validação.
        
    -   Agendamento com timer.
        
    -   Contestação (se aplicável).
        
    -   Execução (efeito da carta).
        

---

## ✅ Passo a Passo para Adicionar uma Nova Carta

---

### 1\. 📦 Definir a carta

Edite ou crie o arquivo que centraliza os dados das cartas, como:

**`gameServer/game/room/roomStructs/cards.go`** (exemplo)

```go
package roomStructs

type CardDefinition struct {
	ID           string
	Name         string
	BaseCost     int
	CanBeBlocked bool
	Contestable  bool
	TimedAction  bool
}

var CardDefinitions = map[string]CardDefinition{
	"assassin": {
		ID:           "assassin",
		Name:         "Assassino",
		BaseCost:     3,
		CanBeBlocked: true,
		Contestable:  true,
		TimedAction:  true,
	},
	"investidor": {
		ID:           "investidor",
		Name:         "Investidor",
		BaseCost:     0,
		CanBeBlocked: false,
		Contestable:  false,
		TimedAction:  false,
	},
	// Adicione sua nova carta aqui:
	"minhoca": {
		ID:           "minhoca",
		Name:         "Minhoca",
		BaseCost:     5,
		CanBeBlocked: false,
		Contestable:  true,
		TimedAction:  true,
	},
}
```

---

### 2\. 🧠 Criar o `Event.Type` correspondente à jogada

Abra o arquivo:

**`gameServer/game/room/handlers/actions.go`**

E crie uma nova função para processar o evento da nova carta:

```go
package handlers

import (
	"context"

	"gameServer/game/room/roomStructs"
)

func HandleUseMinhoca(ctx context.Context, room *roomStructs.Room, evt *roomStructs.Event) error {
	payload := evt.Payload.(map[string]interface{})
	targetUUID := payload["target"].(string)

	room.PendingEffects = append(room.PendingEffects, roomStructs.Effect{
		Type:       "minhoca_effect",
		From:       evt.PlayerId,
		To:         targetUUID,
		CardIndex:  -1,
		Executable: false,
		Reason:     "usou minhoca",
	})

	room.CurrentMove = &roomStructs.Move{
		PlayerId: evt.PlayerId,
		Action:     "use_minhoca",
		TargetUUID: targetUUID,
	}

	room.State = roomStructs.WaitingContest
	return nil
}
```

---

### 3\. 🧩 Ligar no `ProcessEvent`

Abra:

**`gameServer/game/room/eventQueue/eventQueue.go`**

E modifique o `ProcessEvent` para adicionar a nova jogada:

```go
switch evt.Type {
case "use_minhoca":
	err := handlers.HandleUseMinhoca(ctx, room, evt)
	if err != nil {
		return err
	}
	redisHandlers.ScheduleNextStep(ctx, rdb, room.ID, roomStructs.Event{
		Type:          "no_contest",
		PlayerId:    "system",
		TimeoutMs: 8000,
	})
```

---

### 4\. 🛠️ Definir o efeito da carta

No `ApplyPendingEffects`, defina o que o efeito da carta faz. Exemplo:

```go
func ApplyPendingEffects(room *roomStructs.Room) {
	for _, eff := range room.PendingEffects {
		if !eff.Executable {
			continue
		}
		switch eff.Type {
		case "minhoca_effect":
			// Lógica real do efeito minhoca
			_ = roomStructs.KillCard(room, eff.To, 1)
		case "kill":
			_ = roomStructs.KillCard(room, eff.To, eff.CardIndex)
		}
	}
	room.PendingEffects = nil
}
```

---

### 5\. 🧪 Criar o JSON da jogada no cliente (exemplo)

```json
{
  "type": "use_minhoca",
  "playerId": "1234-uuid",
  "payload": {
    "target": "uuid-do-inimigo"
  },
  "timeout_millis": 8000
}
```

---

## 💡 Dicas Extras

-   **Contestação** já está sendo tratada automaticamente por `WaitingContest`, `resolve_contest`, etc.
    
-   **Timers**: você só precisa definir `TimeoutMs` e agendar o `ScheduleNextStep` como já está fazendo.
    

---

## ✅ Resumo

Para adicionar uma nova carta e jogada:

| Etapa | Local |
| --- | --- |
| Definir a carta | `roomStructs/cards.go` |
| Criar handler da jogada | `handlers/actions.go` |
| Ligar no ProcessEvent | `eventQueue/eventQueue.go` |
| Definir efeito da carta | `ApplyPendingEffects` em `eventQueue.go` |
| Criar payload de teste | Cliente/WebSocket |
