package redisFuncs

import (
	"context"
	"sync"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

// roomMutexes serializa, dentro desta instância, tudo que lê-modifica-grava uma sala.
// O lock do Redis é por instância (reentrante), então sozinho não separava a jogada de um
// jogador do tick do gameProcessor: os dois carregavam a sala, e o último a salvar apagava
// o outro (ex.: uma contestação feita no fim da janela sumia).
// ponytail: um mutex por sala, nunca removido (bytes por sala criada); limpar ao apagar a sala se pesar.
var roomMutexes sync.Map

// WithRoom carrega a versão mais nova da sala já com o mutex dela e roda fn.
func WithRoom(ctx context.Context, rdb *redis.Client, roomID string, fn func(*rooms.Room)) error {
	m, _ := roomMutexes.LoadOrStore(roomID, &sync.Mutex{})
	mu := m.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	room, err := LoadRoom(ctx, rdb, roomID)
	if err != nil {
		return err
	}
	fn(room)
	return nil
}
