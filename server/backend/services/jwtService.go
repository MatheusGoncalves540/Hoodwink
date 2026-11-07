package services

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/structures"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// LoginAudience identifica o token de login (HS256).
	LoginAudience = "hoodwink-login"
	// TicketAudience identifica o ticket de sala (EdDSA).
	TicketAudience = "hoodwink-ticket"
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
	now := time.Now().UTC()
	exp := now.Add(j.expiration).Unix()
	claims := jwt.MapClaims{
		"id":       data.Id,
		"username": data.Username,
		"provider": data.Provider,
		"email":    data.Email,
		"temp":     data.Temp,
		"ip":       data.IP,
		"ua":       data.UA,
		"iat":      now.Unix(),
		"aud":      LoginAudience,
		"exp":      exp,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

func (j *JWTService) ValidateJWT(tokenStr string) (structures.UserClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return []byte(j.secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return structures.UserClaims{}, errors.New("token inválido")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return structures.UserClaims{}, errors.New("token inválido")
	}
	if !audValidForLogin(claims) {
		return structures.UserClaims{}, errors.New("token inválido")
	}

	id, _ := claims["id"].(string)
	username, _ := claims["username"].(string)
	provider, _ := claims["provider"].(string)
	email, _ := claims["email"].(string)
	temp, _ := claims["temp"].(bool)

	// Claims novas (ip/ua/iat) são sempre lidas com comma-ok: um token antigo
	// sem elas continua válido.
	ip, _ := claims["ip"].(string)
	ua, _ := claims["ua"].(string)
	var iat int64
	if v, ok := claims["iat"].(float64); ok {
		iat = int64(v)
	}

	return structures.UserClaims{
		Id:       id,
		Username: username,
		Provider: provider,
		Email:    email,
		Temp:     temp,
		IP:       ip,
		UA:       ua,
		Iat:      iat,
	}, nil
}

// audValidForLogin aceita o token de login atual (aud: hoodwink-login) e o
// token antigo sem aud até vencer. Recusa qualquer outra audiência (ex.: um
// ticket com aud hoodwink-ticket nunca vale como login).
func audValidForLogin(claims jwt.MapClaims) bool {
	aud, present := claims["aud"]
	if !present {
		return true
	}
	if s, ok := aud.(string); ok {
		return s == LoginAudience
	}
	if list, ok := aud.([]interface{}); ok {
		for _, item := range list {
			if s, ok := item.(string); ok && s == LoginAudience {
				return true
			}
		}
		return false
	}
	return false
}
