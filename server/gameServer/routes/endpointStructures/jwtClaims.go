package endpointStructures

import "github.com/golang-jwt/jwt/v5"

type ClaimsHoodwink struct {
	PlayerID string `json:"playerId"`
	Username string `json:"username"`
	RoomId   string `json:"roomId"`
	jwt.RegisteredClaims
}

type ClaimsBackend struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Provider string `json:"provider"`
	Email    string `json:"email"`
	Temp     bool   `json:"temp,omitempty"`
	Exp      int64  `json:"exp"`
}
