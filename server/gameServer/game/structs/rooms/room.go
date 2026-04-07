package rooms

import (
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
)

type Room struct {
	ID                        string                                                `json:"id"`
	Rules                     structs.Rules                                         `json:"rules"`
	TimeoutType               string                                                `json:"timeoutType"`
	Name                      string                                                `json:"name"`
	Password                  string                                                `json:"password" validate:"max=24"`
	MaxPlayers                int                                                   `json:"maxPlayers"`
	CustomMatch               bool                                                  `json:"customMatch"`
	Turn                      int                                                   `json:"turn"`
	Tax                       int                                                   `json:"tax"`
	DoubledCardValues         map[structs.TypePlayerPlays]structs.DoubledCardValues `json:"doubledCardValues"`
	Players                   map[string]*players.Player                            `json:"players"`
	Deck                      []structs.CardName                                    `json:"deck"`
	CurrentPlayer             string                                                `json:"currentPlayer"`
	GameEvent                 *structs.GameEvent                                    `json:"gameEvent"`
	PendingEffects            []structs.EffectDTO                                   `json:"pendingEffects"`
	PendingPresentationEvents []structs.PresentationEventDTO                        `json:"pendingPresentationEvents"`
	GameOver                  bool                                                  `json:"gameOver"`
	StartTime                 time.Time                                             `json:"startTime"`
	Created                   time.Time                                             `json:"created"`
	Chat                      []structs.ChatMessage                                 `json:"chat"`
}

type PublicRoomForUpdates struct {
	ID                string                                                `json:"id"`
	Name              string                                                `json:"name"`
	Turn              int                                                   `json:"turn"`
	Tax               int                                                   `json:"tax"`
	DoubledCardValues map[structs.TypePlayerPlays]structs.DoubledCardValues `json:"doubledCardValues"`
	Players           map[string]players.PublicPlayerForUpdates             `json:"players"`
	Deck              int                                                   `json:"deck"`
	CurrentPlayer     string                                                `json:"currentPlayer"`
	GameEvent         *structs.GameEvent                                    `json:"gameEvent"`
	PendingEffects    []structs.EffectDTO                                   `json:"pendingEffects"`
	GameOver          bool                                                  `json:"gameOver"`
	StartTime         time.Time                                             `json:"startTime"`
	Chat              []structs.ChatMessage                                 `json:"chat"`
}
