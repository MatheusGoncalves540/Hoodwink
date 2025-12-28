package rules

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
)

func GetCardRules(registryRules *Registry, roomData roomStructs.Room, card string) (*CardRules, error) {
	gameRules, err := registryRules.Get(string(roomData.Rules))
	if err != nil {
		return nil, err
	}
	cardRules := CardRules(gameRules.Cards[card])
	return &cardRules, nil
}
