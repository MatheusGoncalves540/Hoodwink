package auth

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = os.Getenv("JWT_SECRET")

// UserClaims representa os dados que estarão no token JWT
func GenerateJWT(data UserClaims) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	var claims jwt.MapClaims
	if err := json.Unmarshal(jsonBytes, &claims); err != nil {
		return "", err
	}

	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ValidateJWT valida o token JWT e retorna o e-mail do usuário
func ValidateJWT(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if Email, ok := claims["email"].(string); ok {
			return Email, nil
		}
		return "", errors.New("claim 'email' ausente ou inválido")
	}

	return "", errors.New("token inválido")
}
