// Code generated pelo enginebump (make engine-bump). NÃO edite à mão.
//
// Package engines registra as cópias congeladas do motor por versão.
// O runner (T10) consulta Registry para acionar o motor da versão da partida.
package engines

import (
	"context"

	enginev1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/engine"
	roomsv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/rooms"
	structsv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs"
	rulesv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/rules"
)

// EngineV1 liga as funções do motor v1 aos tipos congelados dessa versão.
type EngineV1 struct {
	ApplyPlay func(context.Context, roomsv1.Sink, *roomsv1.Room, *structsv1.PlayerPlay, *rulesv1.Registry) bool
	Tick      func(context.Context, roomsv1.Sink, *rulesv1.Registry, *roomsv1.Room)
}

// Version identifica a versão deste motor.
func (EngineV1) Version() int { return 1 }

// V1 é o motor congelado da versão 1.
var V1 = EngineV1{ApplyPlay: enginev1.ApplyPlay, Tick: enginev1.Tick}

// Engine é a interface mínima comum a todas as versões congeladas.
type Engine interface {
	Version() int
}

// Registry associa o número da versão ao motor congelado.
var Registry = map[int]Engine{
	1: V1,
}
