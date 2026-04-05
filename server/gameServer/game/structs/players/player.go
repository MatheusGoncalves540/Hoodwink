package players

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

type Player struct {
	Id                    string         `json:"id"`
	Name                  string         `json:"name"`
	MoveTime              int            `json:"moveTime"`
	Cards                 []structs.Card `json:"cards"`
	Coins                 int            `json:"coins"`
	Investments           []int          `json:"investments"`
	PendingInvestmentCoin bool           `json:"pendingInvestmentCoin"`
	Alive                 bool           `json:"alive"`
}

type PublicPlayerForUpdates struct {
	Id                    string                         `json:"id"`
	Name                  string                         `json:"name"`
	MoveTime              int                            `json:"moveTime"`
	Cards                 []structs.PublicCardForUpdates `json:"cards"`
	Coins                 int                            `json:"coins"`
	Investments           []int                          `json:"investments"`
	PendingInvestmentCoin bool                           `json:"pendingInvestmentCoin"`
	Alive                 bool                           `json:"alive"`
}
