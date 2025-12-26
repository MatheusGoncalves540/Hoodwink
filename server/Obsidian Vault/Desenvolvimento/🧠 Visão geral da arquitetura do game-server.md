-   Como será a comunicação entre cliente, backend e game-server
    
-   Como o JWT do backend é usado para verificar autenticação
    
-   Como o game-server gera e valida o ticket exclusivo da sala

---

| Componente                  | Responsabilidade                                                                |
| --------------------------- | ------------------------------------------------------------------------------- |
| **Backend** (HTTP API)      | Autenticação de usuários, gerenciamento de contas, geração de JWT de identidade |
| **Game-server** (HTTP + WS) | Gerencia salas, controla presença, gera tickets JWT por sala, aceita WS         |

---

## 🔁 Fluxo completo: Autenticado e Guest

### 🧱 Pré-requisitos

-   O **backend** gera um JWT de identidade (chave privada do backend)
    
-   O **game-server** conhece a `jwtPublicKey` do backend (para verificar os tokens de identidade)
    
-   O **game-server** gera sua própria `jwtKey` por sala

---

### 🧑‍💼 Fluxo do **jogador autenticado**

1.  **Login no backend**  
    → Recebe um JWT (`jwtUsuario`) com payload:
    
    ```json
    {
      "uuid": "user-abc-123",
      "nick": "matheus",
      "exp": "..."
    }
    ```
    
2.  **Cliente envia requisição para o game-server:**
    
    ```php-template
    POST /getTicket/<RoomId>
    Header: tokenJwt=<jwtUsuario>
    ```
    
3.  **Game-server valida:**
    
    -   Verifica assinatura do `tokenJwt` com a `jwtPublicKey` do backend
        
    -   Extrai `uuid` e `nick` do payload
        
    -   Verifica se a sala `<RoomId>` existe no Redis
	    
    -   Verifica no redis se: presence:`<PlayerId>` existe.
	    - Se existir: Rejeita conexão (player já está na sala).
	    - `ideia futura`: Verifica se é igual `<RoomId>`:
		    - Se for igual: 
			- Se for diferente: 
		- Se não existir: Procede (o Player não está conectado em nenhuma sala)
        
    -   Gera um **JWT exclusivo daquela sala**:
        
        ```json
        {
          "uuid": "user-abc-123",
          "nick": "matheus",
          "RoomId": "abc123",
          "exp": "30min"
        }
        ```
        
    -   Assina com a `jwtKey` da sala
        
4.  **Game-server responde com:**
    
    ```json
    {
      "ticket": "<jwtDaSala>"
    }
    ```
    
5.  **Cliente conecta no WebSocket:**
    
    ```php-template
    /enterRoom/<RoomId>
    Header: Authorization: Bearer <jwtDaSala>
    ```
    

---

### 👤 Fluxo do **guest**

1.  **Cliente envia diretamente ao game-server:**
    
    ```css
    POST /getTicket/<RoomId>
    Body: { "username": "player123" }
    ```
    
2.  **Game-server valida:**
    
    -   Verifica se a sala permite `guest` (ex: não ranqueada)
        
    -   Gera UUID temporária ou `guest:player123#random`
        
    -   Gera o **JWT da sala**, igual ao fluxo autenticado
        
3.  **Resposta e conexão WS** idênticas:
    
    ```json
    {
      "ticket": "<jwtDaSala>"
    }
    ```
