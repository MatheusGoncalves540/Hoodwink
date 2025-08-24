package services

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService para geração de tokens
type JWTService struct {
	secret string
}

func NewJWTService() *JWTService {
	secret := os.Getenv("JWT_SECRET")
	return &JWTService{secret: secret}
}

func (j *JWTService) GenerateToken(userId, roomId string) (string, error) {
	expStr := os.Getenv("JWT_EXPIRATION")
	expInt := 2 // default value
	if expStr != "" {
		if val, err := strconv.Atoi(expStr); err == nil {
			expInt = val
		}
	}
	claims := jwt.MapClaims{
		"userId": userId,
		"roomId": roomId,
		"exp":    time.Now().Add(time.Hour * time.Duration(expInt)).Unix(), // expira em expInt horas
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}
