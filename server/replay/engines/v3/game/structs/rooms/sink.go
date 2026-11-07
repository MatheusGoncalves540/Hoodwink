package rooms

import (
	"context"
	"time"
)

// Sink é o ponto de saída do motor puro para o mundo externo: transmitir e
// agendar/limpar timeouts. A persistência (Save) fica com o WithRoom, que salva
// quando o fn devolve aplicado. No replay/testes o sink é no-op (rdb == nil).
type Sink interface {
	// Broadcast transmite a versão atualizada da sala (ou uma versão confidencial
	// para um subconjunto de jogadores).
	Broadcast(ctx context.Context, room *Room, confidencial *Room, confidencialPlayerIds []string) error
	// RegistryTimeout agenda a sala na fila de timeouts do processador para expiresAt.
	RegistryTimeout(ctx context.Context, room *Room, expiresAt time.Time)
	// ClearTimeout remove a sala da fila de timeouts (ex.: fim de jogo).
	ClearTimeout(ctx context.Context, room *Room)
}

// RoomLock é o lock distribuído ativo durante uma chamada do WithRoom. A
// implementação concreta (Redis) fica em redisFuncs; o motor só conhece esta
// interface e a usa para não gravar por cima de outro pod caso o lock tenha sido
// perdido durante a chamada.
type RoomLock interface {
	// Err retorna o erro do contexto derivado (context.Canceled) quando o lock foi perdido.
	Err() error
	// Context retorna o contexto derivado usado para a renovação em segundo plano.
	Context() context.Context
}

// noopSink é o Sink que não faz nada, usado quando rdb é nil (replay/testes).
type noopSink struct{}

func (noopSink) Broadcast(context.Context, *Room, *Room, []string) error { return nil }

func (noopSink) RegistryTimeout(context.Context, *Room, time.Time) {}

func (noopSink) ClearTimeout(context.Context, *Room) {}

// NoopSink retorna um Sink que não faz nada (replay/testes sem Redis).
func NoopSink() Sink { return noopSink{} }

// SinkOrNoop devolve o sink recebido ou um no-op quando ele for nil.
func SinkOrNoop(sink Sink) Sink {
	if sink == nil {
		return noopSink{}
	}
	return sink
}
