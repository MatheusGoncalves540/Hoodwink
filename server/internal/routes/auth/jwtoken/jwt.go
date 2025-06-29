package jwtoken

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = os.Getenv("JWT_SECRET")
var jwtExpiration = 24 * time.Hour // pode ser parametrizado via env

// UserClaims representa os dados que estarão no token JWT
func GenerateJWT(data UserClaims) (string, error) {
	exp := time.Now().Add(jwtExpiration).Unix()
	claims := jwt.MapClaims{
		"id":       data.Id,
		"username": data.Username,
		"provider": data.Provider,
		"email":    data.Email,
		"temp":     data.Temp,
		"exp":      exp,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ValidateJWT valida o token JWT e retorna o UserClaims
func ValidateJWT(tokenStr string) (UserClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return UserClaims{}, errors.New("token inválido")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return UserClaims{}, errors.New("token inválido")
	}
	return UserClaims{
		Id:       claims["id"].(string),
		Username: claims["username"].(string),
		Provider: claims["provider"].(string),
		Email:    claims["email"].(string),
		Temp:     claims["temp"].(bool),
	}, nil
}
