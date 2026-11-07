package services

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink-replay/format"
)

// replaySigningKid identifica a chave de assinatura dos replays no cabeçalho .hwr.
// O client/web embute a chave pública correspondente por este kid (plano 05, T13).
const replaySigningKid = "replay-v1"

// ReplayService carrega a chave de assinatura de replay e fornece o signer Ed25519.
type ReplayService struct {
	signer *format.Ed25519Signer
}

// NewReplayService carrega REPLAY_SIGNING_KEY (base64 de PEM PKCS8 Ed25519 em uma
// linha). Falha se a chave faltar ou for inválida, sem imprimir o valor.
func NewReplayService() (*ReplayService, error) {
	return newReplayService(os.Getenv("REPLAY_SIGNING_KEY"))
}

func newReplayService(encoded string) (*ReplayService, error) {
	signer, err := newReplaySigner(encoded)
	if err != nil {
		return nil, err
	}
	return &ReplayService{signer: signer}, nil
}

// Signer devolve o signer Ed25519 pronto para format.Encode. Nunca nil.
func (s *ReplayService) Signer() *format.Ed25519Signer { return s.signer }

func newReplaySigner(encoded string) (*format.Ed25519Signer, error) {
	if strings.TrimSpace(encoded) == "" {
		return nil, errors.New("REPLAY_SIGNING_KEY não definida")
	}

	der, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("REPLAY_SIGNING_KEY inválida: base64 inválido")
	}

	block, _ := pem.Decode(der)
	if block == nil {
		return nil, errors.New("REPLAY_SIGNING_KEY inválida: PEM inválido")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("REPLAY_SIGNING_KEY inválida: PKCS8 inválido")
	}

	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("REPLAY_SIGNING_KEY não é Ed25519")
	}

	return format.NewEd25519Signer(replaySigningKid, edKey), nil
}
