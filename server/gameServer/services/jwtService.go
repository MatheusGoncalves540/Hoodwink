package services

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/golang-jwt/jwt/v5"
)

// TicketAudience identifica o ticket de sala assinado pelo backend (EdDSA).
const TicketAudience = "hoodwink-ticket"

// JWTService só verifica tickets EdDSA assinados pelo backend com a chave
// pública TICKET_PUBLIC_KEY. Não guarda mais segredo algum.
type JWTService struct {
	publicKey ed25519.PublicKey
}

func NewJWTService() *JWTService {
	publicKey, err := parseTicketPublicKey(os.Getenv("TICKET_PUBLIC_KEY"))
	if err != nil {
		log.Fatalf("Erro ao carregar TICKET_PUBLIC_KEY: %v", err)
	}
	return &JWTService{publicKey: publicKey}
}

// parseTicketPublicKey decodifica base64 → PEM → PKIX e exige uma chave Ed25519.
func parseTicketPublicKey(encoded string) (ed25519.PublicKey, error) {
	if strings.TrimSpace(encoded) == "" {
		return nil, errors.New("TICKET_PUBLIC_KEY não definida")
	}

	der, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("TICKET_PUBLIC_KEY inválida: base64 inválido")
	}

	block, _ := pem.Decode(der)
	if block == nil {
		return nil, errors.New("TICKET_PUBLIC_KEY inválida: PEM inválido")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("TICKET_PUBLIC_KEY inválida: PKIX inválido")
	}

	edKey, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("TICKET_PUBLIC_KEY não é Ed25519")
	}

	return edKey, nil
}

// ParseToken verifica o ticket: só EdDSA, com aud hoodwink-ticket e exp válido.
func (j *JWTService) ParseToken(tokenStr string) (jwt.MapClaims, error) {
	if tokenStr == "" {
		return nil, errors.New("token vazio")
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		return j.publicKey, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
		jwt.WithAudience(TicketAudience),
	)
	if err != nil || !token.Valid {
		utils.LogError("ticket inválido")
		return nil, errors.New("token inválido")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		utils.LogError("claims inválidas")
		return nil, errors.New("claims inválidas")
	}

	return claims, nil
}

// ParseTokenFromRequest extrai o ticket da query string (?Ticket=).
func (j *JWTService) ParseTokenFromRequest(r *http.Request) (jwt.MapClaims, error) {
	queryToken := r.URL.Query().Get("Ticket")
	if queryToken != "" {
		return j.ParseToken(queryToken)
	}

	return nil, errors.New("nenhum token encontrado no header ou query")
}
