package connection

import (
	"context"
	"encoding/json"
	"log"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/runtime/chat/commands"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// ProcessDebugCommand executa comandos de debug recebidos pelo WebSocket e devolve
// aplicado. Só roda com UNSAFE_DEBUG=true (padrão desligado); sem isso nada é
// executado, o estado da sala não muda e nada é transmitido. Comando de debug marca
// a partida como não replayável.
func ProcessDebugCommand(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerPlay *structs.PlayerPlay) bool {
	if !utils.IsUnsafeDebugMode() {
		return false
	}

	var debugCommands = map[structs.TypePlayerPlays]func(){
		"COMMAND_SEND_BROADCAST": func() {
			commands.CommandSendBroadcast(ctx, rdb, roomData, playerPlay.Payload)
		},
		"COMMAND_PRINT_ROOM": func() {
			printRoomSanitized(roomData)
		},
	}

	if cmd, ok := debugCommands[playerPlay.Type]; ok {
		log.Printf("debug command %s player=%s", playerPlay.Type, playerPlay.PlayerId)
		cmd()
		roomData.NonReplayable = true
		if err := redisFuncs.SendUpdatedRoomData(ctx, rdb, roomData, nil, []string{}); err != nil {
			return false
		}
		return true
	}
	return false
}

// ProcessStartCommand é o start legítimo da partida (botão "abrir as cortinas"):
// qualquer jogador presente na sala pode iniciar enquanto ela não começou. Não é
// comando de debug e a partida continua replayável.
func ProcessStartCommand(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerPlay *structs.PlayerPlay) bool {
	if playerPlay.Type != "COMMAND_START_GAME" {
		return false
	}
	if !roomData.StartTime.IsZero() || roomData.GameOver {
		return false
	}
	if _, err := roomData.GetPlayer(playerPlay.PlayerId); err != nil {
		return false
	}
	commands.CommandStartGame(ctx, rdb, roomData)
	return true
}

// printRoomSanitized imprime a sala sem a senha nem o baralho/mãos completos,
// usando a mesma projeção pública enviada aos clientes.
func printRoomSanitized(roomData *rooms.Room) {
	data, err := json.MarshalIndent(roomData.GetPublicRoomForUpdates(), "", "  ")
	if err != nil {
		log.Printf("debug command COMMAND_PRINT_ROOM erro: %v", err)
		return
	}
	log.Println(string(data))
}
