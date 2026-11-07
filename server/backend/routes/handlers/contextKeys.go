package handlers

// ContextKey é uma chave customizada para armazenar dados no contexto.
type ContextKey string

// UserContextKey guarda o structures.UserClaims validado pelo JWTMiddleware.
// Fica aqui (e não no pacote de middlewares) para o handler poder ler sem criar
// ciclo de import: middlewares → handlers.
const UserContextKey ContextKey = "user"
