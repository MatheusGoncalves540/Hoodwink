package auth

import (
	"errors"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var jwtSecret string

func init() {
	godotenv.Load(".env")
	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		utils.LogDebug("⚠️ JWT_SECRET não definido no ambiente")
	}
}

// ParseToken verifica e retorna claims válidas de um token JWT
func parseToken(tokenStr string) (jwt.MapClaims, error) {
	if tokenStr == "" {
		utils.LogDebug("token vazio")
		return nil, errors.New("token vazio")
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			utils.LogDebug("método de assinatura inválido")
			return nil, errors.New("método de assinatura inválido")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		utils.LogDebug("token inválido")
		return nil, errors.New("token inválido")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.LogDebug("claims inválidas")
		return nil, errors.New("claims inválidas")
	}

	return claims, nil
}

// ParseTokenFromRequest extrai o token JWT de headers ou query params
func ParseTokenFromRequest(r *http.Request) (jwt.MapClaims, error) {
	// via query string: ?token=<token>
	queryToken := r.URL.Query().Get("Ticket")
	if queryToken != "" {
		utils.LogDebug("Usando token da query string")
		return parseToken(queryToken)
	}

	utils.LogDebug("nenhum token encontrado no header ou query")
	return nil, errors.New("nenhum token encontrado no header ou query")
}
