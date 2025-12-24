package events

import (
	"context"
	"encoding/json"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

// PushPlayerPlay adiciona um novo evento à fila de eventos da sala no Redis.
// Parâmetros:
//
//	ctx: contexto para timeout/cancelamento
//	rdb: cliente Redis
//	RoomId: identificador da sala
//	playerPlay: evento a ser adicionado (struct Event)
//
// Retorno:
//
//	error: erro de serialização ou comunicação com Redis
func PushPlayerPlay(ctx context.Context, rdb *redis.Client, RoomId string, playerPlay roomStructs.PendingEvent) error {
	// Serializa o evento para JSON
	data, err := json.Marshal(playerPlay)
	if err != nil {
		// Retorna erro se não conseguir serializar
		return err
	}
	// Adiciona o evento no início da fila de eventos (LPUSH)
	return rdb.LPush(ctx, "room:"+RoomId+":playQueue", data).Err()
}

// Remove e retorna o próximo evento da fila de eventos do Redis para a sala especificada.
// Usa BRPop para esperar até que um evento esteja disponível.
// Retorna o evento desempilhado ou nil se não houver evento.
// Em caso de erro de deserialização ou Redis, retorna o erro correspondente.
func PopEvent(ctx context.Context, rdb *redis.Client, RoomId string) (*roomStructs.PendingEvent, error) {
	// Busca o próximo evento na fila (bloqueante até existir)
	res, err := rdb.BRPop(ctx, 0, "room:"+RoomId+":playQueue").Result()
	if err != nil {
		return nil, err // erro ao acessar o Redis
	}
	if len(res) < 2 {
		return nil, nil // nenhum evento disponível
	}
	var playerPlay roomStructs.PendingEvent
	// Converte o JSON armazenado em struct Event
	err = json.Unmarshal([]byte(res[1]), &playerPlay)
	if err != nil {
		return nil, err // erro ao desserializar o evento
	}
	return &playerPlay, nil // retorna o evento desempilhado
}

// ScheduleNextStep agenda a execução de um evento para o futuro, usando time.AfterFunc.
// Parâmetros:
//
//	ctx: contexto para timeout/cancelamento
//	rdb: cliente Redis
//	RoomId: identificador da sala
//	playerPlay: evento a ser agendado (struct Event)
//
// Não retorna erro, apenas agenda a execução.
func ScheduleNextStep(ctx context.Context, rdb *redis.Client, RoomId string, playerPlay roomStructs.PendingEvent) {
}
