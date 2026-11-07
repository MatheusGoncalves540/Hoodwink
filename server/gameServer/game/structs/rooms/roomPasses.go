package rooms

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

// PassesState guarda quem já disse "deixar passar" na janela de reação atual.
// WindowKey identifica a janela (tipo|autor|expiração): quando a janela muda, os
// passes antigos deixam de valer sozinhos, sem precisar limpar o campo em cada
// protocolo que abre uma janela nova.
type PassesState struct {
	WindowKey string   `json:"windowKey,omitempty"`
	Players   []string `json:"players,omitempty"`
}

// reactionWindows são as janelas em que outros jogadores podem reagir (duvidar ou
// bloquear) e que, por isso, podem fechar antes do relógio quando todos já
// responderam. Janelas de anúncio puro (Pegar moeda, Trocar cartas, Contestação)
// ficam de fora: elas existem para a mesa ler o resultado.
var reactionWindows = map[structs.GameEventType]bool{
	structs.EventCardKilled:  true,
	structs.EventEarnCoins:   true,
	structs.EventStealCoins:  true,
	structs.EventInvestment:  true,
	structs.EventChangedCard: true,
	structs.EventRevivedCard: true,
}

// IsReactionWindow diz se a janela atual aceita reação de quem não jogou.
func (r *Room) IsReactionWindow() bool {
	if r.GameEvent == nil {
		return false
	}
	if strings.HasPrefix(string(r.GameEvent.Type), "CARD_PLAYED_") {
		return true
	}
	return reactionWindows[r.GameEvent.Type]
}

// reactionWindowKey identifica a janela de reação corrente.
func (r *Room) reactionWindowKey() string {
	if r.GameEvent == nil {
		return ""
	}
	return string(r.GameEvent.Type) + "|" + r.GameEvent.PlayerID + "|" +
		strconv.FormatInt(r.GameEvent.ExpiresAt.UnixMilli(), 10)
}

// ReactionEligiblePlayers devolve, em ordem determinística, quem ainda pode reagir
// na janela atual. É de propósito um superconjunto: quem não tem moeda para o
// Egoísta continua na lista e apenas segura o relógio (como hoje), mas ninguém que
// poderia agir fica de fora e perde a chance de bloquear.
func (r *Room) ReactionEligiblePlayers() []string {
	if !r.IsReactionWindow() {
		return nil
	}

	// Carta morta: só o dono da carta pode responder (Guardiã ou Kamikaze) —
	// guardianValidation.go e selfishValidation.go recusam os outros.
	if r.GameEvent.Type == structs.EventCardKilled {
		if target := payloadStringField(r.GameEvent.Payload, "targetPlayer"); target != "" {
			if player := r.Players[target]; player != nil && player.Alive {
				return []string{target}
			}
			return nil
		}
		// payload ilegível: cai no conjunto conservador abaixo
	}

	eligible := make([]string, 0, len(r.Players))
	for _, id := range r.SortedPlayerIDs() {
		player := r.Players[id]
		if player == nil || !player.Alive || id == r.GameEvent.PlayerID {
			continue
		}
		eligible = append(eligible, id)
	}
	return eligible
}

// payloadStringField lê um campo de texto do payload do evento, que chega como
// struct (na jogada) ou como map (depois da ida e volta do Redis).
func payloadStringField(payload any, field string) string {
	if payload == nil {
		return ""
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		return ""
	}
	value, _ := fields[field].(string)
	return value
}

// RegisterPass anota o "deixar passar" do jogador na janela atual e informa se
// todos os elegíveis já responderam.
func (r *Room) RegisterPass(playerID string) bool {
	key := r.reactionWindowKey()
	if r.Passes.WindowKey != key {
		r.Passes = PassesState{WindowKey: key, Players: []string{}}
	}
	for _, id := range r.Passes.Players {
		if id == playerID {
			return r.reactionIsComplete()
		}
	}
	r.Passes.Players = append(r.Passes.Players, playerID)
	return r.reactionIsComplete()
}

// reactionIsComplete diz se a janela pode fechar: existe pelo menos um elegível e
// todos os elegíveis já passaram.
func (r *Room) reactionIsComplete() bool {
	eligible := r.ReactionEligiblePlayers()
	if len(eligible) == 0 {
		return false
	}
	passed := make(map[string]bool, len(r.Passes.Players))
	for _, id := range r.Passes.Players {
		passed[id] = true
	}
	for _, id := range eligible {
		if !passed[id] {
			return false
		}
	}
	return true
}
