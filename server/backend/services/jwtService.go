package services

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/structures"
	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret     string
	expiration time.Duration
}

func NewJWTService() *JWTService {
	secret := os.Getenv("JWT_SECRET")
	expStr := os.Getenv("JWT_EXPIRATION")
	expInt, err := strconv.Atoi(expStr)
	if err != nil {
		expInt = 24
	}
	exp := time.Duration(expInt) * time.Hour
	return &JWTService{
		secret:     secret,
		expiration: exp,
	}
}

func (j *JWTService) GenerateJWT(data structures.UserClaims) (string, error) {
	exp := time.Now().UTC().Add(j.expiration).Unix()
	claims := jwt.MapClaims{
		"id":       data.Id,
		"username": data.Username,
		"provider": data.Provider,
		"email":    data.Email,
		"temp":     data.Temp,
		"exp":      exp,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

func (j *JWTService) ValidateJWT(tokenStr string) (structures.UserClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return []byte(j.secret), nil
	})
	if err != nil || !token.Valid {
		return structures.UserClaims{}, errors.New("token inválido")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return structures.UserClaims{}, errors.New("token inválido")
	}
	return structures.UserClaims{
		Id:       claims["id"].(string),
		Username: claims["username"].(string),
		Provider: claims["provider"].(string),
		Email:    claims["email"].(string),
		Temp:     claims["temp"].(bool),
	}, nil
}
