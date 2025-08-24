## 🧩 Adicionando Novos Services e Handlers

Este projeto segue uma arquitetura modular baseada na separação de responsabilidades. Toda nova funcionalidade deve ser dividida entre:

-   **Service**: lógica de negócio (em `server/services`)
    
-   **Handler**: interface HTTP (em `server/routes/handlers`)
    
-   **Roteamento**: endpoint e rota (em `server/routes/routes.go`)
    

---

### ✅ 1. Criar um novo Service

1.  Crie o arquivo do novo service em:  
    `server/services/roomService.go`
    
2.  Defina a struct do service, com seu `New` e métodos:
    

```go
package services

import (
	"errors"
	"gorm.io/gorm"
)

type RoomService struct {
	db *gorm.DB
}

func NewRoomService(db *gorm.DB) *RoomService {
	return &RoomService{db}
}

func (s *RoomService) CriarSala(nome string) error {
	if nome == "" {
		return errors.New("nome obrigatório")
	}
	// lógica para criar sala
	return nil
}
```

3.  Registre no `SetupServices`:
    

```go
func SetupServices(db *gorm.DB) *Services {
	userService := NewUserService(db)
	roomService := NewRoomService(db) // Adicionado

	return &Services{
		UserService: userService,
		RoomService: roomService, // Adicionado
	}
}
```

4.  Adicione o campo no tipo `Services`:
    

```go
type Services struct {
	UserService *UserService
	RoomService *RoomService // Novo campo
}
```

---

### ✅ 2. Criar um novo Handler

1.  Crie o arquivo em:  
    `server/routes/handlers/roomHandler.go`
    
2.  Implemente o handler recebendo o `*RoomService`:
    

```go
package handlers

import (
	"encoding/json"
	"net/http"
)

type CriarSalaPayload struct {
	Nome string `json:"nome"`
}

func (h *Handler) CriarSala(w http.ResponseWriter, r *http.Request) {
	var payload CriarSalaPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Nome == "" {
		utils.SendError(w, "Nome inválido", http.StatusBadRequest)
		return
	}

	if err := h.RoomService.CriarSala(payload.Nome); err != nil {
		utils.SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
```

---

### ✅ 3. Atualizar o handler principal (`handler.go`)

1.  Adicione o campo na struct:
    

```go
type Handler struct {
	UserService *services.UserService
	RoomService *services.RoomService // Novo campo
}
```

2.  Atualize a função `NewHandler`:
    

```go
func NewHandler(s *services.Services) *Handler {
	return &Handler{
		UserService: s.UserService,
		RoomService: s.RoomService, // Adicionado
	}
}
```

---

### ✅ 4. Registrar a nova rota

1.  No arquivo `server/routes/routes.go`, registre a nova rota:
    

```go
r.Post("/salas", h.CriarSala)
```

---

### ✅ Exemplo de requisição

```http
POST /salas
Content-Type: application/json

{
  "nome": "Sala Principal"
}
```

---

### 📌 Observações

-   Sempre que criar um novo service, **adicione ele no `SetupServices()`**.
    
-   Sempre que criar um novo handler, **adicione a chamada no `routes.go`**.
    
-   A struct `Handler` centraliza o acesso aos services para uso nas rotas.
    

---