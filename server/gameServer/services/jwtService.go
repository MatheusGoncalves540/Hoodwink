package services

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/golang-jwt/jwt/v5"
)

// JWTService para geração de tokens
type JWTService struct {
	jwtSecret        string
	backendJWTSecret string
}

func NewJWTService() *JWTService {
	jwtSecret := os.Getenv("JWT_SECRET")
	backendJWTSecret := os.Getenv("BACKEND_JWT_SECRET")
	if jwtSecret == "" || backendJWTSecret == "" {
		utils.LogError("⚠️ JWT secrets não definidos no ambiente")
		return nil
	}

	return &JWTService{jwtSecret, backendJWTSecret}
}

func (j *JWTService) GenerateToken(player *endpointStructures.ClaimsBackend, roomId string) (string, error) {
	expStr := os.Getenv("JWT_EXPIRATION")
	expInt := 2 // valor padrão
	if expStr != "" {
		if val, err := strconv.Atoi(expStr); err == nil {
			expInt = val
		}
	}
	claims := endpointStructures.ClaimsHoodwink{
		PlayerID: player.Id,
		Username: player.Username,
		RoomId:   roomId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expInt))),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.jwtSecret))
}

// ParseToken verifica e retorna claims válidas de um token JWT
// useBackendSecret: true → usa BACKEND_JWT_SECRET, false → usa JWT_SECRET
func (j *JWTService) ParseToken(tokenStr string, useBackendSecret bool) (jwt.MapClaims, error) {
	if tokenStr == "" {
		utils.LogError("token vazio")
		return nil, errors.New("token vazio")
	}

	// Decide qual secret usar
	secret := j.jwtSecret
	if useBackendSecret {
		secret = j.backendJWTSecret
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			utils.LogError("método de assinatura inválido")
			return nil, errors.New("método de assinatura inválido")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		utils.LogError("token inválido")
		return nil, errors.New("token inválido")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.LogError("claims inválidas")
		return nil, errors.New("claims inválidas")
	}

	return claims, nil
}

// ParseTokenFromRequest extrai o token JWT de headers ou query params
// (sempre usa JWT_SECRET, nunca o BACKEND_JWT_SECRET)
func (j *JWTService) ParseTokenFromRequest(r *http.Request) (jwt.MapClaims, error) {
	// via query string: ?token=<token>
	queryToken := r.URL.Query().Get("Ticket")
	if queryToken != "" {
		utils.LogDebug("Usando token da query string")
		return j.ParseToken(queryToken, false) // sempre JWT_SECRET
	}

	utils.LogError("nenhum token encontrado no header ou query")
	return nil, errors.New("nenhum token encontrado no header ou query")
}
