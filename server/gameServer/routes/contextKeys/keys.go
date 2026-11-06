package contextKeys

// ContextKey é uma chave customizada para armazenar dados no contexto
type ContextKey string

const RequestIDKey ContextKey = "requestID"

const PlayerContextKey ContextKey = "player"
