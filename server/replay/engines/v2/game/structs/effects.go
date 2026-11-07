package structs

import (
	"encoding/json"
	"fmt"
)

type EffectCause string

type EffectDTO struct {
	Cause        EffectCause     `json:"cause"`
	SourcePlayer string          `json:"sourcePlayer"`
	Payload      json.RawMessage `json:"payload"`
}

// ===== EFEITO (precisa ser executado) =====
type Effect struct {
	Cause        EffectCause
	SourcePlayer string
	Payload      any
}

func NewEffect(cause EffectCause, sourcePlayer string, payload any) Effect {
	return Effect{Cause: cause, SourcePlayer: sourcePlayer, Payload: payload}
}

func (e Effect) ToDTO() (EffectDTO, error) {
	var payload json.RawMessage
	if e.Payload != nil {
		data, err := json.Marshal(e.Payload)
		if err != nil {
			return EffectDTO{}, fmt.Errorf("falha ao serializar payload do efeito %s: %w", e.Cause, err)
		}
		payload = data
	}

	return EffectDTO{
		Cause:        e.Cause,
		SourcePlayer: e.SourcePlayer,
		Payload:      payload,
	}, nil
}

// ProtectionBlocks diz se uma proteção da Guardiã já colocada segura a morte por essa causa.
// Contestação e perda por ganância nunca são seguradas. O Kamikaze é: rules.html, "caso a carta
// já estivesse protegida antes, o ataque pode ser parado" (só não dá para reagir com a Guardiã na hora).
func ProtectionBlocks(cause EffectCause) bool {
	return cause != EffectContestPenalty && cause != EffectGreed
}

// Tipos de efeitos
const (
	EffectContest        EffectCause = "CONTEST"
	EffectContestPenalty EffectCause = "CONTEST_PENALTY"
	EffectGreed          EffectCause = "GREED"
	EffectBasicKill      EffectCause = "KILL_CARD"
	EffectTakeCoins      EffectCause = "TAKE_COINS"
	EffectSwapCards      EffectCause = "SWAP_CARDS"
	EffectEarnCoins      EffectCause = "EARN_COINS"
	EffectStealCoins     EffectCause = "STEAL_COINS"
	EffectCardKilled     EffectCause = "CARD_KILLED"
	EffectRevivedCard    EffectCause = "CARD_REVIVED"
	EffectChangedCard    EffectCause = "CARD_CHANGED"
	EffectInvestment     EffectCause = "INVESTMENT"
	EffectRichnessDenied EffectCause = "RICHNESS_DENIED"

	EffectAssassin     EffectCause = "ASSASSIN"
	EffectKamikaze     EffectCause = "KAMIKAZE"
	EffectTrillionaire EffectCause = "TRILLIONAIRE"
	EffectPolitical    EffectCause = "POLITICAL"
	EffectRebel        EffectCause = "REBEL"
	EffectClairvoyant  EffectCause = "CLAIRVOYANT"
	EffectGuardian     EffectCause = "GUARDIAN"
	EffectTrickster    EffectCause = "TRICKSTER"
	EffectGravedigger  EffectCause = "GRAVEDIGGER"
	EffectCroupier     EffectCause = "CROUPIER"
	EffectInvestor     EffectCause = "INVESTOR"
	EffectSelfish      EffectCause = "SELFISH"
)
