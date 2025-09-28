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

// JWTService para geração de tokens
type JWTService struct {
	secret string
}

func NewJWTService() *JWTService {
	secret := os.Getenv("JWT_SECRET")
	return &JWTService{secret: secret}
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
	return token.SignedString([]byte(j.secret))
}

// ParseToken verifica e retorna claims válidas de um token JWT
// useBackendSecret: true → usa BACKEND_JWT_SECRET, false → usa JWT_SECRET
func (j *JWTService) ParseToken(tokenStr string, useBackendSecret bool) (jwt.MapClaims, error) {
	if tokenStr == "" {
		utils.LogDebug("token vazio")
		return nil, errors.New("token vazio")
	}

	// Decide qual secret usar
	secret := jwtSecret
	if useBackendSecret {
		secret = os.Getenv("BACKEND_JWT_SECRET")
		if secret == "" {
			utils.LogDebug("⚠️ BACKEND_JWT_SECRET não definido no ambiente")
			return nil, errors.New("backend secret não definido")
		}
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			utils.LogDebug("método de assinatura inválido")
			return nil, errors.New("método de assinatura inválido")
		}
		return []byte(secret), nil
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
// (sempre usa JWT_SECRET, nunca o BACKEND_JWT_SECRET)
func (j *JWTService) ParseTokenFromRequest(r *http.Request) (jwt.MapClaims, error) {
	// via query string: ?token=<token>
	queryToken := r.URL.Query().Get("Ticket")
	if queryToken != "" {
		utils.LogDebug("Usando token da query string")
		return j.ParseToken(queryToken, false) // sempre JWT_SECRET
	}

	utils.LogDebug("nenhum token encontrado no header ou query")
	return nil, errors.New("nenhum token encontrado no header ou query")
}
