package jwtoken

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = os.Getenv("JWT_SECRET")

// UserClaims representa os dados que estarão no token JWT
func GenerateJWT(data UserClaims) (string, error) {
	claims := jwt.MapClaims{
		"id":       data.Id,
		"username": data.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ValidateJWT valida o token JWT e retorna o e-mail do usuário
func ValidateJWT(tokenStr string) (UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return UserClaims{}, err
	}

	if claims, ok := token.Claims.(*jwt.MapClaims); ok && token.Valid {
		id, _ := (*claims)["id"].(string)
		username, _ := (*claims)["username"].(string)
		return UserClaims{Id: id, Username: username}, nil
	}
	return UserClaims{}, errors.New("token inválido")
}
