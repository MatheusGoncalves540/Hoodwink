package jwt

// UserClaims representa os dados que estarão no token JWT
type UserClaims struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Level    int    `json:"lvl"`
}
