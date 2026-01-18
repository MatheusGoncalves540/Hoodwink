package engine

import (
	"encoding/json"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

func BuildEffect(dto structs.EffectDTO) (structs.Effect, error) {
	switch dto.Cause {

	case structs.EffectContest:
		var payload structs.ContestPayload
		if err := json.Unmarshal(dto.Payload, &payload); err != nil {
			return structs.Effect{}, err
		}
		return structs.Effect{
			Cause:        dto.Cause,
			SourcePlayer: dto.SourcePlayer,
			Payload:      &payload,
		}, nil

	default:
		return structs.Effect{}, fmt.Errorf("efeito desconhecido")
	}
}
