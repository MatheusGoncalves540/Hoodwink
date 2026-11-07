package services

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/structures"
	"github.com/golang-jwt/jwt/v5"
)

// TicketClaims é o payload do ticket de sala. As tags json mantêm os mesmos
// nomes que o gameServer já espera no WebSocket (playerId, username, roomId, exp).
// ip/ua entram só como registro (rastreio); iat sai do IssuedAt (RegisteredClaims).
type TicketClaims struct {
	PlayerID string `json:"playerId"`
	Username string `json:"username"`
	RoomID   string `json:"roomId"`
	IP       string `json:"ip,omitempty"`
	UA       string `json:"ua,omitempty"`
	jwt.RegisteredClaims
}

// TicketService assina o ticket de sala com Ed25519.
type TicketService struct {
	privateKey ed25519.PrivateKey
	expiration time.Duration
}

// NewTicketService carrega TICKET_PRIVATE_KEY (base64 de PEM PKCS8 em uma linha)
// e TICKET_EXPIRATION (horas, padrão 2). Falha se a chave faltar ou for inválida,
// sem imprimir o valor.
func NewTicketService() (*TicketService, error) {
	expInt, err := strconv.Atoi(os.Getenv("TICKET_EXPIRATION"))
	if err != nil || expInt <= 0 {
		expInt = 2
	}
	return newTicketService(os.Getenv("TICKET_PRIVATE_KEY"), time.Duration(expInt)*time.Hour)
}

func newTicketService(encoded string, expiration time.Duration) (*TicketService, error) {
	if strings.TrimSpace(encoded) == "" {
		return nil, errors.New("TICKET_PRIVATE_KEY não definida")
	}

	der, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("TICKET_PRIVATE_KEY inválida: base64 inválido")
	}

	block, _ := pem.Decode(der)
	if block == nil {
		return nil, errors.New("TICKET_PRIVATE_KEY inválida: PEM inválido")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("TICKET_PRIVATE_KEY inválida: PKCS8 inválido")
	}

	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("TICKET_PRIVATE_KEY não é Ed25519")
	}

	return &TicketService{
		privateKey: edKey,
		expiration: expiration,
	}, nil
}

// GenerateTicket assina um ticket EdDSA com aud hoodwink-ticket.
func (t *TicketService) GenerateTicket(player structures.UserClaims, roomId string) (string, error) {
	now := time.Now().UTC()
	claims := TicketClaims{
		PlayerID: player.Id,
		Username: player.Username,
		RoomID:   roomId,
		IP:       player.IP,
		UA:       player.UA,
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{TicketAudience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.expiration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(t.privateKey)
}
