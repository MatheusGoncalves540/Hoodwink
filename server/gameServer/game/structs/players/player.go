package players

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

type Player struct {
	Id    string         `json:"id"`
	Name  string         `json:"name"`
	Cards []structs.Card `json:"cards"`
	Coins int            `json:"coins"`
	Alive bool           `json:"alive"`
}

type PublicPlayerForUpdates struct {
	Id    string                         `json:"id"`
	Name  string                         `json:"name"`
	Cards []structs.PublicCardForUpdates `json:"cards"`
	Coins int                            `json:"coins"`
	Alive bool                           `json:"alive"`
}
