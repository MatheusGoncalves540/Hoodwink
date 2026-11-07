package structures

// UserClaims representa os dados que estarão no token JWT
type UserClaims struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Provider string `json:"provider"`
	Email    string `json:"email"`
	Temp     bool   `json:"temp,omitempty"`
	// IP e UA entram no token só como registro (rastreio de trapaça); nunca são
	// usados para recusar. Iat é o instante de emissão (Unix) para o refresh.
	IP  string `json:"ip,omitempty"`
	UA  string `json:"ua,omitempty"`
	Iat int64  `json:"iat,omitempty"`
}
