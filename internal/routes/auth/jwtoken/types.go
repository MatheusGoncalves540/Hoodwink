package jwtoken

// UserClaims representa os dados que estarão no token JWT
type UserClaims struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Provider string `json:"provider"`
	Email    string `json:"email"`
	Temp     bool   `json:"temp,omitempty"`
}
